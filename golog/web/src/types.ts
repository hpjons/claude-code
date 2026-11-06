// The same shapes as internal/model in the Go backend.

export type Contact = "unknown" | "sharp" | "erosional" | "gradational";

export interface Lithology {
  pattern: string;
  percent: number;
  colour?: string; // "#rrggbb"; the pattern's own colour if not set
}

export interface Grain {
  base?: number; // phi
  top?: number; // phi
}

export interface Bed {
  id: string;
  thickness: number; // metres
  contact: Contact; // at the base of the bed
  lithology: Lithology[];
  grain: Grain;
  inBed?: string; // structure drawn inside the grain-size profile
  symbols?: string[]; // structures, fossils and traces beside the bed
  bioturbation?: number; // bioturbation index 0–6
  palaeocurrents?: number[]; // degrees from north
  facies?: string;
  formation?: string; // a formation starts at this bed
  age?: string; // an age unit starts at this bed
  notes?: string;
}

export type ColumnKind =
  | "age" | "formation" | "scale" | "lithology" | "grain" | "symbols" | "bioturbation" | "palaeocurrent" | "facies" | "notes";

export interface Column {
  kind: ColumnKind;
  title: string;
  width: number; // px at 100 %
  visible: boolean;
}

export type Paper = "A4" | "A3" | "Letter";

/** The part of a log that is drawn and edited. */
export interface LogContent {
  units: "m" | "cm";
  scale: number; // print scale 1:scale
  paper: Paper;
  palaeocurrent: "arrows" | "numbers";
  fill: "colour" | "white"; // lithologies in colour, or black and white
  hideGrainClasses?: boolean; // leave out the grain-size names (clay … boulder) above the grain-size column
  hideGrainGroups?: boolean; // leave out the group names (mud, sand, gravel) above them
  hideGrainTitle?: boolean; // leave out the grain-size column's title
  columns: Column[]; // left to right
  beds: Bed[]; // beds[0] is the lowest
}

export interface Log extends LogContent {
  id: string;
  projectId: string;
  name: string;
  revision: number;
  created: string;
  updated: string;
}

export interface LogSummary {
  id: string;
  name: string;
  updated: string;
  bedCount: number;
}

export interface Project {
  id: string;
  name: string;
  created: string;
  updated: string;
  logs: LogSummary[];
}

export interface ProjectSummary {
  id: string;
  name: string;
  updated: string;
  logCount: number;
}

// Statistics, as calculated by the Go backend (internal/stats).

/** The part of the section that has one property (thickness in metres). */
export interface Share {
  key: string; // "" = not recorded
  beds: number;
  thickness: number;
  percent: number;
}

/** An observed order metric compared with the same metric in shuffled copies of the section. */
export interface OrderTest {
  observed: number;
  shuffles: number;
  mean: number;
  p: number; // share of shuffled copies scoring as high or higher
  evidence: string;
  bins: { start: number; width: number; counts: number[] };
}

export interface Statistics {
  summary: {
    beds: number;
    thickness: number;
    meanBed: number;
    medianBed: number;
    minBed: number;
    maxBed: number;
    sdBed: number;
    lithologies: Share[];
    facies: Share[];
    grain: Share[];
    trends: Share[];
    contacts: Share[];
  };
  facies?: { codes: string[]; beds: number[]; counts: number[][]; prob: number[][]; transitions: number; m: OrderTest };
  runs?: { thickening: number[]; thinning: number[]; r: OrderTest };
  notes?: string[];
}
