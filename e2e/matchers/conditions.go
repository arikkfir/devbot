package matchers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Type string
type Status string
type Reason string
type Message types.GomegaMatcher

func ConditionWith(params ...interface{}) types.GomegaMatcher {
	var matchers []types.GomegaMatcher
	for _, param := range params {
		switch v := param.(type) {
		case Type:
			matchers = append(matchers, HaveField("Type", Equal(string(v))))
		case Status:
			matchers = append(matchers, HaveField("Status", Equal(v1.ConditionStatus(v))))
		case Reason:
			matchers = append(matchers, HaveField("Reason", Equal(string(v))))
		case Message:
			matchers = append(matchers, HaveField("Message", types.GomegaMatcher(v)))
		default:
			Fail(fmt.Sprintf("Unknown check '%T' in Condition()", v))
		}
	}
	return And(matchers...)
}
