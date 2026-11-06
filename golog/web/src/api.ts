import type { Log, LogContent, Project, ProjectSummary, Statistics } from "./types";

/** How the columns of a CSV file are read: field key → column index. */
export interface CSVOptions {
  mapping: Record<string, number>;
  units: "m" | "cm";
  topFirst: boolean;
}

/** A SedLog file read by the server, not yet saved. */
export interface ImportResult {
  format: "slg" | "tem" | "csv" | "golog";
  name: string;
  content: LogContent;
  warnings: string[];
  csv?: CSVOptions & { headers: string[]; rows: number; sample: string[][]; fields: { key: string; label: string }[] };
}

/** An error reply from the server, with the message GoLog wrote for people. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly body: unknown,
  ) {
    super(message);
  }
}

async function call<T>(method: string, url: string, body?: unknown): Promise<T> {
  let res: Response;
  try {
    res = await fetch(url, {
      method,
      headers: body === undefined ? {} : { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(0, "GoLog isn't running. Start it again and reload this page.", null);
  }
  if (res.status === 204) return undefined as T;
  const data: unknown = await res.json().catch(() => null);
  if (!res.ok) {
    const msg = (data as { error?: string } | null)?.error ?? `${res.status} ${res.statusText}`;
    throw new ApiError(res.status, msg, data);
  }
  return data as T;
}

export const api = {
  health: () => call<{ app: string; version: string; canStop: boolean }>("GET", "/api/health"),
  /** Stops the GoLog program (for people who started it by double-clicking). */
  stop: () => call<{ status: string }>("POST", "/api/quit"),
  projects: () => call<ProjectSummary[]>("GET", "/api/projects"),
  createProject: (name: string) => call<Project>("POST", "/api/projects", { name }),
  project: (id: string) => call<Project>("GET", `/api/projects/${id}`),
  renameProject: (id: string, name: string) => call<Project>("PATCH", `/api/projects/${id}`, { name }),
  deleteProject: (id: string) => call<void>("DELETE", `/api/projects/${id}`),
  createLog: (projectId: string, name: string, content?: LogContent) => call<Log>("POST", `/api/projects/${projectId}/logs`, { name, content }),
  /** Reads a SedLog .slg, .tem or CSV file (base64) without saving it. */
  importFile: (data: string, csv?: CSVOptions) => call<ImportResult>("POST", "/api/import", { data, csv }),
  log: (id: string) => call<Log>("GET", `/api/logs/${id}`),
  saveLog: (log: Log) => call<Log>("PUT", `/api/logs/${log.id}`, log),
  deleteLog: (id: string) => call<void>("DELETE", `/api/logs/${id}`),
  /** Summaries and order tests for the log as it is on screen. */
  stats: (content: LogContent, shuffles?: number) => call<Statistics>("POST", "/api/stats", { content, shuffles }),
};
