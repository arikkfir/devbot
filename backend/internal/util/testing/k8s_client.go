package testing

import (
	"bytes"
	"context"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/k8s"
	"github.com/arikkfir/devbot/backend/internal/util/strings"
	. "github.com/arikkfir/justest"
	"io"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"time"
)

const (
	DevbotNamespace                              = "devbot"
	DevbotRepositoryControllerServiceAccountName = "devbot-controller"
)

var (
	k8sServiceAccountKind = reflect.TypeOf(corev1.ServiceAccount{}).Name()
	k8sClusterRoleKind    = reflect.TypeOf(rbacv1.ClusterRole{}).Name()
)

type KClient struct {
	Client  client.Client
	Manager manager.Manager
	ctx     context.Context
}

func K(ctx context.Context, t T) *KClient {
	userHomeDir, err := os.UserHomeDir()
	With(t).Verify(err).Will(BeNil()).OrFail()

	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	With(t).Verify(err).Will(BeNil()).OrFail()

	scheme := runtime.NewScheme()
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
		Client: client.Options{
			Scheme: scheme,
			Cache: &client.CacheOptions{
				DisableFor: nil,
			},
		},
		LeaderElection:         false,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		PprofBindAddress:       "0",
	})
	With(t).Verify(err).Will(BeNil()).OrFail()
	With(t).Verify(k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Environment{})).Will(Succeed()).OrFail()
	With(t).Verify(k8s.AddOwnershipIndex(ctx, mgr.GetFieldIndexer(), &apiv1.Deployment{})).Will(Succeed()).OrFail()

	go func() {
		if err := mgr.Start(ctx); err != nil {
			t.Logf("Kubernetes manager failed: %+v", err)
		}
	}()

	time.Sleep(3 * time.Second)
	return &KClient{Client: mgr.GetClient(), Manager: mgr, ctx: ctx}
}

func (k *KClient) CreateNamespace(t T) *KNamespace {
	devbotGitOpsName := "devbot-gitops"

	r := &corev1.Namespace{ObjectMeta: ctrl.ObjectMeta{Name: strings.RandomHash(7)}}
	With(t).Verify(k.Client.Create(k.ctx, r)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(k.Client.Delete(k.ctx, r)).Will(Succeed()).OrFail() })

	sa := &corev1.ServiceAccount{ObjectMeta: ctrl.ObjectMeta{Name: devbotGitOpsName, Namespace: r.Name}}
	With(t).Verify(k.Client.Create(k.ctx, sa)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(k.Client.Delete(k.ctx, sa)).Will(Succeed()).OrFail() })

	role := &rbacv1.Role{
		ObjectMeta: ctrl.ObjectMeta{Name: devbotGitOpsName, Namespace: r.Name},
		Rules:      []rbacv1.PolicyRule{{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}}},
	}
	With(t).Verify(k.Client.Create(k.ctx, role)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(k.Client.Delete(k.ctx, role)).Will(Succeed()).OrFail() })

	rb := &rbacv1.RoleBinding{
		ObjectMeta: ctrl.ObjectMeta{Name: devbotGitOpsName, Namespace: r.Name},
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: devbotGitOpsName},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: devbotGitOpsName}},
	}
	With(t).Verify(k.Client.Create(k.ctx, rb)).Will(Succeed()).OrFail()
	t.Cleanup(func() { With(t).Verify(k.Client.Delete(k.ctx, rb)).Will(Succeed()).OrFail() })

	printPodLogs := func(clientset *kubernetes.Clientset, pod *corev1.Pod) {
		podLogOpts := corev1.PodLogOptions{}
		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
		podLogs, err := req.Stream(k.ctx)
		if err != nil {
			t.Logf("Failed to get logs for pod '%s' (in order to obtain logs for printing): %+v", pod.Name, err)
			return
		}
		defer podLogs.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, podLogs); err != nil {
			t.Logf("Failed to copy logs for pod '%s' (in order to obtain logs for printing): %+v", pod.Name, err)
			return
		}

		t.Logf("Logs for pod '%s' of job '%s':\n%s", pod.Name, pod.Labels["job-name"], buf.String())
	}

	printJobLogs := func(clientset *kubernetes.Clientset, job *v1.Job) {
		jobs := &v1.JobList{}
		if err := k.Client.List(k.ctx, jobs, client.InNamespace(r.Name)); err != nil {
			t.Logf("Failed to list jobs (in order to obtain logs for printing): %+v", err)
			return
		}
		for _, job := range jobs.Items {
			pods := &corev1.PodList{}
			if err := k.Client.List(k.ctx, pods, client.InNamespace(r.Name), client.MatchingLabels{"job-name": job.Name}); err != nil {
				t.Logf("Failed to list pods for job '%s' (in order to obtain logs for printing): %+v", job.Name, err)
			}
			for _, pod := range pods.Items {
				printPodLogs(clientset, &pod)
			}
		}
	}

	t.Cleanup(func() {
		if t.Failed() {
			kConfig := k.Manager.GetConfig()
			clientset, err := kubernetes.NewForConfig(kConfig)
			if err != nil {
				t.Logf("Failed to create Kubernetes clientset (in order to obtain logs for printing): %+v", err)
				return
			}

			jobs := &v1.JobList{}
			if err := k.Client.List(k.ctx, jobs, client.InNamespace(r.Name)); err != nil {
				t.Logf("Failed to list jobs (in order to obtain logs for printing): %+v", err)
				return
			}

			for _, job := range jobs.Items {
				printJobLogs(clientset, &job)
			}
		}
	})

	return &KNamespace{Name: r.Name, k: k}
}
