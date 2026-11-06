package main

import (
	"context"
	"sync"

	"gocalc/engine"
	"gocalc/i18n"
)

// App struct
type App struct {
	ctx context.Context

	// mu guards ans. Wails runs each bound-method call from the frontend on
	// its own goroutine, so two quick submissions can call Evaluate
	// concurrently.
	mu  sync.Mutex
	ans string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// EvaluationResult is the JSON-serializable outcome of evaluating one
// expression, returned to the frontend.
type EvaluationResult struct {
	Input    string `json:"input"`
	Result   string `json:"result"`
	IsError  bool   `json:"isError"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

// Evaluate parses and computes the given expression using the Phase 1
// math engine and reports the outcome to the frontend.
func (a *App) Evaluate(expression string) EvaluationResult {
	// Hold the lock across the whole evaluation, not just the reads and
	// writes of ans, so concurrent calls each see the previous call's
	// result rather than racing on the same old ans.
	a.mu.Lock()
	defer a.mu.Unlock()

	result, err := engine.Evaluate(expression, a.ans)
	if err != nil {
		return EvaluationResult{
			Input:    expression,
			IsError:  true,
			ErrorMsg: err.Error(),
		}
	}

	a.ans = result
	return EvaluationResult{
		Input:  expression,
		Result: result,
	}
}

// InputPlaceholder returns the translated placeholder text for the input
// bar, so the frontend never hardcodes English UI strings.
func (a *App) InputPlaceholder() string {
	return i18n.T("type an expression, press Enter…")
}
