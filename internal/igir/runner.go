package igir

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
)

type Runner struct {
	cfg *config.Config
}

func NewRunner(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Run(ctx context.Context, args []string) error {
	bin := strings.TrimSpace(r.cfg.Igir.Binary)
	if bin == "" {
		bin = "igir"
	}

	if _, err := exec.LookPath(bin); err == nil {
		return runCommand(ctx, bin, args)
	}

	if !r.cfg.Igir.UseNpxFallback {
		return fmt.Errorf("igir binary not found in PATH and npx fallback disabled")
	}

	if _, err := exec.LookPath("npx"); err != nil {
		return fmt.Errorf("igir not found and npx unavailable: %w", err)
	}

	npxArgs := append([]string{"--yes", "igir@latest"}, args...)
	return runCommand(ctx, "npx", npxArgs)
}

func runCommand(ctx context.Context, bin string, args []string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run %s %v: %w", bin, args, err)
	}
	return nil
}
