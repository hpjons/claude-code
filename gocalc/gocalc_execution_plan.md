# Project Specification: SpeedCrunch-Style Calculator Prototype (Wails + Go)

This document contains step-by-step instructions for **Claude Code** to bootstrap, build, and verify a high-precision, keyboard-driven math REPL prototype using **Wails v2** and **Go**.

---

## 1. Project Overview & Product Vision

- **Goal:** Build an interactive mathematical expression evaluator inspired by SpeedCrunch.
- **Form Factor:** Desktop window (via native Webview), dark theme by default, focused around an interactive single-line input with a scrollable history log above it.
- **Philosophy:**
  - Keyboard-first interaction (no on-screen numpad buttons).
  - Clean separation of concerns: The mathematical engine must reside in pure Go, testable independently via `go test`. The UI handles rendering, history display, and keystroke events.
  - Zero floating-point rounding quirks (e.g., `0.1 + 0.2` must equal `0.3`, not `0.30000000000000004`).

---

## 2. Technical Stack

- **Desktop Framework:** Wails v2 (or v3 preview if available/configured)
- **Backend:** Go (1.21+)
- **Frontend:** Vanilla HTML5, modern CSS (flexbox/grid, CSS variables), and vanilla ES6 JavaScript (keep it lightweight and dependency-free for Phase 1).
- **Math Engine:** Pure Go using `math/big` or decimal arithmetic (`github.com/cockroachdb/apd/v3` or `github.com/shopspring/decimal`) for exact precision calculations.

---

## 3. Phased Implementation Roadmap

### Phase 1: Core Math Engine (CLI / Go Only)
> **Rule for Claude Code:** Do not touch Wails or frontend code in this phase. Get the calculation engine verified via automated unit tests first.

1. **Parser & Evaluator Architecture:**
   - Implement an expression tokenizer and parser (Pratt parser or Shunting-yard algorithm).
   - Support basic arithmetic operators: `+`, `-`, `*`, `/`, `^` (power), `%` (modulo), and parentheses `(`, `)`.
   - Handle unary operators (e.g., `-5 + 3`).
   - Support standard operator precedence:
     1. Parentheses `()`
     2. Exponentiation `^`
     3. Unary minus `-`
     4. Multiplication `*`, Division `/`, Modulo `%`
     5. Addition `+`, Subtraction `-`
2. **Numeric Representation:**
   - Use high-precision arithmetic to avoid IEEE-754 floating-point artifacts.
   - Format results cleanly: trim redundant trailing decimal zeros (e.g., `5.00` becomes `5`).
3. **Automated Testing:**
   - Create comprehensive unit tests in `engine_test.go` checking:
     - `0.1 + 0.2` $\to$ `0.3`
     - `2 + 3 * 4` $\to$ `14`
     - `(2 + 3) * 4` $\to$ `20`
     - `2 ^ 3 ^ 2` $\to$ `512` (right-associative) or `64` (standard left)
     - `10 / 4` $\to$ `2.5`
     - Division by zero returns an informative error string.
     - Malformed syntax (e.g., `2 + * 3`, unbalanced `((2+3)`) returns friendly syntax errors.
   - Run `go test -v ./...` to ensure 100% pass rate.

---

### Phase 2: Wails Project Setup & Backend Bridge

1. **Initialize Project:**
   - Run `wails init -n speedcrunch-go -t vanilla` (or configure the existing directory).
   - Ensure `wails.json` and project structure follow idiomatic Wails layout.
2. **Expose Calculator Service:**
   - Create an application service struct (e.g., `App` or `CalculatorService`):
     ```go
     type EvaluationResult struct {
         Input     string `json:"input"`
         Result    string `json:"result"`
         IsError   bool   `json:"isError"`
         ErrorMsg  string `json:"errorMsg,omitempty"`
     }

     func (a *App) Evaluate(expression string) EvaluationResult
     ```
   - Wire the engine from Phase 1 directly into this method.
   - Run `wails build` or `wails dev` to trigger TypeScript/JavaScript binding generation under `frontend/wailsjs/`.

---

### Phase 3: Minimal REPL Frontend

1. **Layout Structure:**
   - **History Container (`#history`):** A vertically scrolling log occupying the top/middle of the window.
     - Each entry contains two lines: the original expression (dimmed, right/left aligned) and the result (prominent, copyable).
     - New evaluations scroll to the bottom automatically.
   - **Input Bar (`#input`):** A fixed bottom input line with an active prompt symbol (`> `).
2. **Keyboard Ergonomics:**
   - The input field must auto-focus on application launch and keep focus whenever the window is clicked.
   - **`Enter` Key:** Sends the input string to Go via `window.go.main.App.Evaluate(...)`.
   - **`Up` / `Down` Arrow Keys:** Cycle through past input history (command recall).
   - **`Esc` Key:** Clears the current input line.
3. **Visual Design:**
   - SpeedCrunch / terminal-inspired dark theme:
     - Background: Deep slate/dark gray (`#1e1e24` or `#121214`).
     - Input text: Clean monospace font (`JetBrains Mono`, `Fira Code`, or system monospace).
     - Normal result: High-contrast white/light gray (`#f0f0f0`).
     - Error result: Soft red/coral (`#ff6b6b`).
     - Selection/Highlight: Accent blue (`#3b82f6`).

---

## 4. Claude Code Operating Instructions

When executing this plan:

1. **Verification-Driven Development:**
   - Execute terminal commands via tool use (`go test ./...`, `wails build`).
   - If tests fail, inspect the error output, fix the root cause, and re-run tests before moving to the next task.
2. **Keep the Code Base Small and Inspectable:**
   - Do not pull in heavy frontend frameworks (no React, Vue, or Webpack configurations for this prototype).
   - Keep the frontend to a single `index.html`, `style.css`, and `main.js`.
3. **Extensibility for Future Phases:**
   - Design the parser token types so that variables (e.g., `ans`, `pi`, `x`), mathematical functions (`sin`, `sqrt`), and unit identifiers can be registered easily in later updates.