import { describe, expect, it } from "vitest";
import { History } from "./history";

function setup() {
  const doc = { text: "a" };
  let t = 0;
  const hist = new History(
    () => JSON.stringify(doc),
    (s) => Object.assign(doc, JSON.parse(s)),
    () => t,
  );
  const edit = (text: string, source: unknown = null, dt = 100) => {
    t += dt;
    doc.text = text;
    hist.record(source);
  };
  return { doc, hist, edit };
}

describe("History", () => {
  it("undoes and redoes single changes", () => {
    const { doc, hist, edit } = setup();
    edit("b");
    edit("c");
    expect(hist.undo()).toBe(true);
    expect(doc.text).toBe("b");
    expect(hist.undo()).toBe(true);
    expect(doc.text).toBe("a");
    expect(hist.undo()).toBe(false);
    expect(hist.redo()).toBe(true);
    expect(doc.text).toBe("b");
  });
  it("merges quick changes from the same source", () => {
    const { doc, hist, edit } = setup();
    const input = {};
    edit("ab", input);
    edit("abc", input);
    edit("abcd", input);
    edit("abcd!", {}); // another source
    hist.undo();
    expect(doc.text).toBe("abcd");
    hist.undo();
    expect(doc.text).toBe("a");
  });
  it("starts a new step after a pause", () => {
    const { doc, hist, edit } = setup();
    const input = {};
    edit("ab", input);
    edit("abc", input, 5000);
    hist.undo();
    expect(doc.text).toBe("ab");
  });
  it("does not merge into a step that was undone and redone", () => {
    const { doc, hist, edit } = setup();
    const input = {};
    edit("ab", input);
    hist.undo();
    hist.redo();
    edit("abc", input);
    hist.undo();
    expect(doc.text).toBe("ab");
  });
  it("clears redo after a new change and ignores non-changes", () => {
    const { hist, edit } = setup();
    edit("b");
    hist.undo();
    expect(hist.canRedo).toBe(true);
    edit("a"); // same as the current state: nothing to record
    expect(hist.canRedo).toBe(true);
    edit("z");
    expect(hist.canRedo).toBe(false);
  });
});
