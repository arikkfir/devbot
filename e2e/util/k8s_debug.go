package util

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/asaskevich/govalidator"
	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	yamlutil "sigs.k8s.io/yaml"

	apiv1 "github.com/arikkfir/devbot/api/v1"
)

func PrintK8sDebugInfo(ctx SpecContext, c client.Client, cfg *rest.Config, nsName string) {
	GinkgoHelper()
	if !ctx.SpecReport().Failed() || ctx.SpecReport().State == ginkgotypes.SpecStateInterrupted {
		return
	}

	printList(ctx, c, cfg, &corev1.NamespaceList{}, client.MatchingFields{"metadata.name": nsName})
	printList(ctx, c, cfg, &apiv1.RepositoryList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &apiv1.ApplicationList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &apiv1.EnvironmentList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &apiv1.DeploymentList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &corev1.ServiceAccountList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &rbacv1.ClusterRoleList{}, client.MatchingLabels{"devbot.kfirs.com/purpose": "test", "devbot.kfirs.com/target": nsName})
	printList(ctx, c, cfg, &rbacv1.ClusterRoleBindingList{}, client.MatchingLabels{"devbot.kfirs.com/purpose": "test", "devbot.kfirs.com/target": nsName})
	printList(ctx, c, cfg, &rbacv1.RoleList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &rbacv1.RoleBindingList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &corev1.ConfigMapList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &corev1.SecretList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &batchv1.JobList{}, client.InNamespace(nsName))
	printList(ctx, c, cfg, &corev1.PodList{}, client.InNamespace(nsName))
}

func printList(ctx SpecContext, c client.Client, cfg *rest.Config, list client.ObjectList, opts ...client.ListOption) {
	GinkgoHelper()
	clientset, err := kubernetes.NewForConfig(cfg)
	Expect(err).To(BeNil())

	Expect(c.List(ctx, list, opts...)).To(Succeed())

	itemsField := reflect.ValueOf(list).Elem().FieldByName("Items")
	itemsCount := itemsField.Len()

	rawItemTypeName := reflect.TypeOf(list).Elem().Name()
	itemTypeNameWithoutListSuffix := strings.TrimSuffix(rawItemTypeName, "List")
	snakeCaseItemTypeName := govalidator.CamelCaseToUnderscore(itemTypeNameWithoutListSuffix)
	spaceDelimitedItemTypeName := strings.ReplaceAll(snakeCaseItemTypeName, "_", " ")
	titledItemTypeName := strings.ToTitle(string(spaceDelimitedItemTypeName[0])) + spaceDelimitedItemTypeName[1:]

	if itemsCount > 0 {
		for i := 0; i < itemsCount; i++ {
			itemValue := itemsField.Index(i)

			// Reset managed fields to avoid cluttering the output
			itemValue.FieldByName("ObjectMeta").FieldByName("ManagedFields").Set(reflect.ValueOf([]metav1.ManagedFieldsEntry{}))

			// Print title
			itemName := itemValue.FieldByName("ObjectMeta").FieldByName("Name").String()
			if itemName == "default" || itemName == "kube-root-ca.crt" {
				continue
			}
			printTitle(GinkgoWriter, '=', titledItemTypeName+" "+itemName)

			// Print object
			o := itemValue.Interface()
			printObjectAsYAML(GinkgoWriter, o)

			// If it's a pod, print its logs
			if pod, ok := o.(corev1.Pod); ok {
				GinkgoWriter.Printf("-[ Logs ]%s\n", strings.Repeat("-", 80-len("-[ Logs ]")))
				req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
				podLogs, err := req.Stream(ctx)
				Expect(err).To(BeNil())
				Expect(io.Copy(GinkgoWriter, podLogs)).Error().To(Succeed())
				podLogs.Close()
			}
		}
	}
}

func printTitle(w io.Writer, lineChar rune, title string) {
	GinkgoHelper()
	Expect(fmt.Fprintf(w, "%s[ %s ]%s\n", string(lineChar), title, strings.Repeat(string(lineChar), 80-5-len(title)))).Error().To(Succeed())
}

func printObjectAsYAML(w io.Writer, o any) {
	GinkgoHelper()
	b, err := yamlutil.Marshal(o)
	Expect(err).To(BeNil())

	tokens, err := lexers.Get("yaml").Tokenise(nil, string(b))
	Expect(err).To(BeNil())
	Expect(formatters.Get("terminal256").Format(w, styles.Get("solarized-dark"), tokens)).To(Succeed())
}
