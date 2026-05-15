package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"sov.fleet/sAgents/81000-active-source/natvs"
	"sov.fleet/sLatentLingua/02000-logic-libraries/webnf"
	)
func init() {
	// Initialize Sovereign webnf logging
	handler := webnf.NewSlogWebnfHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	autoAccept := flag.Bool("auto-accept", false, "Bypass human approval for agent plans")
	parallel := flag.Int("parallel", 1, "Number of parallel sessions for the same task (1-5)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Usage: natvs-engine [--auto-accept] [--parallel N] [Objective] [Context_Paths...]")
		os.Exit(1)
	}

	objective := args[0]
	contextPaths := args[1:]

	// 1. Detect the Fleet Root
	gitRoot, err := natvs.FindGitRoot()
	if err != nil {
		slog.Error("Failed to find git root", "error", err)
		os.Exit(1)
	}
	slog.Info("Fleet Root Detected", "path", gitRoot)

	// 2. Initialize the Agent (Jules)
	// Dynamic binary path resolution
	julesBinary := filepath.Join(gitRoot, "00FLOW", "sForge", "91000-ext-artifact-cognition-intelligence", "jules", "jules.exe")
	agent := natvs.NewJulesAgent(julesBinary)

	// 3. Initialize the Orchestrator
	orchestrator := natvs.Orchestrator{
		Agent: agent,
	}

	// 4. Define the Task
	task := natvs.Task{
		ID:         fmt.Sprintf("TASK-%d", time.Now().Unix()),
		Objective:  objective,
		Context:    natvs.ResolvePaths(gitRoot, contextPaths),
		AutoAccept: *autoAccept,
		Parallel:   *parallel,
	}

	// 5. Run the NATVS Loop
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	results, err := orchestrator.Run(ctx, task)
	if err != nil {
		slog.Error("NATVS Loop failed", "task_id", task.ID, "error", err)
		os.Exit(1)
	}

	// 6. Retain the Summary in the Permanent Narrative Ledger (80600)
	finalResult := results[len(results)-1]
	if finalResult.Stage == natvs.Synthesis && finalResult.Output != "" {
	        // Use .webnf extension for the ledger
	        ledgerPath := filepath.Join(gitRoot, "000ALL", "sCognition", "80600-Agent-Narrative-Ledger", task.ID+".webnf")
	        err := os.WriteFile(ledgerPath, []byte(finalResult.Output), 0644)

		if err == nil {
			slog.Info("Transformation Summary committed to Narrative Ledger", "ledger_path", ledgerPath)
		} else {
			slog.Error("Failed to write to Narrative Ledger", "ledger_path", ledgerPath, "error", err)
		}
	}

	slog.Info("Task completed successfully", "task_id", task.ID)
}
