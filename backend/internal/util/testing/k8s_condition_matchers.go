package testing

import (
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BeTrueDueTo(args ...string) types.GomegaMatcher {
	if len(args) == 0 {
		panic("BeTrueDueTo expected at least one argument")
	} else if len(args) == 1 {
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionTrue), HaveField("Reason", args[0]))
	} else if len(args) == 2 {
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionTrue), HaveField("Reason", args[0]), HaveField("Message", MatchRegexp(args[1])))
	} else {
		var reArgs []interface{}
		for _, arg := range args[2:] {
			reArgs = append(reArgs, arg)
		}
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionTrue), HaveField("Reason", args[0]), HaveField("Message", MatchRegexp(args[1], reArgs...)))
	}
}

func BeFalseDueTo(args ...string) types.GomegaMatcher {
	if len(args) == 0 {
		panic("BeFalseDueTo expected at least one argument")
	} else if len(args) == 1 {
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionFalse), HaveField("Reason", args[0]))
	} else if len(args) == 2 {
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionFalse), HaveField("Reason", args[0]), HaveField("Message", MatchRegexp(args[1])))
	} else {
		var reArgs []interface{}
		for _, arg := range args[2:] {
			reArgs = append(reArgs, arg)
		}
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionFalse), HaveField("Reason", args[0]), HaveField("Message", MatchRegexp(args[1], reArgs...)))
	}
}

func BeUnknownDueTo(args ...string) types.GomegaMatcher {
	if len(args) == 0 {
		panic("BeFalseDueTo expected at least one argument")
	} else if len(args) == 1 {
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionUnknown), HaveField("Reason", args[0]))
	} else if len(args) == 2 {
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionUnknown), HaveField("Reason", args[0]), HaveField("Message", MatchRegexp(args[1])))
	} else {
		var reArgs []interface{}
		for _, arg := range args[2:] {
			reArgs = append(reArgs, arg)
		}
		return And(Not(BeNil()), HaveField("Status", metav1.ConditionUnknown), HaveField("Reason", args[0]), HaveField("Message", MatchRegexp(args[1], reArgs...)))
	}
}
