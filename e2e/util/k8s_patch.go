package util

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JSONPatchOperation string

const (
	JSONPatchOperationAdd     JSONPatchOperation = "add"
	JSONPatchOperationRemove  JSONPatchOperation = "remove"
	JSONPatchOperationReplace JSONPatchOperation = "replace"
	JSONPatchOperationCopy    JSONPatchOperation = "copy"
	JSONPatchOperationMove    JSONPatchOperation = "move"
	JSONPatchOperationTest    JSONPatchOperation = "test"
)

type JSONPatchItem struct {
	Op    JSONPatchOperation `json:"op"`
	Path  string             `json:"path"`
	Value any                `json:"value"`
}

func PatchK8sObject(ctx context.Context, c client.Client, obj client.Object, patches ...JSONPatchItem) {
	GinkgoHelper()
	b, err := json.Marshal(patches)
	Expect(err).To(BeNil())
	Expect(c.Patch(ctx, obj, client.RawPatch(types.JSONPatchType, b))).To(Succeed())
}
