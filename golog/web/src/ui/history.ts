/**
 * Undo and redo for a document that is edited in place. After each change the editor calls
 * `record` with the source of the change (e.g. the input being typed in); quick changes from
 * the same source become one step, so typing a word is undone in one go.
 */
export class History {
  private undoStack: string[] = [];
  private redoStack: string[] = [];
  private current: string;
  private lastSource: unknown = undefined;
  private lastTime = -Infinity;

  constructor(
    private snapshot: () => string,
    private restore: (state: string) => void,
    private now: () => number = () => Date.now(),
    private mergeMs = 1000,
    private limit = 200,
  ) {
    this.current = snapshot();
  }

  get canUndo(): boolean {
    return this.undoStack.length > 0;
  }

  get canRedo(): boolean {
    return this.redoStack.length > 0;
  }

  /** Notes that the document has changed. `source` groups quick changes into one step; null never merges. */
  record(source: unknown = null): void {
    const next = this.snapshot();
    if (next === this.current) return;
    const t = this.now();
    const merge = source !== null && source === this.lastSource && t - this.lastTime < this.mergeMs && this.undoStack.length > 0;
    if (!merge) {
      this.undoStack.push(this.current);
      if (this.undoStack.length > this.limit) this.undoStack.shift();
    }
    this.current = next;
    this.redoStack = [];
    this.lastSource = source;
    this.lastTime = t;
  }

  undo(): boolean {
    return this.step(this.undoStack, this.redoStack);
  }

  redo(): boolean {
    return this.step(this.redoStack, this.undoStack);
  }

  /** Forgets everything, e.g. when the document is replaced by a newer version from elsewhere. */
  reset(): void {
    this.undoStack = [];
    this.redoStack = [];
    this.current = this.snapshot();
    this.lastSource = undefined;
  }

  private step(from: string[], to: string[]): boolean {
    const state = from.pop();
    if (state === undefined) return false;
    to.push(this.current);
    this.current = state;
    this.lastSource = undefined; // the next edit starts a new step
    this.restore(state);
    return true;
  }
}
