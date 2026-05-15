package natvs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sov.fleet/sLatentLingua/02000-logic-libraries/webnf"
)

// JulesAgent is a Go wrapper for the Google Jules CLI.
type JulesAgent struct {
	BinaryPath string
	Codec      *webnf.WebnfCodec
}

func NewJulesAgent(binaryPath string) *JulesAgent {
	return &JulesAgent{
		BinaryPath: binaryPath,
		Codec:      &webnf.WebnfCodec{},
	}
}

func (j *JulesAgent) Negotiate(ctx context.Context, task Task) Result {
	slog.Info("Negotiating with Jules", "binary", j.BinaryPath)
	_, err := exec.LookPath(j.BinaryPath)
	if err != nil {
		return Result{Stage: Negotiation, Success: false, Error: err}
	}
	return Result{Stage: Negotiation, Success: true, Output: "Jules ready."}
}

func (j *JulesAgent) runJules(ctx context.Context, autoAccept bool, args ...string) ([]byte, error) {
	slog.Debug("Executing Jules command", "args", args)
	cmd := exec.CommandContext(ctx, j.BinaryPath, args...)

	// Force headless mode via env
	cmd.Env = append(os.Environ(), "TERM=dumb", "JULES_HEADLESS=true")

	if autoAccept {
		// Auto-Accept Mode: Automatically authorize plans
		cmd.Stdin = strings.NewReader("y\n")
	} else {
		// Two-Phase Mode: Connect to host Stdin for human approval
		cmd.Stdin = os.Stdin
	}

	return cmd.CombinedOutput()
}

func (j *JulesAgent) Assimilate(ctx context.Context, task Task) Result {
	slog.Info("Assimilating workspace context", "context", task.Context)
	// Use 'jules new' with context-init to establish the session
	out, err := j.runJules(ctx, true, "new", "--repo", strings.Join(task.Context, ","), "initial context ingestion")
	if err != nil {
		return Result{Stage: Assimilation, Success: false, Error: err, Output: string(out)}
	}
	return Result{Stage: Assimilation, Success: true, Output: string(out)}
}

func (j *JulesAgent) Transform(ctx context.Context, task *Task) Result {
	slog.Info("Transforming code via Jules", "objective", task.Objective, "remediation", task.Feedback != "")

	// 1. Serialize Task to webnf for the directive
	taskData, _ := j.Codec.Marshal(task)

	directive := `
CRITICAL: You MUST produce a file named 'CHRONICLE.webnf' in the root.
The file MUST conform to the :NATVS:Protocol:v1 grammar.

### SOVEREIGN TASK REQUEST (weBNF)
` + string(taskData)

	args := []string{"new", directive}
	if task.Parallel > 1 {
		args = append(args, "--parallel", fmt.Sprintf("%d", task.Parallel))
	}

	out, err := j.runJules(ctx, task.AutoAccept, args...)
	if err != nil {
		return Result{Stage: Transformation, Success: false, Error: err, Output: string(out)}
	}
	return Result{Stage: Transformation, Success: true, Output: string(out)}
}

func (j *JulesAgent) Verify(ctx context.Context, task Task) Result {
	slog.Info("Verifying transformation results")

	// Ensure we have the latest changes from the cloud session
	pullOut, err := j.runJules(ctx, true, "remote", "pull")
	if err != nil {
		return Result{Stage: Verification, Success: false, Error: err, Output: string(pullOut)}
	}

	gitRoot, _ := FindGitRoot()
	goBin := filepath.Join(gitRoot, "00FLOW", "sForge", "92000-external-toolchains", "go", "bin", "go.exe")

	slog.Info("Running validation tests", "go_bin", goBin)
	cmd := exec.CommandContext(ctx, goBin, "test", "./...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Stage: Verification, Success: false, Error: err, Output: string(out)}
	}
	return Result{Stage: Verification, Success: true, Output: string(out)}
}

func (j *JulesAgent) Synthesize(ctx context.Context, task Task) Result {
	slog.Info("Synthesizing final results")

	summaryPath := "CHRONICLE.webnf"
	summaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		slog.Warn("Final chronicle not found", "path", summaryPath)
		return Result{Stage: Synthesis, Success: false, Error: err, Output: "Chronicle missing."}
	}

	// 2. Validate and Unmarshal Chronicle
	var chronicle Chronicle
	if err := j.Codec.Unmarshal(summaryData, &chronicle); err != nil {
		slog.Error("Grammar violation in CHRONICLE.webnf", "error", err)
		return Result{Stage: Synthesis, Success: false, Error: err, Output: string(summaryData)}
	}

	slog.Info("Chronicle validated successfully", "agent", chronicle.Agent, "rationale", chronicle.Rationale)
	return Result{Stage: Synthesis, Success: true, Output: string(summaryData)}
}

func (j *JulesAgent) Cleanup(ctx context.Context, task Task) Result {
	gitRoot, err := FindGitRoot()
	if err != nil {
		return Result{Stage: Cleanup, Success: false, Error: err}
	}

	cachePath := filepath.Join(gitRoot, "00FLOW", "sForge", "C0880-repository-cache")
	slog.Info("Surgical cleaning of repository cache", "path", cachePath)

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{Stage: Cleanup, Success: true, Output: "Cache directory does not exist."}
		}
		slog.Warn("Failed to read cache directory", "error", err)
		return Result{Stage: Cleanup, Success: true} // Non-blocking
	}

	for _, entry := range entries {
		if entry.Name() == ".gitkeep" {
			continue
		}

		fullPath := filepath.Join(cachePath, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			slog.Warn("Failed to delete cache item", "path", fullPath, "error", err)
		}
	}

	return Result{Stage: Cleanup, Success: true, Output: "C0880 cache purged surgically."}
}
