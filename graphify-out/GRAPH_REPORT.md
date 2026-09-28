# Graph Report - img-upload-view  (2026-09-28)

## Corpus Check
- Corpus is ~5,721 words - fits in a single context window. You may not need a graph.

## Summary
- 196 nodes · 266 edges · 23 communities (10 shown, 4 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 7 edges (avg confidence: 0.89)
- Token cost: 41,876 input · 0 output

## Community Hubs (Navigation)
- Web TS Config
- Image HTTP Handlers
- Web Runtime Deps
- Web Dev Deps
- Error Response Helpers
- API Bootstrap & Config
- Docker Compose Stack
- Server & Middleware Setup
- Image Data Model
- CORS & Recover Middleware
- Gallery Page
- CLI Stub
- Welcome Component
- Go Module

## God Nodes (most connected - your core abstractions)
1. `compilerOptions` - 16 edges
2. `ErrorResponse` - 15 edges
3. `Application` - 11 edges
4. `routes()` - 9 edges
5. `ImageModel` - 8 edges
6. `UploadImage()` - 7 edges
7. `GetImages()` - 7 edges
8. `WriteJSON()` - 7 edges
9. `main()` - 6 edges
10. `GetImageByID()` - 6 edges

## Surprising Connections (you probably didn't know these)
- `postgres (dev compose service)` --semantically_similar_to--> `postgres (pgvector pg15 service)`  [INFERRED] [semantically similar]
  compose.dev.yaml → compose.yaml
- `Local dev app config (config.dev.yaml)` --semantically_similar_to--> `Production app config (configs/config.yaml)`  [INFERRED] [semantically similar]
  config.dev.yaml → configs/config.yaml
- `main()` --calls--> `Serve()`  [EXTRACTED]
  cmd/api/main.go → internal/app/api/server.go
- `main()` --calls--> `NewErrorResponse()`  [EXTRACTED]
  cmd/api/main.go → pkg/errors/errors.go
- `Application` --references--> `ErrorResponse`  [EXTRACTED]
  internal/config/application.go → pkg/errors/errors.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Production docker compose stack (nginx -> frontend/API -> Postgres)** — compose_reverse_proxy, compose_frontend, compose_api_service, compose_postgres [EXTRACTED 1.00]
- **Image upload storage and static serving via shared volume** — compose_api_service, compose_uploaded_images, compose_reverse_proxy [INFERRED 0.85]

## Communities (23 total, 4 thin omitted)

### Community 0 - "Web TS Config"
Cohesion: 0.08
Nodes (26): **/*, **/.client/**/*, DOM, DOM.Iterable, ES2022, node, .react-router/types/**/*, **/.server/**/* (+18 more)

### Community 1 - "Image HTTP Handlers"
Cohesion: 0.16
Nodes (18): net/http.HandlerFunc, net/http.Header, net/url.Values, envelope, DeleteImage(), generateUniqueFilename(), GetImageByID(), GetImages() (+10 more)

### Community 2 - "Web Runtime Deps"
Cohesion: 0.09
Nodes (21): isbot, react, react-dom, react-router, @react-router/node, @react-router/serve, dependencies, isbot (+13 more)

### Community 3 - "Web Dev Deps"
Cohesion: 0.11
Nodes (19): @react-router/dev, tailwindcss, @tailwindcss/vite, @types/node, @types/react, @types/react-dom, typescript, vite (+11 more)

### Community 4 - "Error Response Helpers"
Cohesion: 0.29
Nodes (7): envelope, ErrorResponse, net/http.Request, net/http.ResponseWriter, WriteFile(), NewErrorResponse(), writeJSON()

### Community 5 - "API Bootstrap & Config"
Cohesion: 0.15
Nodes (11): main(), Config, IImageModel, DBConfig, database/sql.DB, time.Duration, Models, NewModels() (+3 more)

### Community 6 - "Docker Compose Stack"
Cohesion: 0.21
Nodes (14): api-network, api-service (Go API service), postgres (dev compose service), postgres_data volume (dev), frontend (Vite web service), postgres (pgvector pg15 service), postgres_data volume, repository-network (+6 more)

### Community 7 - "Server & Middleware Setup"
Cohesion: 0.26
Nodes (10): zerologWriter, github.com/rs/zerolog.Logger, Serve(), T, New(), WithErrorResponse(), WithTrustedOrigins(), WithZerolog() (+2 more)

### Community 8 - "Image Data Model"
Cohesion: 0.33
Nodes (3): Image, ImageModel, time.Time

### Community 9 - "CORS & Recover Middleware"
Cohesion: 0.40
Nodes (3): net/http.Handler, Middlewares[T], Middlewares[T]

## Knowledge Gaps
- **53 isolated node(s):** `github.com/khofesh/img-upload-view`, `envelope`, `Middlewares[T]`, `Middlewares[T]`, `envelope` (+48 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 83 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Application` connect `Image HTTP Handlers` to `Error Response Helpers`, `API Bootstrap & Config`, `Server & Middleware Setup`?**
  _High betweenness centrality (0.061) - this node is a cross-community bridge._
- **Why does `ErrorResponse` connect `Error Response Helpers` to `Image HTTP Handlers`, `Server & Middleware Setup`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **Why does `routes()` connect `Image HTTP Handlers` to `CORS & Recover Middleware`, `Server & Middleware Setup`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **What connects `github.com/khofesh/img-upload-view`, `envelope`, `Middlewares[T]` to the rest of the system?**
  _53 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Web TS Config` be split into smaller, more focused modules?**
  _Cohesion score 0.07977207977207977 - nodes in this community are weakly interconnected._
- **Should `Web Runtime Deps` be split into smaller, more focused modules?**
  _Cohesion score 0.09090909090909091 - nodes in this community are weakly interconnected._
- **Should `Web Dev Deps` be split into smaller, more focused modules?**
  _Cohesion score 0.10526315789473684 - nodes in this community are weakly interconnected._