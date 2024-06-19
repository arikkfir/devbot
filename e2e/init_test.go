package e2e_test

import (
	"embed"
)

var (
	//go:embed all:repositories/*
	repositoriesFS embed.FS
)
