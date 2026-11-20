"use client";

import * as React from "react";
import { toast } from "sonner";
import { ActivityIcon, ArrowUpRightIcon, BoxesIcon, GlobeIcon, NetworkIcon, PlayIcon, PlusIcon, RadarIcon, RotateCcwIcon, SaveIcon, SearchIcon, ShieldCheckIcon, TerminalSquareIcon, Trash2Icon, WorkflowIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { t } from "@/i18n";
import { api } from "@/lib/api";
import type { Agent, Tool, ToolRegistryRecord } from "@/lib/types";

// Traffic tools are host tools gated by the global traffic capture switch: bindable, but
// only usable when capture is on. Keep in sync with traffic.SeedToolMetas.
const TRAFFIC_TOOL_KEYS = new Set(["traffic_search", "traffic_get"]);

// paramRows flattens schema.properties into an ordered row list for editing.
// name/type/required are read-only (welded to the Go handler); description/default
// are editable. Non-scalar params (array/object) can't carry an editable default.
// parentKey is set for rows that live inside an array param's items.properties.
type ParamRow = {
  name: string;
  type: string;
  required: boolean;
  description: string;
  defaultStr: string; // "" = no default declared
  scalar: boolean;
  parentKey?: string; // set when this is a sub-field of an array param's items
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function toRows(schema: Record<string, any>): ParamRow[] {
  const props = (schema?.properties ?? {}) as Record<string, Record<string, unknown>>;
  const required = new Set<string>((schema?.required as string[]) ?? []);
  const rows: ParamRow[] = [];
  for (const [name, p] of Object.entries(props)) {
    const type = String(p?.type ?? "");
    const hasDefault = p != null && "default" in p && (p as Record<string, unknown>).default != null;
    rows.push({
      name, type,
      required: required.has(name),
      description: String(p?.description ?? ""),
      defaultStr: hasDefault ? String((p as Record<string, unknown>).default) : "",
      scalar: ["string", "integer", "number", "boolean"].includes(type),
    });
    // Expand items.properties for array params so sub-fields are editable.
    if (type === "array") {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const itemProps = (p as any)?.items?.properties as Record<string, Record<string, unknown>> | undefined;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const itemRequired = new Set<string>(((p as any)?.items?.required as string[]) ?? []);
      if (itemProps) {
        for (const [subName, subP] of Object.entries(itemProps)) {
          const subType = String((subP as Record<string, unknown>)?.type ?? "");
          const subHasDefault = subP != null && "default" in subP && subP.default != null;
          rows.push({
            name: subName, type: subType,
            required: itemRequired.has(subName),
            description: String(subP?.description ?? ""),
            defaultStr: subHasDefault ? String(subP.default) : "",
            scalar: ["string", "integer", "number", "boolean"].includes(subType),
            parentKey: name,
          });
        }
      }
    }
  }
  return rows;
}

// coerceDefault turns the edited default string back into a typed JSON value, or
// undefined to drop the "default" key entirely.
function coerceDefault(type: string, raw: string): unknown {
  const s = raw.trim();
  if (s === "") return undefined;
  if (type === "integer" || type === "number") {
    const n = Number(s);
    return Number.isNaN(n) ? undefined : n;
  }
  if (type === "boolean") return s === "true";
  return raw;
}

// applyRows writes edited rows back into a deep-copied schema (structure untouched).
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function applyRows(schema: Record<string, any>, rows: ParamRow[]): Record<string, any> {
  const next = structuredClone(schema ?? {});
  const props = (next.properties ?? {}) as Record<string, Record<string, unknown>>;
  for (const r of rows) {
    if (r.parentKey) {
      // Sub-field of an array param's items.properties
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const parent = props[r.parentKey] as any;
      const subP = parent?.items?.properties?.[r.name] as Record<string, unknown> | undefined;
      if (!subP) continue;
      subP.description = r.description;
      if (r.scalar) {
        const dv = coerceDefault(r.type, r.defaultStr);
        if (dv === undefined) delete subP.default;
        else subP.default = dv;
      }
    } else {
      const p = props[r.name];
      if (!p) continue;
      p.description = r.description;
      if (r.scalar) {
        const dv = coerceDefault(r.type, r.defaultStr);
        if (dv === undefined) delete p.default;
        else p.default = dv;
      }
    }
  }
  return next;
}

// ToolEditor is the full edit form for one tool, rendered inside the drawer.
function ToolEditor({
  tool,
  agents,
  captureOn,
  onSaved,
  onClose,
}: {
  tool: Tool;
  agents: Agent[];
  captureOn: boolean;
  onSaved: () => void;
  onClose: () => void;
}) {
  // traffic tools can't be bound/enabled until the global traffic capture switch is on.
  const trafficGated = TRAFFIC_TOOL_KEYS.has(tool.key) && !captureOn;
  const [description, setDescription] = React.useState(tool.description);
  const [bound, setBound] = React.useState<string[]>(tool.agents);
  const [enabled, setEnabled] = React.useState(tool.enabled);
  const [rows, setRows] = React.useState<ParamRow[]>(() => toRows(tool.schema));
  const [saving, setSaving] = React.useState(false);

  const setRow = (i: number, patch: Partial<ParamRow>) =>
    setRows((rs) => rs.map((r, j) => (j === i ? { ...r, ...patch } : r)));
  const toggleAgent = (k: string) =>
    setBound((b) => (b.includes(k) ? b.filter((x) => x !== k) : [...b, k]));

  async function save() {
    setSaving(true);
    try {
      await api.saveTool(tool.key, {
        description,
        schema: applyRows(tool.schema, rows),
        agents: bound,
        enabled,
      });
      toast.success(t('Tool "{key}" saved', { key: tool.key }));
      onSaved();
      onClose();
    } catch (e) {
      toast.error(t("Save failed: {error}", { error: (e as Error).message }));
    } finally {
      setSaving(false);
    }
  }
  async function reset() {
    try {
      await api.resetTool(tool.key);
      toast.success(t('Restored "{key}" to the code default', { key: tool.key }));
      onSaved();
      onClose();
    } catch (e) {
      toast.error(t("Restore failed: {error}", { error: (e as Error).message }));
    }
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-4">
        {trafficGated && (
          <div className="border-amber-500/40 bg-amber-500/10 text-muted-foreground rounded-md border px-3 py-2 text-xs">
            {(() => {
              const [before, after] = t("This tool depends on {capture}.").split("{capture}");
              return (
                <>
                  {before}
                  <b>{t("traffic capture")}</b>
                  {after}
                </>
              );
            })()}{" "}
            {t('Enable traffic capture under "System settings" first before binding it to agents and enabling it.')}
          </div>
        )}
        {/* binding + switch */}
        <div className="flex flex-wrap items-center gap-x-6 gap-y-3">
          <div className="grid gap-1.5">
            <Label className="text-muted-foreground text-xs">{t("Bind agents (controls which agents get this tool)")}</Label>
            <div className="flex flex-wrap gap-3">
              {agents.map((ag) => (
                <label key={ag.key} className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={bound.includes(ag.key)}
                    disabled={trafficGated}
                    onCheckedChange={() => toggleAgent(ag.key)}
                  />
                  {ag.name}
                  <span className="text-muted-foreground font-mono text-xs">{ag.key}</span>
                </label>
              ))}
              {agents.length === 0 && <span className="text-muted-foreground text-xs">{t("(no agents)")}</span>}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Switch checked={enabled} onCheckedChange={setEnabled} id={`en-${tool.key}`} />
            <Label htmlFor={`en-${tool.key}`} className="text-sm">
              {t("Enable")}
            </Label>
          </div>
        </div>

        {/* description */}
        <div className="grid gap-1.5">
          <Label className="text-muted-foreground text-xs">{t("Tool description (sent to the model)")}</Label>
          <Textarea
            className="font-mono text-xs"
            rows={6}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>

        {/* params */}
        <div className="grid gap-2">
          <Label className="text-muted-foreground text-xs">
            {t("Parameters (name / type / required are read-only; description and default are editable)")}
          </Label>
          {rows.length === 0 && <span className="text-muted-foreground text-xs">{t("(no parameters)")}</span>}
          {rows.map((r, i) => (
            <div
              key={(r.parentKey ?? "") + "." + r.name}
              className={r.parentKey
                ? "border-l-2 border-muted ml-3 pl-3 grid gap-2 py-2"
                : "grid gap-2 rounded-md border p-3"
              }
            >
              <div className="flex flex-wrap items-center gap-2">
                {r.parentKey && <span className="text-muted-foreground font-mono text-[10px]">↳</span>}
                <span className="font-mono text-sm">{r.name}</span>
                <Badge variant="secondary" className="px-1.5 py-0 text-[10px]">
                  {r.type || "?"}
                </Badge>
                {r.required && (
                  <Badge variant="outline" className="px-1.5 py-0 text-[10px]">
                    {t("Required")}
                  </Badge>
                )}
                {r.parentKey && (
                  <span className="text-muted-foreground text-[10px]">{t("items sub-field")}</span>
                )}
              </div>
              <div className="grid gap-2">
                <div className="grid gap-1">
                  <Label className="text-muted-foreground text-[11px]">{t("Description")}</Label>
                  <Input
                    className="text-xs"
                    value={r.description}
                    onChange={(e) => setRow(i, { description: e.target.value })}
                  />
                </div>
                <div className="grid gap-1">
                  <Label className="text-muted-foreground text-[11px]">{t("Default")}</Label>
                  <Input
                    className="text-xs"
                    placeholder={r.scalar ? t("(empty = no default)") : t("scalars only")}
                    disabled={!r.scalar}
                    value={r.defaultStr}
                    onChange={(e) => setRow(i, { defaultStr: e.target.value })}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <Separator className="mt-4" />
      <div className="flex flex-wrap gap-2 p-4">
        <Button size="sm" onClick={save} disabled={saving}>
          <SaveIcon /> {t("Save")}
        </Button>
        <Button size="sm" variant="outline" onClick={reset}>
          <RotateCcwIcon /> {t("Restore defaults")}
        </Button>
      </div>
    </div>
  );
}

// ToolGridCard presents each registered tool as a marketplace-style visual tile.
function ToolGridCard({ tool, onClick }: { tool: Tool; onClick: () => void }) {
  const key = tool.key.toLowerCase();
  const Icon = /http|web|url|browser|request/.test(key)
    ? GlobeIcon
    : /scan|nmap|nuclei|discover|recon/.test(key)
      ? RadarIcon
      : /traffic|proxy|network|packet/.test(key)
        ? NetworkIcon
        : /agent|workflow|plan|task/.test(key)
          ? WorkflowIcon
          : /shell|command|script|terminal/.test(key)
            ? TerminalSquareIcon
            : /evidence|finding|security|guard/.test(key)
              ? ShieldCheckIcon
              : /asset|tool|registry|list/.test(key)
                ? BoxesIcon
                : ActivityIcon;
  const specialty = tool.description?.trim() || "A connected capability in the ARTEX tool ecosystem.";
  return (
    <button
      type="button"
      onClick={onClick}
      className="group flex min-h-[260px] flex-col overflow-hidden rounded-xl border border-border/70 bg-card text-left shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-sky-500/50 hover:shadow-xl hover:shadow-sky-950/10 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <div className="relative flex h-32 items-center justify-between overflow-hidden border-b border-white/10 bg-[radial-gradient(ellipse_at_75%_20%,rgba(14,165,233,0.25),transparent_48%),linear-gradient(135deg,#07111f,#101d32_55%,#111827)] px-5">
        <div className="absolute -right-5 -top-12 size-44 rounded-full border border-sky-300/15" />
        <div className="absolute -right-1 -top-6 size-32 rounded-full border border-sky-300/10" />
        <div className="absolute bottom-0 left-0 h-px w-full bg-gradient-to-r from-transparent via-sky-400/60 to-transparent" />
        <div className="relative grid size-16 place-items-center rounded-2xl border border-sky-300/25 bg-sky-400/10 text-sky-200 shadow-lg shadow-sky-950/30">
          <Icon className="size-8" strokeWidth={1.5} />
        </div>
        <div className="relative flex flex-col items-end gap-2">
          <span className="rounded-full border border-white/15 bg-black/20 px-2.5 py-1 text-[10px] font-medium uppercase tracking-[0.16em] text-sky-100/80">
            {tool.system ? "Native capability" : "Extension"}
          </span>
          <span className="font-mono text-[10px] text-slate-400">ARTEX / TOOL</span>
        </div>
      </div>
      <div className="flex flex-1 flex-col gap-3 p-4">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <h3 className="truncate text-sm font-semibold tracking-tight group-hover:text-sky-500">{tool.key}</h3>
            <p className="mt-1 line-clamp-2 min-h-9 text-xs leading-relaxed text-muted-foreground">{specialty}</p>
          </div>
          <ArrowUpRightIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-sky-500" />
        </div>
        <div className="mt-auto flex flex-wrap items-center gap-1.5">
          {tool.agents.slice(0, 3).map((agent) => (
            <Badge key={agent} variant="outline" className="px-2 py-0.5 text-[10px]">{agent}</Badge>
          ))}
          {tool.agents.length > 3 && <Badge variant="outline" className="px-2 py-0.5 text-[10px]">+{tool.agents.length - 3}</Badge>}
          {tool.agents.length === 0 && <span className="text-[10px] text-muted-foreground">No agent assigned</span>}
        </div>
        <div className="flex items-center justify-between border-t pt-3">
          <div className="flex items-center gap-2">
            <span className={tool.enabled ? "size-1.5 rounded-full bg-emerald-500" : "size-1.5 rounded-full bg-muted-foreground"} />
            <span className="text-[10px] text-muted-foreground">{tool.enabled ? "Enabled" : "Disabled"}</span>
            {tool.deferred && <Badge variant="outline" className="px-1.5 py-0 text-[9px]">Deferred</Badge>}
          </div>
          <span className="text-[10px] tabular-nums text-muted-foreground">{tool.calls ?? 0} calls</span>
        </div>
      </div>
    </button>
  );
}

function RegistryGridCard({ entry }: { entry: ToolRegistryRecord }) {
  const m = entry.manifest;
  const source = m.canonical_repository || m.sources?.[0]?.url;
  return (
    <article className="group flex min-h-[290px] flex-col overflow-hidden rounded-xl border border-border/70 bg-card shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-sky-500/50 hover:shadow-xl hover:shadow-sky-950/10">
      <div className="relative flex h-32 items-center justify-between overflow-hidden border-b border-white/10 bg-[radial-gradient(ellipse_at_75%_20%,rgba(14,165,233,0.25),transparent_48%),linear-gradient(135deg,#07111f,#101d32_55%,#111827)] px-5">
        <div className="absolute -right-5 -top-12 size-44 rounded-full border border-sky-300/15" />
        <div className="absolute -right-1 -top-6 size-32 rounded-full border border-sky-300/10" />
        <div className="relative grid size-16 place-items-center rounded-2xl border border-sky-300/25 bg-sky-400/10 text-sky-200">
          <GlobeIcon className="size-8" strokeWidth={1.5} />
        </div>
        <div className="relative flex flex-col items-end gap-2">
          <span className="rounded-full border border-white/15 bg-black/20 px-2.5 py-1 text-[10px] font-medium uppercase tracking-[0.16em] text-sky-100/80">External registry</span>
          <span className="font-mono text-[10px] text-slate-400">{m.execution_kind || "external"}</span>
        </div>
      </div>
      <div className="flex flex-1 flex-col gap-3 p-4">
        <div>
          <h3 className="text-sm font-semibold tracking-tight">{entry.name}</h3>
          <p className="mt-1 line-clamp-3 min-h-12 text-xs leading-relaxed text-muted-foreground">{m.description || "Registered external tool. Capability details are awaiting verification."}</p>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {(m.capabilities ?? []).slice(0, 4).map((capability) => (
            <Badge key={capability} variant="outline" className="px-2 py-0.5 text-[10px]">{capability.replaceAll("_", " ")}</Badge>
          ))}
        </div>
        <div className="mt-auto flex flex-wrap gap-1.5 border-t pt-3">
          <Badge variant="secondary" className="text-[10px]">{m.lifecycle || "candidate"}</Badge>
          <Badge variant="outline" className="text-[10px]">Trust: {m.trust || "unknown"}</Badge>
          <Badge variant="outline" className="text-[10px]">{m.availability || "unknown"}</Badge>
          {m.approval_required && <Badge variant="outline" className="text-[10px]">Approval required</Badge>}
        </div>
        <div className="flex items-center justify-between gap-2 text-[10px] text-muted-foreground">
          <span className="truncate">{[m.license, m.version].filter(Boolean).join(" · ") || "License/version not verified"}</span>
          {source ? <a href={source} target="_blank" rel="noreferrer" className="shrink-0 text-sky-500 hover:underline">Source ↗</a> : <span>Source pending</span>}
        </div>
      </div>
    </article>
  );
}

export default function ToolsPage() {
  const [tools, setTools] = React.useState<Tool[]>([]);
  const [registryEntries, setRegistryEntries] = React.useState<ToolRegistryRecord[]>([]);
  const [agents, setAgents] = React.useState<Agent[]>([]);
  const [captureOn, setCaptureOn] = React.useState(false);
  const [selectedKey, setSelectedKey] = React.useState<string | null>(null);
  const [customEdit, setCustomEdit] = React.useState<Tool | "new" | null>(null);

  const reload = React.useCallback(() => {
    api.tools().then(setTools).catch(() => setTools([]));
    api.toolRegistry().then(setRegistryEntries).catch(() => setRegistryEntries([]));
  }, []);
  React.useEffect(() => {
    reload();
    api.agents().then(setAgents).catch(() => {});
    api.settings().then((s) => setCaptureOn(!!s.traffic_capture)).catch(() => {});
  }, [reload]);

  const [query, setQuery] = React.useState("");

  const selected = tools.find((t) => t.key === selectedKey) ?? null;
  const matchTool = React.useCallback((t: Tool) => {
    const q = query.trim().toLowerCase();
    if (!q) return true;
    return (
      t.key.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q) ||
      t.agents.some((a) => a.toLowerCase().includes(q))
    );
  }, [query]);
  const systemTools = tools.filter((t) => t.system && matchTool(t));
  const customTools = tools.filter((t) => !t.system && matchTool(t));
  const allSystemCount = tools.filter((t) => t.system).length;
  const allCustomCount = tools.filter((t) => !t.system).length;
  // clicking a built-in tool opens the binding drawer; a custom tool opens its full editor.
  const openTool = (t: Tool) => (t.system ? setSelectedKey(t.key) : setCustomEdit(t));

  return (
    <div className="flex flex-1 flex-col gap-4 md:gap-6">
      <div className="relative overflow-hidden rounded-2xl border border-sky-500/20 bg-[radial-gradient(ellipse_at_85%_0%,rgba(14,165,233,0.18),transparent_45%),linear-gradient(120deg,#07111f,#111c2d_65%,#101827)] p-6 text-white shadow-lg">
        <div className="absolute -right-10 -top-24 size-72 rounded-full border border-sky-300/10" />
        <div className="absolute -right-1 -top-16 size-52 rounded-full border border-sky-300/10" />
        <div className="relative flex flex-wrap items-end justify-between gap-5">
          <div className="max-w-2xl">
            <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-sky-300/20 bg-sky-300/10 px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-sky-200">
              <BoxesIcon className="size-3.5" /> Tool intelligence library · draft
            </div>
            <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Discover your security toolkit.</h1>
            <p className="mt-2 max-w-xl text-sm leading-relaxed text-slate-300">Explore capabilities, compare specialties and assemble the right toolset. Documentation is here when you need it — discovery comes first.</p>
          </div>
          <div className="relative w-full sm:w-72">
            <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
            <Input
              className="h-10 border-white/15 bg-white/10 pl-9 text-sm text-white placeholder:text-slate-400 focus-visible:ring-sky-400"
              placeholder="Search tools and specialties…"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>
        </div>
        <div className="relative mt-6 flex flex-wrap gap-x-8 gap-y-3 border-t border-white/10 pt-4 text-xs text-slate-300">
          <span><b className="mr-1.5 text-white">{tools.length}</b> registered capabilities</span>
          <span><b className="mr-1.5 text-white">{tools.filter((tool) => tool.enabled).length}</b> enabled</span>
          <span><b className="mr-1.5 text-white">{tools.filter((tool) => tool.system).length}</b> native tools</span>
          <span><b className="mr-1.5 text-white">{tools.filter((tool) => !tool.system).length}</b> extensions</span>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold tracking-tight">Browse the ecosystem</h2>
          <p className="text-muted-foreground text-xs">Select a card to inspect configuration and agent bindings.</p>
        </div>
        <div className="text-muted-foreground flex items-center gap-1.5 text-[11px]"><ActivityIcon className="size-3.5" /> Live catalogue data</div>
      </div>

      <Tabs defaultValue="system">
        <TabsList>
          <TabsTrigger value="system">{t("System tools")}</TabsTrigger>
          <TabsTrigger value="custom">{t("Custom tools")}</TabsTrigger>
          <TabsTrigger value="ecosystem">{t("External ecosystem")} ({registryEntries.length})</TabsTrigger>
        </TabsList>

        <TabsContent value="system">
          <Card>
            <CardHeader>
              <CardTitle>{t("System tools")}</CardTitle>
              <CardDescription>
                {query.trim()
                  ? t(
                      "{count} / {total} matching; click a card to edit the description, parameter defaults and agent bindings",
                      { count: systemTools.length, total: allSystemCount },
                    )
                  : t("{total} total; click a card to edit the description, parameter defaults and agent bindings", {
                      total: allSystemCount,
                    })}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {systemTools.length === 0 ? (
                <p className="text-muted-foreground py-6 text-center text-sm">
                  {query.trim() ? t("No matching system tools") : t("(no system tools yet, waiting for the backend seed)")}
                </p>
              ) : (
                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                  {systemTools.map((t) => (
                    <ToolGridCard key={t.key} tool={t} onClick={() => openTool(t)} />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="custom">
          <Card>
            <CardHeader>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <CardTitle>{t("Custom tools")}</CardTitle>
                  <CardDescription>
                    {query.trim()
                      ? t("shell/command/script/http, {count} / {total} matching; click a card to edit", {
                          count: customTools.length,
                          total: allCustomCount,
                        })
                      : t(
                          "shell (bash declaration) / command (command template) / script (Python) / http (API), {total} total; click a card to edit",
                          { total: allCustomCount },
                        )}
                  </CardDescription>
                </div>
                <Button size="sm" onClick={() => setCustomEdit("new")}>
                  <PlusIcon /> {t("Create custom tool")}
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {customTools.length === 0 ? (
                <p className="text-muted-foreground py-6 text-center text-sm">
                  {query.trim() ? t("No matching custom tools") : t("(no custom tools yet; click 'Create custom tool' in the top right)")}
                </p>
              ) : (
                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                  {customTools.map((t) => (
                    <ToolGridCard key={t.key} tool={t} onClick={() => openTool(t)} />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="ecosystem">
          <Card>
            <CardHeader>
              <CardTitle>{t("External tool ecosystem")}</CardTitle>
              <CardDescription>
                {t("Curated external tools from the unified registry. Registration is descriptive only; it does not install, enable or execute a tool.")}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {registryEntries.filter((entry) => {
                const q = query.trim().toLowerCase();
                return !q || entry.name.toLowerCase().includes(q) ||
                  entry.manifest.description?.toLowerCase().includes(q) ||
                  (entry.manifest.capabilities ?? []).some((capability) => capability.toLowerCase().includes(q));
              }).length === 0 ? (
                <p className="text-muted-foreground py-6 text-center text-sm">{t("No registered external tools match this search.")}</p>
              ) : (
                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                  {registryEntries.filter((entry) => {
                    const q = query.trim().toLowerCase();
                    return !q || entry.name.toLowerCase().includes(q) ||
                      entry.manifest.description?.toLowerCase().includes(q) ||
                      (entry.manifest.capabilities ?? []).some((capability) => capability.toLowerCase().includes(q));
                  }).map((entry) => <RegistryGridCard key={entry.id} entry={entry} />)}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Sheet open={!!selected} onOpenChange={(o) => !o && setSelectedKey(null)}>
        <SheetContent
          side="right"
          className="gap-0 p-0 data-[side=right]:w-[30vw] data-[side=right]:sm:max-w-[30vw]"
        >
          {selected && (
            <>
              <SheetHeader className="px-4">
                <SheetTitle className="font-mono">{selected.key}</SheetTitle>
                <SheetDescription>{t("Edit the description, parameter defaults and agent bindings")}</SheetDescription>
              </SheetHeader>
              <ToolEditor
                key={selected.key}
                tool={selected}
                agents={agents}
                captureOn={captureOn}
                onSaved={reload}
                onClose={() => setSelectedKey(null)}
              />
            </>
          )}
        </SheetContent>
      </Sheet>

      <CustomToolDialog
        edit={customEdit}
        agents={agents}
        onClose={() => setCustomEdit(null)}
        onSaved={() => {
          setCustomEdit(null);
          reload();
        }}
      />
    </div>
  );
}

// ---- custom tool editor ----

type ExecState = {
  command: string;
  code: string;
  method: string;
  url: string;
  headers: string;
  body: string;
  timeout_ms: string;
  proxy: string;
  use_recording_proxy: boolean;
};

function CustomToolDialog({
  edit,
  agents,
  onClose,
  onSaved,
}: {
  edit: Tool | "new" | null;
  agents: Agent[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const isNew = edit === "new";
  const tool = edit && edit !== "new" ? edit : null;
  const [key, setKey] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [kind, setKind] = React.useState<"shell" | "command" | "script" | "http">("shell");
  const [ex, setEx] = React.useState<ExecState>({
    command: "", code: "", method: "GET", url: "", headers: "", body: "", timeout_ms: "", proxy: "", use_recording_proxy: false,
  });
  const [schemaText, setSchemaText] = React.useState("");
  const [bound, setBound] = React.useState<string[]>([]);
  const [deferred, setDeferred] = React.useState(false);
  const [enabled, setEnabled] = React.useState(true);
  const [saving, setSaving] = React.useState(false);
  const [paramsText, setParamsText] = React.useState("");
  const [testing, setTesting] = React.useState(false);
  const [testResult, setTestResult] = React.useState<{ output: string; is_error: boolean } | null>(null);

  // (re)load form state when opening.
  React.useEffect(() => {
    if (!edit) return;
    setParamsText(""); setTestResult(null);
    if (edit === "new") {
      setKey(""); setDescription(""); setKind("shell");
      setEx({ command: "", code: "", method: "GET", url: "", headers: "", body: "", timeout_ms: "", proxy: "", use_recording_proxy: false });
      setSchemaText(""); setBound([]); setDeferred(false); setEnabled(true);
      return;
    }
    const t = edit;
    const e = (t.exec ?? {}) as Record<string, unknown>;
    setKey(t.key);
    setDescription(t.description);
    setKind((t.kind as "shell" | "command" | "script" | "http") ?? "shell");
    setEx({
      command: String(e.command ?? ""),
      code: String(e.code ?? ""),
      method: String(e.method ?? "GET"),
      url: String(e.url ?? ""),
      headers: e.headers ? JSON.stringify(e.headers, null, 2) : "",
      body: String(e.body ?? ""),
      timeout_ms: e.timeout_ms ? String(e.timeout_ms) : "",
      proxy: String(e.proxy ?? ""),
      use_recording_proxy: !!e.use_recording_proxy,
    });
    setSchemaText(t.schema && Object.keys(t.schema).length ? JSON.stringify(t.schema, null, 2) : "");
    setBound(t.agents ?? []);
    setDeferred(!!t.deferred);
    setEnabled(t.enabled);
  }, [edit]);

  const toggleAgent = (k: string) => setBound((b) => (b.includes(k) ? b.filter((x) => x !== k) : [...b, k]));

  function buildExec(): Record<string, unknown> {
    if (kind === "shell") return {};
    const t = ex.timeout_ms ? Number(ex.timeout_ms) : undefined;
    if (kind === "command") return { command: ex.command, timeout_ms: t };
    if (kind === "script") return { code: ex.code, timeout_ms: t };
    let headers: Record<string, string> = {};
    if (ex.headers.trim()) {
      try { headers = JSON.parse(ex.headers); } catch { /* validated on save */ }
    }
    return { method: ex.method, url: ex.url, headers, body: ex.body, timeout_ms: t, proxy: ex.proxy, use_recording_proxy: ex.use_recording_proxy };
  }

  async function save() {
    if (isNew && !/^[a-z][a-z0-9_]*$/.test(key.trim())) {
      toast.error(t("key must start with a lowercase letter and contain only lowercase letters/digits/underscores"));
      return;
    }
    let schema: Record<string, unknown> = {};
    if (schemaText.trim()) {
      try { schema = JSON.parse(schemaText); } catch { toast.error(t("Invalid parameter JSON Schema")); return; }
    }
    if (kind === "http" && ex.headers.trim()) {
      try { JSON.parse(ex.headers); } catch { toast.error(t("Invalid headers JSON")); return; }
    }
    if (kind === "http") {
      const props = (schema as { properties?: Record<string, unknown> }).properties;
      if (!props || Object.keys(props).length === 0) {
        toast.error(t("http tools must provide a parameter JSON Schema (with properties; cannot be empty)"));
        return;
      }
    }
    setSaving(true);
    const payload = { description, schema: kind === "shell" ? {} : schema, agents: bound, enabled, kind, exec: buildExec(), deferred: kind === "shell" ? false : deferred };
    try {
      if (isNew) await api.createCustomTool({ key: key.trim(), ...payload });
      else await api.updateCustomTool(tool!.key, payload);
      toast.success(isNew ? t("Custom tool created") : t("Saved"));
      onSaved();
    } catch (e) {
      toast.error(t("Save failed: {error}", { error: (e as Error).message }));
    } finally {
      setSaving(false);
    }
  }
  async function del() {
    if (!tool) return;
    try {
      await api.deleteCustomTool(tool.key);
      toast.success(t("Deleted"));
      onSaved();
    } catch (e) {
      toast.error(t("Delete failed: {error}", { error: (e as Error).message }));
    }
  }
  // runTest dry-runs the CURRENT form (unsaved) with the sample params, so a
  // template/script/request can be debugged before saving.
  async function runTest() {
    let params: Record<string, unknown> = {};
    if (paramsText.trim()) {
      try { params = JSON.parse(paramsText); } catch { toast.error(t("Invalid test parameter JSON")); return; }
    }
    if (kind === "http" && ex.headers.trim()) {
      try { JSON.parse(ex.headers); } catch { toast.error(t("Invalid headers JSON")); return; }
    }
    setTesting(true);
    setTestResult(null);
    try {
      // key: a saved tool's stored secrets replace the masked placeholders.
      const r = await api.testCustomTool({ key: tool?.key, kind, exec: buildExec(), params });
      setTestResult(r);
    } catch (e) {
      setTestResult({ output: (e as Error).message, is_error: true });
    } finally {
      setTesting(false);
    }
  }

  return (
    <Sheet open={!!edit} onOpenChange={(o) => !o && onClose()}>
      <SheetContent
        side="right"
        className="flex flex-col gap-0 p-0 data-[side=right]:w-[45vw] data-[side=right]:sm:max-w-[45vw] data-[side=right]:min-w-[480px]"
      >
        <SheetHeader className="px-4">
          <SheetTitle>{isNew ? t("Create custom tool") : t("Edit {key}", { key: tool?.key ?? "" })}</SheetTitle>
          <SheetDescription>
            {t(
              "shell = a bash environment declaration (name + description only, telling the model bash is available); command/script/http need an execution spec.",
            )}
          </SheetDescription>
        </SheetHeader>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-4 pb-4">
          <div className="grid gap-1.5">
            <Label className="text-xs">{t("Key")}</Label>
            <Input className="font-mono" placeholder={t("e.g. nmap_scan")} value={key} disabled={!isNew}
              onChange={(e) => setKey(e.target.value)} />
          </div>
          <div className="grid gap-1.5">
            <Label className="text-xs">{t("Description (sent to the model)")}</Label>
            <Textarea rows={2} value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>

          <div className="grid gap-1.5">
            <Label className="text-xs">{t("Type")}</Label>
            <Select value={kind} onValueChange={(v) => setKind(v as "shell" | "command" | "script" | "http")}>
              <SelectTrigger className="w-56"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="shell">{t("shell (bash environment declaration)")}</SelectItem>
                <SelectItem value="command">{t("command (shell command template)")}</SelectItem>
                <SelectItem value="script">{t("script (Python script)")}</SelectItem>
                <SelectItem value="http">{t("http (API request)")}</SelectItem>
              </SelectContent>
            </Select>
            {kind === "shell" && (
              <p className="text-muted-foreground text-xs">{t("Suited to well-known tools like nmap, sqlmap, ffuf — the model already knows their usage; you only need to declare \"can be invoked in bash\". The name + description are appended to the Bash tool's description.")}</p>
            )}
          </div>

          {kind === "command" && (
            <div className="grid gap-1.5">
              <Label className="text-xs">{t("Command template (placeholder {param}, e.g. nmap -p {ports} {target})")}</Label>
              <Textarea className="font-mono text-xs" rows={2} value={ex.command}
                onChange={(e) => setEx({ ...ex, command: e.target.value })} />
            </div>
          )}
          {kind === "script" && (
            <div className="grid gap-1.5">
              <Label className="text-xs">{t("Python body (params arrive via stdin JSON / os.environ[\"TOOL_X\"])")}</Label>
              <Textarea className="font-mono text-xs" rows={10} value={ex.code}
                placeholder={'import json,sys\nargs=json.load(sys.stdin)\nprint(...)'}
                onChange={(e) => setEx({ ...ex, code: e.target.value })} />
            </div>
          )}
          {kind === "http" && (
            <div className="grid gap-2">
              <div className="flex gap-2">
                <div className="grid gap-1.5">
                  <Label className="text-xs">{t("Method")}</Label>
                  <Input className="w-24" value={ex.method} onChange={(e) => setEx({ ...ex, method: e.target.value })} />
                </div>
                <div className="grid flex-1 gap-1.5">
                  <Label className="text-xs">{t("URL (may contain {param})")}</Label>
                  <Input className="font-mono text-xs" value={ex.url} onChange={(e) => setEx({ ...ex, url: e.target.value })} />
                </div>
              </div>
              <div className="grid gap-1.5">
                <Label className="text-xs">{t("Headers (JSON, may contain {param})")}</Label>
                <Textarea className="font-mono text-xs" rows={2} value={ex.headers}
                  placeholder={'{"Authorization": "Bearer {token}"}'} onChange={(e) => setEx({ ...ex, headers: e.target.value })} />
              </div>
              <div className="grid gap-1.5">
                <Label className="text-xs">{t("Body (may contain {param})")}</Label>
                <Textarea className="font-mono text-xs" rows={2} value={ex.body} onChange={(e) => setEx({ ...ex, body: e.target.value })} />
              </div>
              <div className="flex items-center gap-4">
                <div className="grid gap-1.5">
                  <Label className="text-xs">{t("Proxy URL (empty = direct)")}</Label>
                  <Input className="font-mono text-xs w-56" value={ex.proxy} onChange={(e) => setEx({ ...ex, proxy: e.target.value })} />
                </div>
                <label className="mt-4 flex items-center gap-2 text-sm">
                  <Checkbox checked={ex.use_recording_proxy} onCheckedChange={(v) => setEx({ ...ex, use_recording_proxy: !!v })} />
                  {t("Use recording proxy")}
                </label>
              </div>
            </div>
          )}

          {kind !== "shell" && (
            <div className="flex items-center gap-3">
              <div className="grid gap-1.5">
                <Label className="text-xs">{t("Timeout (ms, empty = default)")}</Label>
                <Input type="number" className="w-32" value={ex.timeout_ms} onChange={(e) => setEx({ ...ex, timeout_ms: e.target.value })} />
              </div>
            </div>
          )}

          {kind !== "shell" && (
            <div className="grid gap-1.5">
              <Label className="text-xs">
                {kind === "http"
                  ? t("Parameter JSON Schema (required for http tools, must include properties)")
                  : t("Parameter JSON Schema (empty = a thin {args} shell is generated automatically)")}
              </Label>
              <Textarea className="font-mono text-xs" rows={4} value={schemaText}
                placeholder={'{"type":"object","properties":{"target":{"type":"string"}},"required":["target"]}'}
                onChange={(e) => setSchemaText(e.target.value)} />
            </div>
          )}

          <div className="grid gap-1.5">
            <Label className="text-muted-foreground text-xs">{t("Bind agents")}</Label>
            <div className="flex flex-wrap gap-3">
              {agents.map((a) => (
                <label key={a.key} className="flex items-center gap-2 text-sm">
                  <Checkbox checked={bound.includes(a.key)} onCheckedChange={() => toggleAgent(a.key)} />
                  {a.name}<span className="text-muted-foreground font-mono text-xs">{a.key}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-6">
            <label className="flex items-center gap-2 text-sm">
              <Switch checked={enabled} onCheckedChange={setEnabled} /> {t("Enable")}
            </label>
            {kind !== "shell" && (
              <label className="flex items-center gap-2 text-sm">
                <Switch checked={deferred} onCheckedChange={setDeferred} /> {t("deferred (enable only for many rarely-used tools)")}
              </label>
            )}
          </div>

          {kind !== "shell" && (
            <div className="grid gap-1.5 rounded-md border p-3">
              <Label className="text-xs font-medium">{t("Test run (uses the current form, nothing is saved)")}</Label>
              <Textarea className="font-mono text-xs" rows={2} value={paramsText}
                placeholder={t('Sample params JSON, e.g. {"target":"example.com"}')}
                onChange={(e) => setParamsText(e.target.value)} />
              <div>
                <Button size="sm" variant="outline" onClick={runTest} disabled={testing}>
                  <PlayIcon /> {testing ? t("Running…") : t("Test run")}
                </Button>
              </div>
              {testResult && (
                <pre
                  className={
                    "max-h-64 overflow-auto whitespace-pre-wrap break-words rounded bg-muted p-2 font-mono text-xs " +
                    (testResult.is_error ? "text-destructive" : "")
                  }
                >
                  {testResult.output || t("(no output)")}
                </pre>
              )}
            </div>
          )}
        </div>

        <Separator />
        <div className="flex items-center gap-2 p-4">
          <Button size="sm" onClick={save} disabled={saving}>
            <SaveIcon /> {isNew ? t("Create") : t("Save")}
          </Button>
          {!isNew && (
            <Button size="sm" variant="outline" className="text-destructive" onClick={del}>
              <Trash2Icon /> {t("Delete")}
            </Button>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
