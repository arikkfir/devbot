package util

import (
	"context"
	"encoding/json"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	KubeConfig       *rest.Config
	K8sClientSet     *kubernetes.Clientset
	K8sDynamicClient *dynamic.DynamicClient
	AppRESTClient    *rest.RESTClient
}

func NewK8sClient(kubeConfig *rest.Config) *K8sClient {
	// Create Kubernetes client-set
	k8sClientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client-set")
	}

	// Create Kubernetes dynamic client
	k8sDynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes dynamic client")
	}

	// Create Application client
	httpClient, err := rest.HTTPClientFor(kubeConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create custom Kubernetes HTTP client (for custom REST client)")
	}
	appConfig := *kubeConfig
	appConfig.GroupVersion = &apiv1.GroupVersion
	appConfig.APIPath = "/apis"
	appConfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	if appConfig.UserAgent == "" {
		appConfig.UserAgent = rest.DefaultKubernetesUserAgent() // TODO: consider customizing API user-agent
	}
	appRESTClient, err := rest.RESTClientForConfigAndClient(&appConfig, httpClient)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create custom Kubernetes REST client")
	}

	return &K8sClient{
		KubeConfig:       kubeConfig,
		K8sClientSet:     k8sClientSet,
		K8sDynamicClient: k8sDynamicClient,
		AppRESTClient:    appRESTClient,
	}
}

func (c *K8sClient) GetApplications(ctx context.Context) ([]apiv1.Application, error) {
	raw, err := c.K8sClientSet.RESTClient().
		Get().
		AbsPath(fmt.Sprintf("/apis/%s/%s/applications", apiv1.GroupVersion.Group, apiv1.GroupVersion.Version)).
		DoRaw(ctx)
	if err != nil {
		return nil, errors.New("failed to get applications", err)
	}

	applications := apiv1.ApplicationList{}
	if err := json.Unmarshal(raw, &applications); err != nil {
		return nil, errors.New("failed to unmarshal applications", err)
	}

	return applications.Items, nil
}

func (c *K8sClient) UpdateApplicationStatus(ctx context.Context, app *apiv1.Application) error {
	_, err := c.K8sClientSet.RESTClient().
		Put().
		AbsPath(fmt.Sprintf("/apis/%s/%s", apiv1.GroupVersion.Group, apiv1.GroupVersion.Version)).
		Resource("applications").
		SubResource("status").
		Body(&app).
		DoRaw(ctx)
	if err != nil {
		return errors.New("failed to update application status", err)
	}
	return nil
}
