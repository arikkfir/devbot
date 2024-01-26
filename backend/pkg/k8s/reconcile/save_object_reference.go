package reconcile

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewSaveObjectReferenceAction[T client.Object](target *T) Action {
	return func(ctx context.Context, c client.Client, o client.Object) (*ctrl.Result, error) {
		*target = o.(T)
		return Continue()
	}
}
