package expectations

import (
	"regexp"

	. "github.com/arikkfir/justest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/arikkfir/devbot/backend/api/v1"
)

type (
	ConditionE struct {
		Type    string
		Status  *string
		Reason  *regexp.Regexp
		Message *regexp.Regexp
	}
	RepositorySpecE   struct{}
	RepositoryStatusE struct {
		Conditions    map[string]*ConditionE
		DefaultBranch string
		Revisions     map[string]string
	}
	RepositoryE struct {
		Name   string
		Spec   RepositorySpecE
		Status RepositoryStatusE
	}
	ResourceE struct {
		Object    client.Object
		Name      string
		Namespace string
		Validator func(T, ResourceE)
	}
	DeploymentSpecE struct {
		Repository v1.DeploymentRepositoryReference
	}
	DeploymentStatusE struct {
		Branch                string
		Conditions            map[string]*ConditionE
		LastAttemptedRevision string
		LastAppliedRevision   string
		ResolvedRepository    string
	}
	DeploymentE struct {
		Spec      DeploymentSpecE
		Status    DeploymentStatusE
		Resources []ResourceE
	}
	EnvSpecE   struct{ PreferredBranch string }
	EnvStatusE struct{ Conditions map[string]*ConditionE }
	EnvE       struct {
		Spec        EnvSpecE
		Status      EnvStatusE
		Deployments []DeploymentE
	}
	AppStatusE struct {
		Conditions map[string]*ConditionE
	}
	AppE struct {
		Name         string
		Status       AppStatusE
		Environments []EnvE
	}
	E struct {
		Applications map[string]AppE
	}
)
