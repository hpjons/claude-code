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

	// mu guards session, which isn't safe for concurrent use. Wails runs
	// each bound-method call from the frontend on its own goroutine, so two
	// quick submissions can call Evaluate concurrently.
	mu      sync.Mutex
	session *engine.Session
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{session: engine.NewSession()}
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
	a.mu.Lock()
	defer a.mu.Unlock()

	result, err := a.session.Evaluate(expression)
	if err != nil {
		return EvaluationResult{
			Input:    expression,
			IsError:  true,
			ErrorMsg: err.Error(),
		}
	}

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
