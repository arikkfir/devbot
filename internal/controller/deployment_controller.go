package controller

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1 "github.com/arikkfir/devbot/api/v1"
	"github.com/arikkfir/devbot/internal/util/k8s"
	"github.com/arikkfir/devbot/internal/util/lang"
	stringsutil "github.com/arikkfir/devbot/internal/util/strings"
)

var (
	DeploymentFinalizer = "deployments.finalizers." + apiv1.GroupVersion.Group
	ApplyJobImage       = ""
	BakeJobImage        = ""
	CloneJobImage       = ""
)

type DeploymentReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	DisableJSONLogging bool
	LogLevel           string
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.executeReconciliation(ctx, req).ToResultAndError()
}

func (r *DeploymentReconciler) finalizeObject(_ *k8s.Reconciliation[*apiv1.Deployment]) error {
	// TODO: delete all resources created by this deployment
	return nil
}

func (r *DeploymentReconciler) executeReconciliation(ctx context.Context, req ctrl.Request) *k8s.Result {
	rec, result := k8s.NewReconciliation(ctx, r.Client, req, &apiv1.Deployment{}, DeploymentFinalizer, r.finalizeObject)
	if result != nil {
		return result
	}

	// Finalize object if deleted
	if result := rec.FinalizeObjectIfDeleted(); result != nil {
		return result
	}

	// Initialize the object if not initialized
	if result := rec.InitializeObject(); result != nil {
		return result
	}

	// Fetch source repository
	repoKey := rec.Object.Spec.Repository.GetObjectKey()
	repo := &apiv1.Repository{}
	if err := r.Client.Get(rec.Ctx, repoKey, repo); err != nil {
		if apierrors.IsNotFound(err) {
			rec.Object.Status.SetMaybeStaleDueToRepositoryNotFound("Repository '%s' could not be found", repoKey)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else if apierrors.IsForbidden(err) {
			rec.Object.Status.SetMaybeStaleDueToRepositoryNotAccessible("Repository '%s' is not accessible: %+v", repoKey, err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else {
			rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up repository '%s': %+v", repoKey, err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	}

	// Get controlling environment
	env := &apiv1.Environment{}
	if result := rec.GetRequiredController(env); result != nil {
		return result
	}

	// Get controlling application
	var app *apiv1.Application
	if result := r.getApplication(rec, env, &app); result != nil {
		return result
	}

	// Get repo settings from app
	var repoSettings *apiv1.ApplicationSpecRepository
	for _, appRepoSettings := range app.Spec.Repositories {
		if repoKey.Namespace == appRepoSettings.Namespace && repoKey.Name == appRepoSettings.Name {
			repoSettings = &appRepoSettings
			break
		}
	}
	if repoSettings == nil {
		rec.Object.Status.SetInvalidDueToInternalError("Repository settings of '%s' not found in application", repoKey)
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	// Ensure a persistent volume claim was created
	if result := r.ensurePersistentVolumeClaim(rec); result != nil {
		return result
	}

	// Get the current job, if any
	var job *batchv1.Job
	if result := r.getLastJob(rec, &job); result != nil {
		return result
	}

	// Infer the branch to deploy
	var branch, revision string
	if r, ok := repo.Status.Revisions[env.Spec.PreferredBranch]; ok {
		branch = env.Spec.PreferredBranch
		revision = r
	} else if r, ok := repo.Status.Revisions[repo.Status.DefaultBranch]; ok {
		branch = repo.Status.DefaultBranch
		revision = r
	} else {
		rec.Object.Status.SetMaybeStaleDueToBranchNotFound("Neither branch '%s' nor '%s' found in repository '%s'", env.Spec.PreferredBranch, repo.Status.DefaultBranch, client.ObjectKeyFromObject(repo))
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// If no current job is running, we may want to start from scratch (clone->bake->apply) if branch/revision changed
	if job == nil {

		// If either branch or revision changed, update the status & create a new clone job
		branchChanged := branch != rec.Object.Status.Branch
		if branchChanged {
			rec.Object.Status.Branch = branch
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
		}
		revisionChanged := revision != rec.Object.Status.LastAppliedRevision
		if revisionChanged {
			rec.Object.Status.LastAttemptedRevision = revision
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
		}
		if branchChanged || revisionChanged {
			return r.createNewCloneJob(rec, app, repo)
		}
		return k8s.DoNotRequeue()
	}

	// If branch/revision changed we should wait for the current job to complete, abandon it, update our status, and start from scratch
	if branch != rec.Object.Status.Branch || revision != rec.Object.Status.LastAttemptedRevision {
		if job.Status.Active > 0 {
			// Wait until the currently running job is finished (successfully or not)
			return k8s.RequeueAfter(5 * time.Second)
		}
		rec.Object.Status.Branch = branch
		rec.Object.Status.LastAttemptedRevision = revision
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return r.createNewCloneJob(rec, app, repo)
	}

	// We have a job - progress accordingly
	phase := Phase(job.Labels[PhaseLabel])
	for _, c := range job.Status.Conditions {
		switch c.Type {

		case batchv1.JobSuspended:
			// If the job is suspended or possibly suspended, update the status and wait for it to progress/unsuspend
			switch c.Status {
			case corev1.ConditionTrue, corev1.ConditionUnknown:
				switch phase {
				case PhaseClone:
					rec.Object.Status.SetMaybeStaleDueToCloning("Waiting for suspended clone job")
				case PhaseBake:
					rec.Object.Status.SetMaybeStaleDueToBaking("Waiting for suspended bake job")
				case PhaseApply:
					rec.Object.Status.SetMaybeStaleDueToApplying("Waiting for suspended apply job")
				default:
					panic("unsupported phase: " + phase)
				}
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.DoNotRequeue()
			}

		case batchv1.JobFailed:
			switch c.Status {
			case corev1.ConditionUnknown:
				// Job possibly failed - wait until we know for sure
				switch phase {
				case PhaseClone:
					rec.Object.Status.SetMaybeStaleDueToCloning("Waiting for possibly failed clone job")
				case PhaseBake:
					rec.Object.Status.SetMaybeStaleDueToBaking("Waiting for possibly failed bake job")
				case PhaseApply:
					rec.Object.Status.SetMaybeStaleDueToApplying("Waiting for possibly failed apply job")
				default:
					panic("unsupported phase: " + phase)
				}
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.DoNotRequeue()

			case corev1.ConditionTrue:
				// Job failed - recreate it
				switch phase {
				case PhaseClone:
					return r.createNewCloneJob(rec, app, repo)
				case PhaseBake:
					return r.createNewBakeJob(rec, app, env, repo, *repoSettings)
				case PhaseApply:
					return r.createNewApplyJob(rec, app, env)
				default:
					panic("unsupported phase: " + phase)
				}
			}

		case batchv1.JobComplete:
			switch c.Status {
			case corev1.ConditionUnknown:
				// Job possibly completed - wait until we know for sure
				switch phase {
				case PhaseClone:
					rec.Object.Status.SetMaybeStaleDueToCloning("Waiting for possibly completed clone job")
				case PhaseBake:
					rec.Object.Status.SetMaybeStaleDueToBaking("Waiting for possibly completed bake job")
				case PhaseApply:
					rec.Object.Status.SetMaybeStaleDueToApplying("Waiting for possibly completed apply job")
				default:
					panic("unsupported phase: " + phase)
				}
				if result := rec.UpdateStatus(); result != nil {
					return result
				}
				return k8s.DoNotRequeue()

			case corev1.ConditionTrue:
				// Job completed successfully - create the next one (or clear it entirely; we're done)
				switch phase {
				case PhaseClone:
					return r.createNewBakeJob(rec, app, env, repo, *repoSettings)
				case PhaseBake:
					return r.createNewApplyJob(rec, app, env)
				case PhaseApply:
					rec.Object.Status.SetCurrent()
					rec.Object.Status.LastAppliedRevision = rec.Object.Status.LastAttemptedRevision
					if result := rec.UpdateStatus(); result != nil {
						return result
					}
					return k8s.DoNotRequeue()
				default:
					panic("unsupported phase: " + phase)
				}
			}
		}
	}

	return k8s.DoNotRequeue()
}

func (r *DeploymentReconciler) getApplication(rec *k8s.Reconciliation[*apiv1.Deployment], env *apiv1.Environment, appTarget **apiv1.Application) *k8s.Result {
	appRef := metav1.GetControllerOf(env)
	if appRef == nil {
		rec.Object.Status.SetInvalidDueToInternalError("Could not find application controller for parent environment '%s/%s'", env.Namespace, env.Name)
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	app := &apiv1.Application{}
	if err := r.Client.Get(rec.Ctx, client.ObjectKey{Name: appRef.Name, Namespace: env.Namespace}, app); err != nil {
		if apierrors.IsNotFound(err) {
			rec.Object.Status.SetInvalidDueToControllerNotFound("Could not find application controller of parent environment: %+v", err)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else if apierrors.IsForbidden(err) {
			rec.Object.Status.SetInvalidDueToControllerNotAccessible("Application controller of parent environment is not accessible: %+v", err)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		} else {
			rec.Object.Status.SetInvalidDueToInternalError("Failed to get application controller of parent environment: %+v", err)
			rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	}
	rec.Object.Status.SetValidIfInvalidDueToAnyOf(apiv1.ControllerNotAccessible, apiv1.ControllerNotFound, apiv1.InternalError, apiv1.RepositoryNotSupported)
	rec.Object.Status.SetCurrentIfStaleDueToAnyOf(apiv1.Invalid)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	*appTarget = app
	return k8s.Continue()
}

func (r *DeploymentReconciler) ensurePersistentVolumeClaim(rec *k8s.Reconciliation[*apiv1.Deployment]) *k8s.Result {
	if rec.Object.Status.PersistentVolumeClaimName == "" {
		rec.Object.Status.SetMaybeStaleDueToPersistentVolumeMissing("Persistent volume claim name not set yet")
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		rec.Object.Status.PersistentVolumeClaimName = rec.Object.Name + "-work"
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
	}

	pvc := &corev1.PersistentVolumeClaim{}
	if err := rec.Client.Get(rec.Ctx, client.ObjectKey{Namespace: rec.Object.Namespace, Name: rec.Object.Status.PersistentVolumeClaimName}, pvc); err != nil && apierrors.IsNotFound(err) {
		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:            rec.Object.Status.PersistentVolumeClaimName,
				Namespace:       rec.Object.Namespace,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rec.Object, apiv1.DeploymentGVK)},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: *resource.NewScaledQuantity(5, resource.Giga),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceStorage: *resource.NewScaledQuantity(15, resource.Giga),
					},
				},
				StorageClassName: lang.Ptr("standard"), // TODO: storage-class should be user-configurable
			},
		}
		if err := rec.Client.Create(rec.Ctx, pvc); err != nil {
			rec.Object.Status.SetMaybeStaleDueToPersistentVolumeCreationFailed("Failed creating persistent volume claim: %+v", err)
			if result := rec.UpdateStatus(); result != nil {
				return result
			}
			return k8s.Requeue()
		}
	} else if err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed looking up persistent volume claim '%s': %+v", rec.Object.Status.PersistentVolumeClaimName, err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}
	return nil
}

func (r *DeploymentReconciler) getLastJob(rec *k8s.Reconciliation[*apiv1.Deployment], target **batchv1.Job) *k8s.Result {
	jobs := &batchv1.JobList{}
	nsFilter := client.InNamespace(rec.Object.Namespace)
	ownershipFilter := k8s.OwnedBy(r.Client.Scheme(), rec.Object)
	if err := r.Client.List(rec.Ctx, jobs, nsFilter, ownershipFilter); err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed listing jobs: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// If no jobs found, return nil
	if len(jobs.Items) == 0 {
		*target = nil
		return nil
	}

	// Sort jobs by creation timestamp and take the last one
	slices.SortFunc(jobs.Items, func(i, j batchv1.Job) int {
		return i.CreationTimestamp.Time.Compare(j.CreationTimestamp.Time)
	})
	*target = &jobs.Items[len(jobs.Items)-1]
	return nil
}

func (r *DeploymentReconciler) createNewJobSpec(rec *k8s.Reconciliation[*apiv1.Deployment], image string, phase Phase, app *apiv1.Application, envVars ...corev1.EnvVar) (batchv1.Job, error) {
	log.FromContext(rec.Ctx).WithValues("phase", phase).Info("Creating new job")

	// Prepare job environment variables
	jobEnvVars := []corev1.EnvVar{
		{Name: "DISABLE_JSON_LOGGING", Value: strconv.FormatBool(r.DisableJSONLogging)},
		{Name: "LOG_LEVEL", Value: r.LogLevel},
	}
	for _, kv := range os.Environ() {
		tokens := strings.SplitN(kv, "=", 2)
		if len(tokens) == 2 {
			name := tokens[0]
			value := tokens[1]
			if strings.HasPrefix(name, "OTEL_") && !strings.HasPrefix(name, "OTEL_SERVICE_") {
				jobEnvVars = append(jobEnvVars, corev1.EnvVar{Name: name, Value: value})
			}
		}
	}
	jobEnvVars = append(jobEnvVars, envVars...)

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            stringsutil.RandomHash(7),
			Namespace:       rec.Object.Namespace,
			Labels:          map[string]string{PhaseLabel: string(phase)},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(rec.Object, apiv1.DeploymentGVK)},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            lang.Ptr(int32(10)),                          // TODO: make number of retries configurable
			TTLSecondsAfterFinished: lang.Ptr(int32((5 * time.Minute).Seconds())), // TODO: make TTL configurable
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:       string(phase),
							Image:      image,
							WorkingDir: "/data",
							Env:        jobEnvVars,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewScaledQuantity(100, resource.Milli),
									corev1.ResourceMemory: *resource.NewScaledQuantity(128, resource.Mega),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewScaledQuantity(200, resource.Milli),
									corev1.ResourceMemory: *resource.NewScaledQuantity(256, resource.Mega),
								},
							},
							VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}},
						},
					},
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: app.Spec.ServiceAccountName,
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: rec.Object.Status.PersistentVolumeClaimName,
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (r *DeploymentReconciler) createNewCloneJob(rec *k8s.Reconciliation[*apiv1.Deployment], app *apiv1.Application, repo *apiv1.Repository) *k8s.Result {
	var url string

	// Calculate Git URL based on repository type
	if repo.Spec.GitHub != nil {
		url = fmt.Sprintf("https://github.com/%s/%s", repo.Spec.GitHub.Owner, repo.Spec.GitHub.Name)
	} else {
		rec.Object.Status.SetInvalidDueToRepositoryNotSupported("Unsupported repository")
		rec.Object.Status.SetMaybeStaleDueToInvalid(rec.Object.Status.GetInvalidMessage())
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.DoNotRequeue()
	}

	// Set cloning status
	rec.Object.Status.SetMaybeStaleDueToCloning("Launching clone job")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Create the job object
	job, err := r.createNewJobSpec(rec, CloneJobImage, PhaseClone, app, corev1.EnvVar{Name: "BRANCH", Value: rec.Object.Status.Branch}, corev1.EnvVar{Name: "GIT_URL", Value: url}, corev1.EnvVar{Name: "SHA", Value: rec.Object.Status.LastAttemptedRevision})
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed creating clone job spec: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Create the job in the cluster
	log.FromContext(rec.Ctx).WithValues("jobName", job.Name).Info("Creating clone job")
	if err := r.Client.Create(rec.Ctx, &job); err != nil {
		rec.Object.Status.SetMaybeStaleDueToCloneFailed("Failed creating clone job: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Update status to reflect we're waiting for the clone job to finish
	rec.Object.Status.SetMaybeStaleDueToCloning("Waiting for clone job '%s' to finish", job.Name)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// No requeue necessary - job completion/failure/suspension will trigger reconciliation
	return k8s.DoNotRequeue()
}

func (r *DeploymentReconciler) createNewBakeJob(rec *k8s.Reconciliation[*apiv1.Deployment], app *apiv1.Application, env *apiv1.Environment, repo *apiv1.Repository, repoSettings apiv1.ApplicationSpecRepository) *k8s.Result {
	// Create the job object
	job, err := r.createNewJobSpec(
		rec,
		BakeJobImage,
		PhaseBake,
		app,
		corev1.EnvVar{Name: "ACTUAL_BRANCH", Value: rec.Object.Status.Branch},
		corev1.EnvVar{Name: "APPLICATION_NAME", Value: app.Name},
		corev1.EnvVar{Name: "BASE_DEPLOY_DIR", Value: repoSettings.Path},
		corev1.EnvVar{Name: "ENVIRONMENT_NAME", Value: env.Name},
		corev1.EnvVar{Name: "DEPLOYMENT_NAME", Value: rec.Object.Name},
		corev1.EnvVar{Name: "MANIFEST_FILE", Value: ".devbot.yaml"},
		corev1.EnvVar{Name: "PREFERRED_BRANCH", Value: env.Spec.PreferredBranch},
		corev1.EnvVar{Name: "REPO_DEFAULT_BRANCH", Value: repo.Status.DefaultBranch},
		corev1.EnvVar{Name: "SHA", Value: rec.Object.Status.LastAttemptedRevision},
	)
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed creating bake job spec: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Set a cloning status
	rec.Object.Status.SetMaybeStaleDueToCloning("Launching bake job")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Create the job in the cluster
	log.FromContext(rec.Ctx).WithValues("jobName", job.Name).Info("Creating bake job")
	if err := r.Client.Create(rec.Ctx, &job); err != nil {
		rec.Object.Status.SetMaybeStaleDueToBakingFailed("Failed creating bake job: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Update status to reflect we're waiting for the clone job to finish
	rec.Object.Status.SetMaybeStaleDueToBaking("Waiting for bake job '%s' to finish", job.Name)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// No requeue necessary - job completion/failure/suspension will trigger reconciliation
	return k8s.DoNotRequeue()
}

func (r *DeploymentReconciler) createNewApplyJob(rec *k8s.Reconciliation[*apiv1.Deployment], app *apiv1.Application, env *apiv1.Environment) *k8s.Result {
	// Create the job object
	job, err := r.createNewJobSpec(rec, ApplyJobImage, PhaseApply, app,
		corev1.EnvVar{Name: "APPLICATION_NAME", Value: app.Name},
		corev1.EnvVar{Name: "ENVIRONMENT_NAME", Value: env.Name},
		corev1.EnvVar{Name: "DEPLOYMENT_NAME", Value: rec.Object.Name},
		corev1.EnvVar{Name: "MANIFEST_FILE", Value: ".devbot.yaml"},
	)
	if err != nil {
		rec.Object.Status.SetMaybeStaleDueToInternalError("Failed creating apply job spec: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Set a cloning status
	rec.Object.Status.SetMaybeStaleDueToCloning("Launching apply job")
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// Create the job in the cluster
	log.FromContext(rec.Ctx).WithValues("jobName", job.Name).Info("Creating apply job")
	if err := r.Client.Create(rec.Ctx, &job); err != nil {
		rec.Object.Status.SetMaybeStaleDueToApplyFailed("Failed creating apply job: %+v", err)
		if result := rec.UpdateStatus(); result != nil {
			return result
		}
		return k8s.Requeue()
	}

	// Update status to reflect we're waiting for the clone job to finish
	rec.Object.Status.SetMaybeStaleDueToBaking("Waiting for apply job '%s' to finish", job.Name)
	if result := rec.UpdateStatus(); result != nil {
		return result
	}

	// No requeue necessary - job completion/failure/suspension will trigger reconciliation
	return k8s.DoNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO: watch our environment's application also, and reconcile upon repository configuration changes
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// Only reconcile if the generation has changed
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			},
		})).
		Watches(&batchv1.Job{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			job := obj.(*batchv1.Job)
			controllerRef := metav1.GetControllerOf(job)
			if controllerRef == nil {
				return nil
			} else if controllerRef.APIVersion != apiv1.DeploymentGVK.GroupVersion().String() {
				return nil
			} else if controllerRef.Kind != apiv1.DeploymentGVK.Kind {
				return nil
			} else {
				return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: job.Namespace, Name: controllerRef.Name}}}
			}
		})).
		Watches(&apiv1.Repository{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
			repo := obj.(*apiv1.Repository)
			repoKey := client.ObjectKeyFromObject(repo)

			deploymentsList := &apiv1.DeploymentList{}
			if err := r.List(ctx, deploymentsList); err != nil {
				log.FromContext(ctx).Error(err, "Failed to list deployments")
				return nil
			}

			var requests []ctrl.Request
			for _, d := range deploymentsList.Items {
				if d.Spec.Repository.GetObjectKey() == repoKey {
					// TODO: we should only trigger a reconciliation if the repository:
					//       1. Has a newer revision for the branch this deployment's branch
					//       2. This deployment does not have its environment's preferred branch, and the repository's default branch changed
					requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&d)})
				}
			}
			return requests
		})).
		Complete(r)
}
