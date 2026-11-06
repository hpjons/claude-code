// Lays a log out on paper: one SVG per printed page, at the log's print scale, with margins.
// The same pages are printed and turned into PDF.
import { MARGIN_MM, PAPER_MM, drawKeyPage, drawLog, pageInfo } from "../render/log";
import type { Log } from "../types";

const MM = 96 / 25.4;

export interface Pages {
  svgs: string[]; // each the size of the paper, in mm
  paper: [number, number]; // width, height in mm
  shrink: number; // 1, or less when the columns are wider than the paper
}

/** Width and height of an SVG string's drawing, from its viewBox. */
export function svgSize(svg: string): [number, number] {
  const m = /viewBox="0 0 ([\d.]+) ([\d.]+)"/.exec(svg);
  return m ? [Number(m[1]), Number(m[2])] : [0, 0];
}

/** Puts a drawing on a page at (x, y) px, scaled by `s`. */
function place(svg: string, x: number, y: number, s: number): string {
  const [w, h] = svgSize(svg);
  return svg.replace(/^<svg ([^>]*?) width="[\d.]+" height="[\d.]+"/, `<svg $1 x="${x}" y="${y}" width="${(w * s).toFixed(2)}" height="${(h * s).toFixed(2)}"`);
}

function page(paper: [number, number], content: string): string {
  const [w, h] = paper;
  return (
    `<svg xmlns="http://www.w3.org/2000/svg" width="${w}mm" height="${h}mm" viewBox="0 0 ${(w * MM).toFixed(2)} ${(h * MM).toFixed(2)}">` +
    `<rect width="100%" height="100%" fill="#fff"/>${content}</svg>`
  );
}

export function layoutPages(log: Log, withKey: boolean): Pages {
  const paper = PAPER_MM[log.paper] ?? PAPER_MM.A4;
  const info = pageInfo(log);
  const room: [number, number] = [(paper[0] - 2 * MARGIN_MM) * MM, (paper[1] - 2 * MARGIN_MM) * MM];
  const m = MARGIN_MM * MM;
  const drawn = Array.from({ length: info.pages }, (_, p) => drawLog(log, { zoom: 1, page: p }));
  const shrink = Math.min(1, room[0] / svgSize(drawn[0]!)[0]);

  if (withKey) {
    // The key goes under the last page if it fits there, otherwise on a page of its own.
    const last = info.pages - 1;
    const withK = drawLog(log, { zoom: 1, page: last, key: true });
    if (svgSize(withK)[1] * shrink <= room[1] + 0.5) drawn[last] = withK;
    else if (svgSize(withK)[1] > svgSize(drawn[last]!)[1]) drawn.push(drawKeyPage(log, room[0] / shrink));
  }
  return { svgs: drawn.map((svg) => page(paper, place(svg, m, m, shrink))), paper, shrink };
}

/** The whole log as one drawing (for SVG and PNG), sized in mm so it prints at its scale. */
export function wholeLog(log: Log, withKey: boolean): string {
  const svg = drawLog(log, { zoom: 1, key: withKey });
  const [w, h] = svgSize(svg);
  return svg.replace(/^<svg ([^>]*?) width="[\d.]+" height="[\d.]+"/, `<svg $1 width="${(w / MM).toFixed(2)}mm" height="${(h / MM).toFixed(2)}mm"`);
}
