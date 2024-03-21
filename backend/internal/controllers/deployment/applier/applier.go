package applier

import (
	"bytes"
	"context"
	"fmt"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/rs/zerolog"
	"github.com/secureworks/errors"
	"os/exec"
	"path/filepath"
)

const (
	kubectlBinaryFilePath = "/usr/local/bin/kubectl"
)

// Applier is a struct that can apply a given resources manifest to a target cluster.
//
// TODO: extract applying into a separate worker pool, so we can limit the number of concurrent bakes
type Applier struct {
	kubectlBinaryPath string
}

func NewApplier(kubectlBinaryPath string) *Applier {
	return &Applier{kubectlBinaryPath: kubectlBinaryPath}
}

func NewDefaultApplier() *Applier {
	return NewApplier(kubectlBinaryFilePath)
}

func (b *Applier) Apply(ctx context.Context, manifestFile string) error {
	processCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd := exec.CommandContext(processCtx, b.kubectlBinaryPath,
		"apply",
		fmt.Sprintf("--filename=%s", manifestFile),
		fmt.Sprintf("--dry-run=%s", "server"),
		fmt.Sprintf("--server-side=%v", true),
	)
	cmd.Dir = filepath.Dir(manifestFile)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return errors.New("failed starting kubectl command: %w", err)
	}

	l := zerolog.Ctx(processCtx)
	stderrLogger := l.With().Str("process", "kubectl").Str("output", "stderr").Logger()
	stdoutLogger := l.With().Str("process", "kubectl").Str("output", "stdout").Logger()
	go b.createCommandLogSyncer(processCtx, cmd.Stderr.(*bytes.Buffer), &stderrLogger)()
	go b.createCommandLogSyncer(processCtx, cmd.Stdout.(*bytes.Buffer), &stdoutLogger)()

	if err := cmd.Wait(); err != nil {
		return errors.New("failed running kubectl: %w", err)
	}

	return nil
}

func (b *Applier) createCommandLogSyncer(ctx context.Context, buf *bytes.Buffer, l *zerolog.Logger) func() {
	return func() {
		linesChan := make(chan string, 1000)

		// Background goroutine to read from the command's stderr and send to the lines channel
		go func() {
			if err := stringsutil.AsyncReadLinesFromBuffer(ctx, buf, linesChan); err != nil {
				l.Err(err).Msg("Failed reading output")
			}
		}()

		// Background goroutine to read from the lines channel and write them to the logger
		go func() {
			for {
				if line, ok := <-linesChan; ok {
					l.Warn().Msg(line)
				} else {
					return
				}
			}
		}()

		<-ctx.Done()
	}
}
