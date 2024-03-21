package baker

import (
	"bytes"
	"context"
	stringsutil "github.com/arikkfir/devbot/backend/internal/util/strings"
	"github.com/rs/zerolog"
	"github.com/secureworks/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	kustomizeBinaryFilePath = "/usr/local/bin/kustomize"
	yqBinaryFilePath        = "/usr/local/bin/yq"
)

type processType string

const (
	kustomize processType = "kustomize"
	yq        processType = "yq"
)

// Baker is a struct that can generate a manifest from a kustomization file.
//
// TODO: extract baking into a separate worker pool, so we can limit the number of concurrent bakes
type Baker struct {
	kustomizeBinaryPath string
	yqBinaryPath        string
}

func NewBaker(kustomizeBinaryPath, yqBinaryPath string) *Baker {
	return &Baker{
		kustomizeBinaryPath: kustomizeBinaryPath,
		yqBinaryPath:        yqBinaryPath,
	}
}

func NewDefaultBaker() *Baker {
	return NewBaker(kustomizeBinaryFilePath, yqBinaryFilePath)
}

func (b *Baker) GenerateManifest(ctx context.Context, workDir string, appName, preferredBranch, actualBranch, sha string) (string, error) {
	processCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create target resources file
	resourcesFile, err := os.Create(workDir + "/.devbot.output.resources.yaml")
	if err != nil {
		return resourcesFile.Name(), errors.New("failed creating manifest file: %w", err)
	}
	defer resourcesFile.Close()

	// Create a pipe that connects stdout of the "kustomize build" command to the "kustomize fn" command
	pipeReader, pipeWriter := io.Pipe()

	// This command produces resources from the kustomization file and outputs them to stdout
	kustomizeCmd, kustomizeLogSyncer, err := b.createKustomizeCommand(processCtx, zerolog.Ctx(ctx), workDir, pipeWriter, preferredBranch, actualBranch)
	if err != nil {
		return resourcesFile.Name(), errors.New("failed creating kustomize command: %w", err)
	} else if err := kustomizeCmd.Start(); err != nil {
		return resourcesFile.Name(), errors.New("failed starting kustomize command: %w", err)
	}
	go kustomizeLogSyncer()

	// This command accepts resources via stdin, processes them via the bash function script, and outputs to stdout
	yqCmd, yqLogSyncer := b.createYQCommand(processCtx, zerolog.Ctx(ctx), appName, preferredBranch, actualBranch, sha, pipeReader, resourcesFile)
	if err := yqCmd.Start(); err != nil {
		return resourcesFile.Name(), errors.New("failed starting yq command: %w", err)
	}
	go yqLogSyncer()

	// Wait for kustomize command to finish
	if err := kustomizeCmd.Wait(); err != nil {
		return resourcesFile.Name(), errors.New("failed running kustomize: %w", err)
	} else if err := pipeWriter.Close(); err != nil {
		return resourcesFile.Name(), errors.New("failed closing pipe: %w", err)
	}

	// Wait for yq command to finish
	if err := yqCmd.Wait(); err != nil {
		return resourcesFile.Name(), errors.New("failed running yq: %w", err)
	}

	// success!
	return resourcesFile.Name(), nil
}

func (b *Baker) createKustomizeCommand(ctx context.Context, l *zerolog.Logger, workDir string, stdout io.Writer, preferredBranch, actualBranch string) (*exec.Cmd, func(), error) {
	devbotPaths := []string{
		filepath.Join(workDir, ".devbot", preferredBranch),
		filepath.Join(workDir, ".devbot", actualBranch),
		filepath.Join(workDir, ".devbot"),
	}
	for _, path := range devbotPaths {
		if _, err := os.Stat(path); err == nil {
			cmd := exec.CommandContext(ctx, b.kustomizeBinaryPath, "build")
			cmd.Dir = path
			cmd.Stderr = &bytes.Buffer{}
			cmd.Stdout = stdout
			ll := l.With().Str("process", string(kustomize)).Str("output", "stderr").Logger()
			return cmd, b.createCommandLogSyncer(ctx, cmd.Stderr.(*bytes.Buffer), &ll), nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, errors.New("failed inspecting path repository devbot path: %w", path, err)
		}
	}
	return nil, nil, errors.New("could not find devbot in any of: %+v", devbotPaths)
}

func (b *Baker) createYQCommand(ctx context.Context, l *zerolog.Logger, appName, preferredBranch, actualBranch, sha string, stdin io.Reader, stdout io.Writer) (*exec.Cmd, func()) {
	cmd := exec.CommandContext(ctx, b.yqBinaryPath, `(.. | select(tag == "!!str")) |= envsubst`)
	cmd.Env = append(os.Environ(),
		"APPLICATION="+stringsutil.Slugify(appName),
		"BRANCH="+stringsutil.Slugify(actualBranch),
		"COMMIT_SHA="+sha,
		"ENVIRONMENT="+stringsutil.Slugify(preferredBranch),
	)
	cmd.Stderr = &bytes.Buffer{}
	cmd.Stdin = stdin
	cmd.Stdout = stdout

	ll := l.With().Str("process", string(yq)).Str("output", "stderr").Logger()
	return cmd, b.createCommandLogSyncer(ctx, cmd.Stderr.(*bytes.Buffer), &ll)
}

func (b *Baker) createCommandLogSyncer(ctx context.Context, buf *bytes.Buffer, l *zerolog.Logger) func() {
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
