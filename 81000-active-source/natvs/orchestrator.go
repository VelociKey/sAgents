package natvs

import (
	"context"
	"fmt"
	"log/slog"
)

// Orchestrator coordinates the NATVS loop for a specific agent.
type Orchestrator struct {
	Agent Agent
}

// Run executes the full NATVS protocol for a task, including the Fixer retry loop.
func (o *Orchestrator) Run(ctx context.Context, task Task) ([]Result, error) {
	slog.Info("Starting NATVS Loop", "task_id", task.ID, "auto_accept", task.AutoAccept)

	var results []Result

	// 1. Initial Purge
	res := o.Agent.Cleanup(ctx, task)
	results = append(results, res)
	if !res.Success {
		slog.Warn("Pre-task cleanup incomplete, proceeding", "error", res.Error)
	}

	// 2. Negotiation
	res = o.Agent.Negotiate(ctx, task)
	results = append(results, res)
	if !res.Success {
		return results, fmt.Errorf("negotiation failed: %w", res.Error)
	}

	// 3. Assimilation
	res = o.Agent.Assimilate(ctx, task)
	results = append(results, res)
	if !res.Success {
		return results, fmt.Errorf("assimilation failed: %w", res.Error)
	}

	// 4. The Fixer Loop (Transform -> Verify)
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		slog.Info("Starting Transformation Cycle", "attempt", attempt, "task_id", task.ID)

		// 4a. Transform
		res = o.Agent.Transform(ctx, &task)
		results = append(results, res)
		if !res.Success {
			return results, fmt.Errorf("transformation failed on attempt %d: %w", attempt, res.Error)
		}

		// 4b. Verify
		res = o.Agent.Verify(ctx, task)
		results = append(results, res)
		if res.Success {
			slog.Info("Verification successful", "attempt", attempt)
			break
		}

		slog.Warn("Verification failed, initiating remediation", "attempt", attempt, "error", res.Error)
		if attempt == maxRetries {
			return results, fmt.Errorf("verification failed after %d attempts: %w", maxRetries, res.Error)
		}

		// Pass verification output back as feedback for the next transformation
		task.Feedback = res.Output
	}

	// 5. Synthesis
	res = o.Agent.Synthesize(ctx, task)
	results = append(results, res)
	if !res.Success {
		return results, fmt.Errorf("synthesis failed: %w", res.Error)
	}

	// 6. Final Purge
	res = o.Agent.Cleanup(ctx, task)
	results = append(results, res)

	slog.Info("NATVS Loop Successful", "task_id", task.ID)
	return results, nil
}
