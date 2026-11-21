# Norma process lifecycle integration boundary

## Decision

ARTEX must create a `collection_runs` row at the point a child process is actually started, not when an agent intent is created, when a tool call is approved, or when an output file is ingested.

The pinned dependency is `github.com/Autumn-27/norma v0.4.11`, replaced by `github.com/BaaRai/Norma v0.4.11` in `go.mod`. The process-launch implementation was inspected at that exact tag.

## Confirmed launch paths in Norma v0.4.11

| Execution path | Launch boundary | Lifecycle information available |
|---|---|---|
| Managed foreground/background Bash | `tool.Manager.Spawn` calls `exec.Cmd.Start` | task ID, command, kind, working directory, output path, start/end, exit status |
| Synchronous Bash fallback | `tool.runBashSync` calls `exec.Cmd.Run` | command and result/error, but no persisted task record or output file |
| Interactive PTY | `Manager.openSession` calls `startPTY` | session ID, command, PTY output stream, exit code; no ordinary task output file |

The normal Bash tool chooses the managed path when `ToolContext.Tasks` exists and background tasks are enabled. Otherwise it executes synchronously. Interactive shell tools use the separate PTY path.

## Why existing hooks are insufficient

Norma's harness `PreToolUse` and `PostToolUse` hooks bracket a tool call. They do not bracket an individual process launch. A Bash call can launch multiple processes, shell pipelines, or no process at all; background commands outlive the tool call. A hook-generated run ID would therefore be incorrectly scoped and could not reliably represent process start/exit.

The Norma `SpawnSpec` currently has command, description, kind, working directory and environment, but no host lifecycle callback or process metadata field. `ToolContext` exposes the task manager but no process-start observer. The synchronous Bash fallback and PTY path also bypass the ordinary background-task lifecycle.

## Required integration seam

Add a host-extensible process lifecycle observer in the Norma tool layer (preferably in the `Manager` plus the synchronous Bash and PTY launch paths), with a stable event contract:

- **Starting**: execution ID, tool-use ID, agent/session context, command identity, execution mode, working directory, start timestamp. Emit only after permission/pre-tool gates and immediately before the OS process start.
- **Started**: execution ID plus PID and the process/task/session handle after successful start.
- **Exited**: execution ID, finish timestamp, exit code, terminal status, and output locator when one exists.
- **StartFailed**: execution ID and sanitized start error.
- **Cancelled/Killed**: terminal reason and timestamp.

The observer must be optional and non-blocking with respect to the public tool API; observer errors must fail closed for evidence persistence without silently changing command execution semantics. It must not log environment values or secrets. Command text should be treated as sensitive operator input and redacted or access-controlled in persisted metadata.

## ARTEX wiring sequence

1. Pass an observer from ARTEX's session/tool-context assembly into Norma's process manager.
2. On each real process start, create a collection-run row and retain its UUID against that exact process execution (not the worker intent).
3. On successful exit or failure, finish that same row with the actual status, exit code and output locator.
4. For managed Bash, import the completed output file through `evidence.StoreCollectionArtifact`; register the resulting digest-backed evidence metadata against the same run.
5. For PTY, define a bounded, explicit output-artifact capture policy before claiming evidence completeness. A live ring buffer is not itself immutable evidence.
6. Pass the execution/run ID and registered evidence reference to the parser/ingestion adapter. Persist normalized claims only after evidence registration succeeds.
7. Test multiple process launches inside one tool call, pipelines, background completion after tool return, timeout/kill, start failure, sync fallback, PTY exit, and repeated ingestion.

## Current ARTEX state

ARTEX already has collection-run and evidence metadata persistence, task/tool identity checks, and a confined content-addressed artifact import method. The missing production seam is process-level lifecycle notification in Norma. Do not substitute `agent.RunInfo` (worker/intent scope), a tool-call hook, or an ingestion-attempt UUID for the process execution ID.

This document records a source-inspected integration boundary; it does not claim that process lifecycle wiring or end-to-end evidence ingestion is implemented.