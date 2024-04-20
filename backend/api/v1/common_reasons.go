package v1

const (
	ControllerNotAccessible    = "ControllerNotAccessible"
	ControllerNotFound         = "ControllerNotFound"
	ControllerReferenceMissing = "ControllerReferenceMissing"
	FinalizationFailed         = "FinalizationFailed"
	FinalizerRemovalFailed     = "FinalizerRemovalFailed"
	InProgress                 = "InProgress"
	InternalError              = "InternalError"
)

type ConditionsInverseState map[string]string
