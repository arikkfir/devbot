package v1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (o *Deployment) SetStatusConditionDeployingIfDifferent(status v1.ConditionStatus, reason, message string) bool {
	for i, c := range o.Status.Conditions {
		if c.Type == ConditionTypeDeploying {
			if c.Status != status || c.Reason != reason || c.Message != message {
				c.Status = status
				c.Reason = reason
				c.Message = message
				c.LastTransitionTime = v1.Now()
				c.ObservedGeneration = o.GetGeneration()
				o.Status.Conditions[i] = c
				return true
			}
			return false
		}
	}
	o.Status.Conditions = append(o.Status.Conditions, v1.Condition{
		Type:               ConditionTypeDeploying,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: v1.Now(),
		ObservedGeneration: o.GetGeneration(),
	})
	return true
}

func (o *Deployment) SetStatusConditionDeploying(status v1.ConditionStatus, reason, message string) {
	for i, c := range o.Status.Conditions {
		if c.Type == ConditionTypeDeploying {
			c.Status = status
			c.Reason = reason
			c.Message = message
			c.LastTransitionTime = v1.Now()
			c.ObservedGeneration = o.GetGeneration()
			o.Status.Conditions[i] = c
			return
		}
	}
	o.Status.Conditions = append(o.Status.Conditions, v1.Condition{
		Type:               ConditionTypeDeploying,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: v1.Now(),
		ObservedGeneration: o.GetGeneration(),
	})
}

func (o *Deployment) RemoveStatusConditionDeploying() {
	var newConditions []v1.Condition
	for _, c := range o.Status.Conditions {
		if c.Type != ConditionTypeDeploying {
			newConditions = append(newConditions, c)
		}
	}
	o.Status.Conditions = newConditions
}

func (o *Deployment) GetStatusConditionDeploying() *v1.Condition {
	for _, c := range o.Status.Conditions {
		if c.Type == ConditionTypeDeploying {
			return &c
		}
	}
	return nil
}
