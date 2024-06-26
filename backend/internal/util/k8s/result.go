package k8s

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type Result struct {
	Error        error
	Requeue      bool
	RequeueAfter *time.Duration
}

func (r *Result) ToResultAndError() (ctrl.Result, error) {
	if r == nil {
		return ctrl.Result{}, nil
	} else if r.Error != nil {
		return ctrl.Result{}, r.Error
	} else if r.RequeueAfter != nil {
		return ctrl.Result{RequeueAfter: *r.RequeueAfter}, nil
	} else if r.Requeue {
		return ctrl.Result{Requeue: true}, nil
	} else {
		return ctrl.Result{}, nil
	}
}

func Continue() *Result {
	return nil
}

func DoNotRequeue() *Result {
	return &Result{}
}

func Requeue() *Result {
	return &Result{Requeue: true}
}

func RequeueAfter(interval time.Duration) *Result {
	return &Result{RequeueAfter: &interval}
}

func RequeueDueToError(err error) *Result {
	return &Result{Error: err}
}
