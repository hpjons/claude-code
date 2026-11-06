// Grain-size classes on the Wentworth scale, stored as phi (larger phi = finer).

export interface GrainClass {
  key: string;
  name: string;
  short: string; // label on the log header
  phi: number; // a representative value inside the class
  group: "mud" | "sand" | "gravel";
}

/** From finest to coarsest; the log draws them left to right. */
export const GRAIN_CLASSES: GrainClass[] = [
  { key: "clay", name: "Clay", short: "clay", phi: 9, group: "mud" },
  { key: "silt", name: "Silt", short: "silt", phi: 6, group: "mud" },
  { key: "vf", name: "Very fine sand", short: "vf", phi: 3.5, group: "sand" },
  { key: "f", name: "Fine sand", short: "f", phi: 2.5, group: "sand" },
  { key: "m", name: "Medium sand", short: "m", phi: 1.5, group: "sand" },
  { key: "c", name: "Coarse sand", short: "c", phi: 0.5, group: "sand" },
  { key: "vc", name: "Very coarse sand", short: "vc", phi: -0.5, group: "sand" },
  { key: "gran", name: "Granule", short: "gran", phi: -1.5, group: "gravel" },
  { key: "pebb", name: "Pebble", short: "pebb", phi: -4, group: "gravel" },
  { key: "cobb", name: "Cobble", short: "cobb", phi: -7, group: "gravel" },
  { key: "boul", name: "Boulder", short: "boul", phi: -9, group: "gravel" },
];

/** Position of a phi value along the class axis: 0 = clay … GRAIN_CLASSES.length-1 = boulder. */
export function grainIndex(phi: number): number {
  const c = GRAIN_CLASSES;
  const first = c[0]!;
  const last = c[c.length - 1]!;
  if (phi >= first.phi) return 0;
  if (phi <= last.phi) return c.length - 1;
  for (let i = 1; i < c.length; i++) {
    const a = c[i - 1]!;
    const b = c[i]!;
    if (phi >= b.phi) return i - 1 + (a.phi - phi) / (a.phi - b.phi);
  }
  return c.length - 1;
}

/** The class a phi value belongs to (nearest representative value). */
export function grainClassOf(phi: number): GrainClass {
  return GRAIN_CLASSES[Math.round(grainIndex(phi))]!;
}
