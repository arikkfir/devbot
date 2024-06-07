package bootstrap

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"time"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	gossh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v62/github"
	"github.com/rs/zerolog/log"
	"github.com/secureworks/errors"
	"golang.org/x/crypto/ssh"
	appsv1 "k8s.io/api/apps/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	k8syaml "sigs.k8s.io/yaml"

	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util/lang"
	"github.com/arikkfir/devbot/backend/internal/util/version"
)

var (
	//go:embed all:resources
	resourcesFS embed.FS

	// List of CRDs we will expect to find in our verification phase
	expectedCRDs = []string{
		"applications.devbot.kfirs.com",
		"deployments.devbot.kfirs.com",
		"environments.devbot.kfirs.com",
		"repositories.devbot.kfirs.com",
	}

	// List of deployments we will expect to find in our verification phase
	expectedDeployments = []string{
		"devbot-controller",
		"devbot-github-webhook",
	}

	// Annotations & Labels we will expect to find in our objects during verification phase
	expectedAnnotations = map[string]string{
		"app.kubernetes.io/build": semver.MustParse(version.Version).Build[0],
	}
	expectedLabels = map[string]string{
		"app.kubernetes.io/name":       "devbot",
		"app.kubernetes.io/version":    semver.MustParse(version.Version).FinalizeVersion(),
		"app.kubernetes.io/managed-by": "devbot",
	}
)

type EnvVarJSONPatchOperation struct {
	Op    string    `json:"op"`
	Path  string    `json:"path"`
	Value v1.EnvVar `json:"value"`
}

type GitHubBootstrapper struct {
	webhooksURL        string
	timeout            time.Duration
	ghc                *github.Client
	k8sRestConfig      *rest.Config
	k8sDynamicClient   *dynamic.DynamicClient
	k8sDiscoveryClient *discovery.DiscoveryClient
	k8sRESTMapper      *restmapper.DeferredDiscoveryRESTMapper
}

func NewGitHubBootstrapper(ctx context.Context, pat string, timeout time.Duration, webhooksURL string, restConfig *rest.Config) (*GitHubBootstrapper, error) {

	// Create the GitHub client
	ghc := github.NewClient(nil).WithAuthToken(pat)
	if _, _, err := ghc.Users.Get(ctx, ""); err != nil {
		return nil, errors.New("failed obtaining authenticated session info: %w", err)
	}

	// Create dynamic-structure Kubernetes client for unstructured objects
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.New("failed building dynamic Kubernetes client: %w", err)
	}

	// Create the Kubernetes Discovery client
	dc := discovery.NewDiscoveryClientForConfigOrDie(restConfig)
	mc := memory.NewMemCacheClient(dc)
	rm := restmapper.NewDeferredDiscoveryRESTMapper(mc)

	return &GitHubBootstrapper{
		webhooksURL:        webhooksURL,
		timeout:            timeout,
		ghc:                ghc,
		k8sRestConfig:      restConfig,
		k8sDynamicClient:   dynamicClient,
		k8sDiscoveryClient: dc,
		k8sRESTMapper:      rm,
	}, nil
}

func (b *GitHubBootstrapper) Bootstrap(ctx context.Context, owner, name, visibility string) error {

	// Prepare clone directory
	workspacePath, err := os.MkdirTemp("", fmt.Sprintf("workspace-%s-%s-*", owner, name))
	if err != nil {
		return errors.New("could not create temporary workspace directory: %w", err)
	}
	defer os.RemoveAll(workspacePath)
	log.Info().Str("path", workspacePath).Msg("Created temporary workspace directory")

	// Ensure repository exists
	log.Info().Str("owner", owner).Str("name", name).Msg("Ensuring GitHub repository exists (will create if missing)")
	if _, err := b.ensureRepository(ctx, owner, name, visibility); err != nil {
		return errors.New("failed to ensure repository '%s/%s' exists: %w", owner, name, err)
	}

	// Add deploy key
	log.Info().Msg("Adding deploy key to repository")
	private, _, err := b.addDeployKey(ctx, owner, name)
	if err != nil {
		return errors.New("failed adding deploy key for '%s/%s': %w", owner, name, err)
	}

	// Create Git public key auth
	log.Info().Msg("Preparing public key for authentication")
	publicKeys, err := gossh.NewPublicKeys("git", []byte(private), "")
	if err != nil {
		return errors.New("failed to prepare public key auth: %w", err)
	}

	// Clone repository
	clonePath := filepath.Join(workspacePath, "clone")
	log.Info().Str("path", clonePath).Msg("Cloning")
	repo, err := git.PlainCloneContext(ctx, clonePath, false, &git.CloneOptions{
		Auth:          publicKeys,
		Progress:      io.Discard,
		ReferenceName: "refs/heads/main",
		RemoteName:    "origin",
		SingleBranch:  true,
		Tags:          git.NoTags,
		URL:           fmt.Sprintf("git@github.com:%s/%s.git", owner, name),
	})
	if err != nil {
		return errors.New("failed cloning repository '%s/%s': %w", owner, name, err)
	}

	// Populate repository with Devbot manifests
	log.Info().Msg("Populating Devbot manifests into repository")
	if err := b.populateRepository(ctx, repo, publicKeys); err != nil {
		return errors.New("failed populating repository '%s/%s': %w", owner, name, err)
	}

	// Deploy devbot into cluster
	log.Info().Msg("Deploying devbot into cluster")
	if err := b.deployToCluster(ctx, filepath.Join(clonePath, ".devbot")); err != nil {
		return errors.New("failed deploying devbot into cluster: %w", err)
	}

	// Verify devbot is up & running
	log.Info().Msg("Verifying Devbot is up & running, ready for work...")
	if err := b.verifyDevbotIsReady(ctx); err != nil {
		return errors.New("failed verifying devbot deployment in cluster: %w", err)
	}

	return nil
}

func (b *GitHubBootstrapper) ensureRepository(ctx context.Context, owner, name, visibility string) (*github.Repository, error) {
	if repo, resp, err := b.ghc.Repositories.Get(ctx, owner, name); err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// Repository does not exist - create it first
			repo = &github.Repository{
				Name:        lang.Ptr(name),
				Description: lang.Ptr("GitOps repository"),
				Visibility:  lang.Ptr(visibility),
				IsTemplate:  lang.Ptr(false),
				AutoInit:    lang.Ptr(true),
			}
			if _, _, err := b.ghc.Repositories.Create(ctx, owner, repo); err != nil {
				log.Err(err).Str("response", string(lang.MustReadAll(resp.Body))).Msg("GitHub repository creation request failed")
				return nil, errors.New("failed to create repository: %w", err)
			} else {
				return repo, nil
			}
		} else {
			log.Err(err).Str("response", string(lang.MustReadAll(resp.Body))).Msg("GitHub repository get request failed")
			return nil, errors.New("failed to get repository: %w", err)
		}
	} else {
		return repo, nil
	}
}

func (b *GitHubBootstrapper) addDeployKey(ctx context.Context, owner, name string) (string, string, error) {
	private, public, err := b.generateSSHKeyPair()
	if err != nil {
		return private, public, err
	}

	// Create a deploy key and save it
	deployKey := &github.Key{
		Key:      github.String(public),
		Title:    lang.Ptr("Devbot"),
		ReadOnly: lang.Ptr(false),
	}
	if _, _, err := b.ghc.Repositories.CreateKey(ctx, owner, name, deployKey); err != nil {
		return private, public, errors.New("failed to create deploy key: %w", err)
	}
	return private, public, nil
}

func (b *GitHubBootstrapper) generateSSHKeyPair() (private, public string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", errors.New("failed to generate an SSH key-pair for the GitHub repository deploy-key: %w", err)
	}
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := pem.Block{Type: "RSA PRIVATE KEY", Headers: nil, Bytes: privateDER}
	privatePEM := pem.EncodeToMemory(&privateBlock)
	publicRSAKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", errors.New("failed to generate the public-key part from an SSH key-pair meant to be used as a GitHub repository deploy-key: %w", err)
	}
	publicBytes := ssh.MarshalAuthorizedKey(publicRSAKey)
	return string(privatePEM), string(publicBytes), nil
}

func (b *GitHubBootstrapper) populateRepository(ctx context.Context, repo *git.Repository, auth *gossh.PublicKeys) error {

	// Get clone worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return errors.New("could not get worktree: %w", err)
	}
	clonePath := worktree.Filesystem.Root()

	// Extract devbot installation manifest into clone path
	var extract func(path string) error
	extract = func(path string) error {
		entries, err := resourcesFS.ReadDir(filepath.Join("resources", path))
		if err != nil {
			return errors.New("failed to read embedded directory '%s': %w", path, err)
		}
		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			if entry.IsDir() {
				if err := extract(entryPath); err != nil {
					return err
				}
			} else if data, err := resourcesFS.ReadFile(filepath.Join("resources", entryPath)); err != nil {
				return errors.New("failed to read embedded file '%s': %w", entryPath, err)
			} else {
				fullPath := filepath.Join(clonePath, entryPath)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errors.New("could not create directory '%s': %w", dir, err)
				} else if err := os.WriteFile(fullPath, data, 0644); err != nil {
					return errors.New("failed writing file '%s': %w", fullPath, err)
				} else if _, err := worktree.Add(entryPath); err != nil {
					return errors.New("failed adding file to worktree: %w", err)
				}
			}
		}
		return nil
	}
	if err := extract(""); err != nil {
		return errors.New("failed to extract embedded resources into repository clone: %w", err)
	}

	// Patch kustomization file to add image tags based on our version
	kustomizationFilePath := filepath.Join(clonePath, ".devbot", "kustomization.yaml")
	var kustomization types.Kustomization
	if kustomizationBytes, err := os.ReadFile(kustomizationFilePath); err != nil {
		return errors.New("failed to read kustomization.yaml: %w", err)
	} else if err := yaml.Unmarshal(kustomizationBytes, &kustomization); err != nil {
		return errors.New("failed unmarshalling YAML from '%s': %w", kustomizationFilePath, err)
	}

	// Instruct kustomization to install current version
	shaTag := "sha-" + semver.MustParse(version.Version).Build[0]
	kustomization.Images = append(
		kustomization.Images,
		types.Image{Name: "ghcr.io/arikkfir/devbot/controller", NewTag: shaTag},
		types.Image{Name: "ghcr.io/arikkfir/devbot/github-webhook", NewTag: shaTag},
	)

	// Instruct kustomization to add the build annotation & the version label
	if kustomization.CommonAnnotations == nil {
		kustomization.CommonAnnotations = make(map[string]string)
	}
	kustomization.CommonAnnotations["app.kubernetes.io/build"] = semver.MustParse(version.Version).Build[0]
	kustomization.Labels[0].Pairs["app.kubernetes.io/version"] = semver.MustParse(version.Version).FinalizeVersion()

	// If webhooks are enabled, pass the webhooks URL to the controller
	if b.webhooksURL != "" {
		p := []EnvVarJSONPatchOperation{
			{
				Op:    "add",
				Path:  "/spec/template/spec/containers/0/env/-",
				Value: v1.EnvVar{Name: "GITHUB_WEBHOOKS_URL", Value: b.webhooksURL},
			},
		}
		if patchBytes, err := json.Marshal(p); err != nil {
			return errors.New("failed marshalling patch env var: %w", err)
		} else {
			kustomization.Patches = append(kustomization.Patches, types.Patch{
				Patch: string(patchBytes),
				Target: &types.Selector{
					ResId: resid.ResId{
						Gvk: resid.Gvk{
							Group:   appsv1.SchemeGroupVersion.Group,
							Version: appsv1.SchemeGroupVersion.Version,
							Kind:    "Deployment",
						},
						Name:      "devbot-controller", // TODO: infer controller deployment name from kustomization resources
						Namespace: "devbot",            // TODO: infer namespace from kustomization resources
					},
				},
			})
		}
	}
	if kustomizationBytes, err := k8syaml.Marshal(&kustomization); err != nil {
		return errors.New("failed marshalling YAML: %w", err)
	} else if err := os.WriteFile(kustomizationFilePath, kustomizationBytes, 0644); err != nil {
		return errors.New("failed writing back patched kustomization YAML to '%s': %w", kustomizationFilePath, err)
	} else if _, err := worktree.Add(".devbot/kustomization.yaml"); err != nil {
		return errors.New("failed adding patched '%s' to worktree: %w", kustomizationFilePath, err)
	}

	// Commit
	commitOptions := &git.CommitOptions{
		AllowEmptyCommits: false,
		Author: &object.Signature{
			Name:  "Devbot",
			Email: "devbot@kfirs.com",
			When:  time.Now(),
		},
	}
	if _, err := worktree.Commit("Devbot bootstrap", commitOptions); err != nil {
		return errors.New("could not commit changes: %w", err)
	}

	// Push
	pushOptions := &git.PushOptions{
		Auth:       auth,
		Progress:   io.Discard,
		RefSpecs:   []config.RefSpec{"refs/heads/main:refs/heads/main"},
		RemoteName: "origin",
	}
	if err := repo.PushContext(ctx, pushOptions); err != nil {
		return errors.New("could not push changes: %w", err)
	}

	return nil
}

func (b *GitHubBootstrapper) deployToCluster(ctx context.Context, devbotKustomizePath string) error {

	// Build resources from the devbot kustomization
	resources, err := b.buildResourceMapFromClone(devbotKustomizePath)
	if err != nil {
		return errors.New("could not read resources: %w", err)
	}

	// Apply resources to the cluster
	for _, uns := range *resources {
		// Build a name for debugging which optionally includes the namespace, if it has one
		namespacedName := uns.GetName()
		if uns.GetNamespace() != "" {
			namespacedName = uns.GetNamespace() + "/" + namespacedName
		}

		// Used for discovering metadata about the resource kind
		gvk := uns.GroupVersionKind()

		// Translate the GVK to GVR
		mapping, err := b.k8sRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return errors.New("could not discover REST mapping for GVK '%s' of resource '%s': %w", gvk, namespacedName, err)
		}
		gvr := mapping.Resource

		// Use the GVR to discover all API resources matching its group & version, and then learn whether this GVR
		// is namespaced or not; then create the correct dynamic client accordingly
		gvkAPIResources, err := b.k8sDiscoveryClient.ServerResourcesForGroupVersion(gvr.GroupVersion().Identifier())
		if err != nil {
			return errors.New("could not discover API resources mapped for GVK '%s' of resource '%s': %w", gvk, namespacedName, err)
		}
		var c dynamic.ResourceInterface
		for _, apiResource := range gvkAPIResources.APIResources {
			if apiResource.Kind == gvk.Kind {
				if apiResource.Namespaced {
					c = b.k8sDynamicClient.Resource(gvr).Namespace(uns.GetNamespace())
				} else {
					c = b.k8sDynamicClient.Resource(gvr)
				}
				break
			}
		}
		if c == nil {
			return errors.New("could not discover API resource GVR '%s' of resource '%s': %w", gvr, namespacedName, err)
		}

		// Apply resource to the cluster
		log.Info().Str("namespace", uns.GetNamespace()).Str("name", uns.GetName()).Str("kind", gvk.String()).Msg("Applying resource...")
		if _, err := c.Apply(ctx, uns.GetName(), &uns, metav1.ApplyOptions{FieldManager: "devctl"}); err != nil {
			return errors.New("could not apply resource '%s': %w", namespacedName, err)
		}
	}

	return nil
}

func (b *GitHubBootstrapper) verifyDevbotIsReady(ctx context.Context) error {
	scheme := runtime.NewScheme()
	utilruntime.Must(apiextensions.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	c, err := client.New(b.k8sRestConfig, client.Options{Scheme: scheme})
	if err != nil {
		return errors.New("failed building Kubernetes clientset: %w", err)
	}

	verifyAnnotations := func(tmeta metav1.TypeMeta, ometa metav1.ObjectMeta) error {
		for annotationName, expectedValue := range expectedAnnotations {
			if actualValue, ok := ometa.Annotations[annotationName]; !ok {
				return fmt.Errorf("annotation '%s' missing for %s/%s '%s'", annotationName, tmeta.APIVersion, tmeta.Kind, ometa.Name)
			} else if actualValue != expectedValue {
				return fmt.Errorf("unexpected value '%s' found for annotation '%s' of %s/%s '%s': expected '%s'", actualValue, annotationName, tmeta.APIVersion, tmeta.Kind, ometa.Name, expectedValue)
			}
		}
		return nil
	}
	verifyLabels := func(tmeta metav1.TypeMeta, ometa metav1.ObjectMeta) error {
		for labelName, expectedValue := range expectedLabels {
			if actualValue, ok := ometa.Labels[labelName]; !ok {
				return fmt.Errorf("label '%s' missing for %s/%s '%s'", labelName, tmeta.APIVersion, tmeta.Kind, ometa.Name)
			} else if actualValue != expectedValue {
				return fmt.Errorf("unexpected value '%s' found for label '%s' of %s/%s '%s': expected '%s'", actualValue, labelName, tmeta.APIVersion, tmeta.Kind, ometa.Name, expectedValue)
			}
		}
		return nil
	}
	verifyCRD := func(name string) error {
		crd := &apiextensions.CustomResourceDefinition{}
		if err := c.Get(ctx, client.ObjectKey{Name: name}, crd); err != nil {
			return fmt.Errorf("failed looking up CRD '%s': %w", name, err)
		} else if err := verifyAnnotations(crd.TypeMeta, crd.ObjectMeta); err != nil {
			return fmt.Errorf("failed verifying annotations for CRD '%s': %w", name, err)
		} else if err := verifyLabels(crd.TypeMeta, crd.ObjectMeta); err != nil {
			return fmt.Errorf("failed verifying labels for CRD '%s': %w", name, err)
		} else {
			return nil
		}
	}
	verifyDeployment := func(name string) error {
		d := &appsv1.Deployment{}
		if err := c.Get(ctx, client.ObjectKey{Namespace: "devbot", Name: name}, d); err != nil {
			return fmt.Errorf("failed looking up deployment '%s': %w", name, err)
		} else if err := verifyAnnotations(d.TypeMeta, d.ObjectMeta); err != nil {
			return fmt.Errorf("failed verifying annotations for deployment '%s': %w", name, err)
		} else if err := verifyLabels(d.TypeMeta, d.ObjectMeta); err != nil {
			return fmt.Errorf("failed verifying labels for deployment '%s': %w", name, err)
		} else if d.Status.Replicas <= 0 {
			return fmt.Errorf("no replicas found for deployment '%s'", name)
		} else if d.Status.UpdatedReplicas <= 0 {
			return fmt.Errorf("no up-to-date replicas found for deployment '%s'", name)
		} else if d.Status.ReadyReplicas <= 0 {
			return fmt.Errorf("no ready replicas found for deployment '%s'", name)
		} else if d.Status.AvailableReplicas <= 0 {
			return fmt.Errorf("no available replicas found for deployment '%s'", name)
		} else if d.Status.UnavailableReplicas > 0 {
			return fmt.Errorf("%d unavailable replicas found for deployment '%s'", d.Status.UnavailableReplicas, name)
		} else if d.Status.AvailableReplicas != d.Status.ReadyReplicas {
			return fmt.Errorf("%d available replicas found in deployment '%s', but %d ready replicas were found for it", d.Status.AvailableReplicas, name, d.Status.ReadyReplicas)
		} else if d.Status.AvailableReplicas != d.Status.UpdatedReplicas {
			return fmt.Errorf("%d available replicas found in deployment '%s', but %d up-to-date replicas were found for it", d.Status.AvailableReplicas, name, d.Status.UpdatedReplicas)
		} else if d.Status.AvailableReplicas != d.Status.Replicas {
			return fmt.Errorf("%d available replicas found in deployment '%s', but %d replicas were found for it", d.Status.AvailableReplicas, name, d.Status.Replicas)
		} else {
			return nil
		}
	}

	verify := func() error {
		for _, name := range expectedCRDs {
			if err := verifyCRD(name); err != nil {
				return err
			}
		}
		for _, name := range expectedDeployments {
			if err := verifyDeployment(name); err != nil {
				return err
			}
		}
		return nil
	}

	ticker := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(b.timeout)
	for {
		select {
		case <-ctx.Done():
			return errors.New("terminated")
		case <-ticker.C:
			if err := verify(); err != nil {
				log.Warn().Err(err).Msg("Devbot is not ready yet")
			} else {
				return nil
			}
		case <-timeout.C:
			return fmt.Errorf("timed out waiting for Devbot to become ready")
		}
	}
}

func (b *GitHubBootstrapper) buildResourceMapFromClone(devbotPath string) (*[]unstructured.Unstructured, error) {
	k := krusty.MakeKustomizer(&krusty.Options{
		Reorder:           krusty.ReorderOptionNone,
		AddManagedbyLabel: false,
		LoadRestrictions:  types.LoadRestrictionsNone,
		PluginConfig: &types.PluginConfig{
			PluginRestrictions: types.PluginRestrictionsNone,
			FnpLoadingOptions: types.FnPluginLoadingOptions{
				EnableExec: true,
				EnableStar: false,
				Network:    false,
				Mounts:     nil, // TODO: reevaluate this, useful for bake job
				Env:        nil, // TODO: reevaluate this, useful for bake job
			},
			HelmConfig: types.HelmConfig{
				Enabled: false,
			},
		},
	})

	resourcesMap, err := k.Run(filesys.MakeFsOnDisk(), devbotPath)
	if err != nil {
		return nil, errors.New("could not build Kustomization: %w", err)
	}

	// Translate all resources into Unstructured objects
	var unstructuredObjects []unstructured.Unstructured
	for _, res := range resourcesMap.Resources() {
		gvk := res.GetGvk()
		name := res.GetName()
		if namespace := res.GetNamespace(); namespace != "" {
			name = namespace + "/" + name
		}

		uns := unstructured.Unstructured{}
		if yamlBytes, err := res.AsYAML(); err != nil {
			return nil, errors.New("could not convert resource '%s/%s' to YAML: %w", gvk.Kind, name, err)
		} else if jsonBytes, err := yaml.ToJSON(yamlBytes); err != nil {
			return nil, errors.New("could not convert YAML to JSON for resource '%s/%s': %w", gvk.Kind, name, err)
		} else if err := uns.UnmarshalJSON(jsonBytes); err != nil {
			return nil, errors.New("could not unmarshall JSON and create unstructured resource '%s/%s': %w", gvk.Kind, name, err)
		} else {
			unstructuredObjects = append(unstructuredObjects, uns)
		}
	}

	return &unstructuredObjects, nil
}
