package configuration

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
)

func Parse(cfg interface{}) {
	parser := flags.NewParser(cfg, flags.HelpFlag|flags.PassDoubleDash)
	if _, err := parser.Parse(); err != nil {
		fmt.Printf("ERROR: %s\n\n", err)
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
}
