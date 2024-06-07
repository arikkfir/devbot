package controller

type Phase string

const (
	PhaseLabel string = "devbot.kfirs.com/deployment-phase"
	PhaseClone Phase  = "clone"
	PhaseBake  Phase  = "bake"
	PhaseApply Phase  = "apply"
)
