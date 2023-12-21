package k8s

import (
	"context"
	"encoding/json"
	"github.com/secureworks/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const (
	BreakAnnPrefix = "devbot.kfirs.com/break-"
)

type BreakAnnotation struct {
	Break  bool         `json:"break"`
	Result *ctrl.Result `json:"result"`
}

func (ba *BreakAnnotation) MarshalJSON() ([]byte, error) {
	if !ba.Break {
		return []byte(`{"break":false}`), nil
	} else if ba.Result == nil {
		return []byte(`{"break":true}`), nil
	} else {
		return json.Marshal(struct {
			Break        bool   `json:"break"`
			Requeue      bool   `json:"requeue"`
			RequeueAfter string `json:"requeueAfter"`
		}{
			Break:        true,
			Requeue:      ba.Result.Requeue,
			RequeueAfter: ba.Result.RequeueAfter.String(),
		})
	}
}

func (ba *BreakAnnotation) UnmarshalJSON(b []byte) error {
	// local struct just for parsing
	a := &struct {
		Break        *bool   `json:"break,omitempty"`
		Requeue      *bool   `json:"requeue,omitempty"`
		RequeueAfter *string `json:"requeueAfter,omitempty"`
	}{}

	// parse JSON into local struct
	if err := json.Unmarshal(b, a); err != nil {
		return errors.New("failed parsing annotation: %+v", err)
	}

	// if "break" is not set, assume false if other properties are missing too; otherwise, assume true
	if a.Break == nil {
		if a.Requeue == nil && a.RequeueAfter == nil {
			ba.Break = false
			ba.Result = nil
			return nil
		}
		a.Break = &[]bool{true}[0]
	}

	// if "break" is false, ignore other properties
	if !*a.Break {
		ba.Break = false
		ba.Result = nil
		return nil
	}

	// if "break" is true, set other properties
	ba.Break = true
	ba.Result = &ctrl.Result{}
	if a.Requeue != nil {
		ba.Result.Requeue = *a.Requeue
	}
	if a.RequeueAfter != nil {
		if d, err := time.ParseDuration(*a.RequeueAfter); err != nil {
			return errors.New("failed parsing requeueAfter: %+v", err)
		} else {
			ba.Result.RequeueAfter = d
		}
	}
	return nil
}

func ShouldBreak(ctx context.Context, o client.Object, breakAnnName string) (bool, ctrl.Result) {
	var ba BreakAnnotation
	if o == nil {
		return false, ctrl.Result{}
	} else if o.GetAnnotations() == nil {
		return false, ctrl.Result{}
	} else if v, ok := o.GetAnnotations()[GetBreakAnnotationFullName(breakAnnName)]; !ok {
		return false, ctrl.Result{}
	} else if err := json.Unmarshal([]byte(v), &ba); err != nil {
		log.FromContext(ctx).Error(err, "failed to parse annotation", "annotation", breakAnnName, "value", v)
		return false, ctrl.Result{}
	} else {
		return true, *ba.Result
	}
}

func GetBreakAnnotationFullName(breakAnnName string) string {
	return BreakAnnPrefix + breakAnnName
}
