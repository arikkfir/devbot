package deployment

type Phase string

const (
	PhaseLabel   string = "devbot.kfirs.com/deployment-phase"
	PhaseUnknown Phase  = "unknown"
	PhaseClone   Phase  = "clone"
	PhaseBake    Phase  = "bake"
	PhaseApply   Phase  = "apply"
)
