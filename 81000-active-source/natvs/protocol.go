package natvs

import (
	"context"
)

// Stage represents one of the phases of the NATVS protocol.
type Stage string

const (
	Negotiation    Stage = "Negotiation"
	Assimilation   Stage = "Assimilation"
	Transformation Stage = "Transformation"
	Verification   Stage = "Verification"
	Synthesis      Stage = "Synthesis"
	Cleanup        Stage = "Cleanup"
)

// Task defines a modular objective for a Native Agent.
type Task struct {
	ID          string   `webnf:"task_id"`
	Type        string   `webnf:"type,optional"`
	Objective   string   `webnf:"objective"`
	Context     []string `webnf:"context"`
	Constraints []string `webnf:"constraints,optional"`
	Commentary  string   `webnf:"commentary,optional"`
	AutoAccept  bool     `webnf:"auto_accept"`
	Parallel    int      `webnf:"parallel,optional"`
	Feedback    string   `webnf:"feedback,optional"`
}

// Result captures the output of a NATVS stage.
type Result struct {
	Stage   Stage
	Success bool
	Output  string
	Error   error
}

type Observation struct {
	Type   string `webnf:"type"`
	Detail string `webnf:"detail"`
	Impact string `webnf:"impact"`
}

type Chronicle struct {
	TaskID       string        `webnf:"task_id"`
	Agent        string        `webnf:"agent"`
	Rationale    string        `webnf:"rationale"`
	Observations []Observation `webnf:"observations"`
}

// Agent defines the interface for a Native Agent (like Jules).
type Agent interface {
	Negotiate(ctx context.Context, task Task) Result
	Assimilate(ctx context.Context, task Task) Result
	Transform(ctx context.Context, task *Task) Result
	Verify(ctx context.Context, task Task) Result
	Synthesize(ctx context.Context, task Task) Result
	Cleanup(ctx context.Context, task Task) Result
}
