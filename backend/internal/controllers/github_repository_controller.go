package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	apiv1 "github.com/arikkfir/devbot/backend/api/v1"
	"github.com/arikkfir/devbot/backend/internal/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/secureworks/errors"
	"io"
	"io/fs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"slices"
	"strings"
	"time"
)

var (
	GitHubRepositoryFinalizer = "github.repositories.finalizers." + apiv1.GroupVersion.Group
)

type GitHubRepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type ConditionInstruction struct {
	True    bool
	False   bool
	Unknown bool
	Reason  string
	Message string
}

func (r *GitHubRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	repo := &apiv1.GitHubRepository{}

	// Ensure deletion & finalizers lifecycle processing
	if result, err := util.PrepareReconciliation(ctx, r.Client, req, repo, GitHubRepositoryFinalizer); result != nil || err != nil {
		return *result, err
	}

	// Initialize conditions
	if repo.GetStatusConditionCloned() == nil {
		repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonPending, "Pending initialization")
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to initialize conditions in status of '%s'", req.NamespacedName, err)
		}
	}
	if repo.GetStatusConditionCurrent() == nil {
		repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonPending, "Pending initialization")
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to initialize conditions in status of '%s'", req.NamespacedName, err)
		}
	}

	// Ensure Git URL has been built & set on the object
	correctURL := fmt.Sprintf("https://github.com/%s/%s.git", repo.Spec.Owner, repo.Spec.Name)
	if repo.Status.GitURL != correctURL {
		repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonPending, "Changing Git URL")
		repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonPending, "Changing Git URL")
		repo.Status.GitURL = correctURL
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update Git URL in status of '%s'", req.NamespacedName, err)
		}
	}

	// Ensure local clone path has been set on the object
	correctLocalPath := fmt.Sprintf("/clones/%s/%s", repo.Spec.Owner, repo.Spec.Name)
	if repo.Status.LocalClonePath != correctLocalPath {
		// TODO: delete old local clone path, if it exists
		repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonPending, "Changing local clone path")
		repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonPending, "Changing local clone path")
		repo.Status.LocalClonePath = correctLocalPath
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update local clone path in status of '%s'", req.NamespacedName, err)
		}
	}

	// Ensure parent directory of the local clone path exists & ready
	if err := os.MkdirAll(filepath.Dir(repo.Status.LocalClonePath), 0700); err != nil {
		repo.SetStatusConditionCloned(metav1.ConditionUnknown, apiv1.ReasonMkdirFailed, "Failed to create parent directory of local clone path")
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonMkdirFailed, "Failed to create parent directory of local clone path")
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to create parent directory of local clone at '%s'", repo.Status.LocalClonePath, err)
	}

	// If clone path is missing, perform a clone
	if _, err := os.Stat(repo.Status.LocalClonePath); errors.Is(err, fs.ErrNotExist) {
		// Clone the repository

		var auth transport.AuthMethod
		if repo.Spec.Auth.PersonalAccessToken != nil {
			key := repo.Spec.Auth.PersonalAccessToken.Key
			if key == "" {
				repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonConfigError, "Secret key missing")
				repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonConfigError, "Secret key missing")
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, nil
			}

			secretNamespace := repo.Spec.Auth.PersonalAccessToken.Secret.Namespace
			if secretNamespace == "" {
				secretNamespace = req.Namespace
			}

			secretName := repo.Spec.Auth.PersonalAccessToken.Secret.Name
			if secretName == "" {
				repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonConfigError, "Secret name missing")
				repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonConfigError, "Secret name missing")
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, nil
			}

			secret := &corev1.Secret{}
			if err := r.Get(ctx, client.ObjectKey{Namespace: secretNamespace, Name: secretName}, secret); err != nil {
				repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonConfigError, "Failed getting secret: "+err.Error())
				repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonConfigError, "Failed getting secret: "+err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to get secret '%s/%s'", secretNamespace, secretName, err)
			}

			token := string(secret.Data[key])
			if token == "" {
				repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonConfigError, "Token missing or empty in secret")
				repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonConfigError, "Token missing or empty in secret")
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}

			auth = &http.BasicAuth{
				Username: util.RandomHash(7), // this is ignored by GitHub when using PATs
				Password: token,
			}
		} else {
			repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonConfigError, "Auth not configured")
			repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonConfigError, "Auth not configured")
			if err := r.Status().Update(ctx, repo); err != nil {
				return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
			}
			return ctrl.Result{}, nil
		}

		cloneOptions := &git.CloneOptions{
			URL:  repo.Status.GitURL,
			Auth: auth,
		}

		if _, err := git.PlainCloneContext(ctx, repo.Status.LocalClonePath, false, cloneOptions); err != nil {
			repo.SetStatusConditionCloned(metav1.ConditionFalse, apiv1.ReasonCloneFailed, err.Error())
			repo.SetStatusConditionCurrent(metav1.ConditionFalse, apiv1.ReasonCloneFailed, err.Error())
			if err := r.Status().Update(ctx, repo); err != nil {
				return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
			}
			return ctrl.Result{}, errors.New("failed to clone '%s' to '%s'", cloneOptions.URL, repo.Status.LocalClonePath, err)
		}

	} else if err != nil {
		// Inspecting the target clone path failed

		repo.SetStatusConditionCloned(metav1.ConditionUnknown, apiv1.ReasonStatFailed, err.Error())
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonStatFailed, err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to stat local directory at '%s'", repo.Status.LocalClonePath, err)
	}

	// Ensure "Cloned" condition is set to "True"
	if repo.SetStatusConditionClonedIfDifferent(metav1.ConditionTrue, "Cloned", "Cloned successfully") {
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Clone exists; open it
	gitRepo, err := git.PlainOpen(repo.Status.LocalClonePath)
	if err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonCloneOpenFailed, err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to open clone at '%s'", repo.Status.LocalClonePath, err)
	}

	// Get all branches & tags, create GitHubRepositoryRef CRs for each one
	var branchNames, tagNames []string
	if branches, err := gitRepo.Branches(); err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefInspectionFailed, err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to list repository branches", err)
	} else {
		for {
			if branch, err := branches.Next(); err != nil && !errors.Is(err, io.EOF) {
				repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefInspectionFailed, err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to iterate repository branches", err)
			} else if branch != nil {
				refName := branch.Name().String()
				branchNames = append(branchNames, refName)
				if err := r.ensureGitHubRepositoryRef(ctx, repo, refName); err != nil {
					repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefInspectionFailed, err.Error())
					if err := r.Status().Update(ctx, repo); err != nil {
						return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
					}
					return ctrl.Result{}, errors.New("failed to create ref object for '%s'", refName, err)
				}
			} else {
				break
			}
		}
	}
	if tags, err := gitRepo.Tags(); err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefInspectionFailed, err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to list repository tags", err)
	} else {
		for {
			if tag, err := tags.Next(); err != nil && !errors.Is(err, io.EOF) {
				repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefInspectionFailed, err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to iterate repository tags", err)
			} else if tag != nil {
				refName := tag.Name().String()
				tagNames = append(tagNames, refName)
				if err := r.ensureGitHubRepositoryRef(ctx, repo, refName); err != nil {
					repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonRefInspectionFailed, err.Error())
					if err := r.Status().Update(ctx, repo); err != nil {
						return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
					}
					return ctrl.Result{}, errors.New("failed to create ref object for '%s'", refName, err)
				}
			} else {
				break
			}
		}
	}

	// Delete stale GitHubRepositoryRef objects (i.e. if their ref no longer exists in the Git repository)
	refObjects := &apiv1.GitHubRepositoryRefList{}
	if err := r.List(ctx, refObjects); err != nil {
		repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonObjPruneFailed, err.Error())
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{}, errors.New("failed to list ref objects", err)
	}
	allRefs := append(branchNames, tagNames...)
	for _, refObject := range refObjects.Items {
		if !slices.Contains(allRefs, refObject.Spec.Ref) {
			if err := r.Delete(ctx, &refObject); err != nil {
				repo.SetStatusConditionCurrent(metav1.ConditionUnknown, apiv1.ReasonObjPruneFailed, err.Error())
				if err := r.Status().Update(ctx, repo); err != nil {
					return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
				}
				return ctrl.Result{}, errors.New("failed to delete ref object for '%s'", refObject.Spec.Ref, err)
			}
		}
	}

	// Ensure "Current" condition is set to "True"
	if repo.SetStatusConditionCurrentIfDifferent(metav1.ConditionTrue, "Synced", "Synchronized refs") {
		if err := r.Status().Update(ctx, repo); err != nil {
			return ctrl.Result{}, errors.New("failed to update status of '%s/%s'", repo.Namespace, repo.Name, err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *GitHubRepositoryReconciler) ensureGitHubRepositoryRef(ctx context.Context, repo *apiv1.GitHubRepository, refName string) error {
	// TODO: use other means than labels to lookup GitHubRepositoryRef objects for a given repo & ref
	repoOwnerAsBase64 := strings.Trim(base64.StdEncoding.EncodeToString([]byte(repo.Spec.Owner)), "-=")
	repoNameAsBase64 := strings.Trim(base64.StdEncoding.EncodeToString([]byte(repo.Spec.Name)), "-=")
	refNameAsBase64 := strings.Trim(base64.StdEncoding.EncodeToString([]byte(refName)), "-=")
	refLabels := client.MatchingLabels{
		"github.devbot.kfirs.com/repository-owner": repoOwnerAsBase64,
		"github.devbot.kfirs.com/repository-name":  repoNameAsBase64,
		"github.devbot.kfirs.com/ref":              refNameAsBase64,
	}

	refObjects := &apiv1.GitHubRepositoryRefList{}
	if err := r.List(ctx, refObjects, client.InNamespace(repo.Namespace), refLabels); err != nil {
		return errors.New("failed listing GitHub repository refs", err)
	}
	if len(refObjects.Items) == 0 {
		refObject := &apiv1.GitHubRepositoryRef{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    refLabels,
				Name:      util.RandomHash(7),
				Namespace: repo.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: repo.APIVersion,
						Kind:       repo.Kind,
						Name:       repo.Name,
						UID:        repo.UID,
					},
				},
			},
			Spec: apiv1.GitHubRepositoryRefSpec{
				Ref: refName,
			},
		}
		if err := r.Create(ctx, refObject); err != nil {
			return errors.New("failed creating ref object '%s/%s' for repo '%s'", refObject.Namespace, refObject.Name, repo.Name, err)
		}

	} else if len(refObjects.Items) > 1 {
		var names []string
		for _, refObj := range refObjects.Items {
			names = append(names, refObj.Name)
		}
		return errors.New("multiple ref objects match repo '%s' and ref '%s': %v", repo.Name, refName, names)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitHubRepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.GitHubRepository{}).
		Complete(r)
}
