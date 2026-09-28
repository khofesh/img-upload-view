# claude

@AGENTS.md

Project overview, architecture, commands, and code style live in `AGENTS.md` (imported above).

## Response Style

- Be extremely concise. No pleasantries, no filler.
- When asked to write code, return code only unless explanation is explicitly requested.
- No sycophantic preambles ("Sure!", "Great question!", "Absolutely!").
- No "Here's a function that..." preambles.
- Don't restate the question before answering.
- No "Note:", "Tip:", or "Remember:" appendices unless asked.
- No usage examples unless asked.
- No unsolicited suggestions or improvements beyond what was asked.

## Context Retrieval Policy

Always retrieve the smallest amount of information necessary.
Escalate only when necessary. Stop escalating as soon as sufficient information has been obtained.

Preferred order:

1. Need code structure? `graphify-out/graph.json` (if built).
2. Need directory structure? `rtk tree`
3. Need symbols? `ast-grep`.
4. Need implementation? Source files.
5. Use repository-wide search as last resort

Avoid reading entire directories or the whole repository unless explicitly requested.
Never read `bin/`, `upload/`, `web/node_modules/`, `web/build/`, `web/.react-router/`, or
`web/package-lock.json`.

## RTK (Rust Token Killer) - Token-Optimized CLI

`rtk` is a CLI proxy that filters and compresses command outputs, saving 60-90% tokens.

**Always** prefix commands with `rtk`. If RTK has a dedicated filter, it uses it. If not, it
passes through unchanged. This means RTK is always safe to use.

```bash
rtk ls <path>
rtk read <file>
rtk find <pattern>
rtk git status
rtk git log -10

rtk err <cmd>           # Filter errors only from any command
rtk log <file>          # Deduplicated logs with counts
rtk json <file>         # JSON structure without values
rtk curl <url>          # Compact HTTP responses

rtk docker ps           # Compact container list
rtk docker logs <c>     # Deduplicated logs
```

## Graphify - Codebase context & knowledge graph protocol

The knowledge graph is **not built yet**. Build it with `/graphify .`; output goes to
`graphify-out/graph.json`. Once it exists, read it before searching or reading multiple source
files:

1. Read `graph.json`.
2. Identify the relevant symbols, files, and dependency paths.
3. Read only the source files required for the task.

Graph edges are AST-extracted structural relationships — treat them as authoritative; do not infer
dependencies absent from the graph.

Expected high-impact nodes (trace impact before changing): `config.Application`,
`data.Image` / `IImageModel`, `errors.ErrorResponse`, `middlewares.Middlewares[T]`.

Rebuild after structural changes:

```bash
/graphify . --update     # incremental, AST-only, no LLM tokens
```

## ast-grep

Prefer `ast-grep` over `grep` when searching source code.

```bash
ast-grep --lang go -p 'func $F(app *config.Application) http.HandlerFunc { $$$ }' internal/
ast-grep --lang go -p 'func (m ImageModel) $M($$$) $R { $$$ }' internal/data/
ast-grep --lang tsx -p 'fetch($$$)' web/app/
```

Use `grep` only for Markdown, YAML, JSON, SQL, nginx conf, Makefile, Dockerfiles, and logs.
