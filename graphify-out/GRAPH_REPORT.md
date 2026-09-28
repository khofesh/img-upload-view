# Graph Report - img-upload-view  (2026-09-28)

## Corpus Check
- 24 files · ~9,187 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 267 nodes · 424 edges · 25 communities (12 shown, 4 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 34 edges (avg confidence: 0.87)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- API Bootstrap & S3 Clients
- Data Models & Config Defaults
- Error Responses
- TypeScript Compiler Config
- Image API Handlers
- Frontend Runtime Deps
- Handler Test Fakes
- Frontend Dev Deps
- Compose Services
- Middleware & Logging
- Cleanup CLI
- HTTP Middleware Stack
- Gallery Route
- Config Loader
- Welcome Route
- Go Module Root

## God Nodes (most connected - your core abstractions)
1. `Application` - 17 edges
2. `compilerOptions` - 16 edges
3. `ErrorResponse` - 16 edges
4. `Image` - 13 edges
5. `S3Store` - 13 edges
6. `fakeImageModel` - 10 edges
7. `fakeObjectStore` - 9 edges
8. `CompleteUpload()` - 9 edges
9. `newTestApp()` - 9 edges
10. `routes()` - 9 edges

## Surprising Connections (you probably didn't know these)
- `Postgres service (dev)` --semantically_similar_to--> `Postgres service (prod)`  [INFERRED] [semantically similar]
  compose.dev.yaml → compose.yaml
- `Local dev app config (config.dev.yaml)` --semantically_similar_to--> `Production app config (configs/config.yaml)`  [INFERRED] [semantically similar]
  config.dev.yaml → configs/config.yaml
- `MinIO AIStor service (dev)` --semantically_similar_to--> `MinIO AIStor service (prod)`  [INFERRED] [semantically similar]
  compose.dev.yaml → compose.yaml
- `api-service Go JSON API` --conceptually_related_to--> `Local dev app config (config.dev.yaml)`  [AMBIGUOUS]
  compose.yaml → config.dev.yaml
- `WithErrorResponse()` --references--> `ErrorResponse`  [EXTRACTED]
  internal/middleware/middlewares.go → pkg/errors/errors.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Production service topology behind nginx** — compose_reverse_proxy, compose_frontend, compose_api_service, compose_postgres, compose_minio [EXTRACTED 1.00]

## Communities (25 total, 4 thin omitted)

### Community 0 - "API Bootstrap & S3 Clients"
Cohesion: 0.12
Nodes (14): main(), context.Context, github.com/aws/aws-sdk-go-v2/service/s3.Client, github.com/aws/aws-sdk-go-v2/service/s3.PresignClient, time.Time, fakeObjectStore, NewStorage(), isNotFound() (+6 more)

### Community 1 - "Data Models & Config Defaults"
Cohesion: 0.11
Nodes (12): Storage, IImageModel, ImageModel, DBConfig, database/sql.DB, time.Duration, fakeImageModel, Config (+4 more)

### Community 2 - "Error Responses"
Cohesion: 0.15
Nodes (15): envelope, ErrorResponse, net/http.Header, net/http.Request, net/http.ResponseWriter, net/url.Values, ReadIDParam(), ReadInt() (+7 more)

### Community 3 - "TypeScript Compiler Config"
Cohesion: 0.08
Nodes (26): **/*, **/.client/**/*, DOM, DOM.Iterable, ES2022, node, .react-router/types/**/*, **/.server/**/* (+18 more)

### Community 4 - "Image API Handlers"
Cohesion: 0.20
Nodes (18): github.com/khofesh/img-upload-view/internal/data.Models, net/http.HandlerFunc, createUploadRequest, envelope, DeleteImage(), generateUniqueFilename(), GetImageByID(), GetImages() (+10 more)

### Community 5 - "Frontend Runtime Deps"
Cohesion: 0.09
Nodes (21): isbot, react, react-dom, react-router, @react-router/node, @react-router/serve, dependencies, isbot (+13 more)

### Community 6 - "Handler Test Fakes"
Cohesion: 0.23
Nodes (16): bytes.Reader, github.com/julienschmidt/httprouter.Params, testing.T, newFakeImageModel(), newFakeObjectStore(), TestDeleteImage(), TestGetImageByIDHidesPending(), TestValidateImageFile() (+8 more)

### Community 7 - "Frontend Dev Deps"
Cohesion: 0.11
Nodes (19): @react-router/dev, tailwindcss, @tailwindcss/vite, @types/node, @types/react, @types/react-dom, typescript, vite (+11 more)

### Community 8 - "Compose Services"
Cohesion: 0.16
Nodes (18): api-network, api-service Go JSON API, MinIO AIStor service (dev), Postgres service (dev), postgres_data volume (dev), frontend (Vite web service), MinIO AIStor service (prod), Postgres service (prod) (+10 more)

### Community 9 - "Middleware & Logging"
Cohesion: 0.29
Nodes (9): zerologWriter, github.com/rs/zerolog.Logger, T, New(), WithErrorResponse(), WithTrustedOrigins(), WithZerolog(), Middlewares (+1 more)

### Community 10 - "Cleanup CLI"
Cohesion: 0.47
Nodes (4): main(), Cli(), runCleanup(), usage()

### Community 11 - "HTTP Middleware Stack"
Cohesion: 0.40
Nodes (3): net/http.Handler, Middlewares[T], Middlewares[T]

## Ambiguous Edges - Review These
- `api-service Go JSON API` → `Local dev app config (config.dev.yaml)`  [AMBIGUOUS]
  compose.yaml · relation: conceptually_related_to

## Knowledge Gaps
- **56 isolated node(s):** `baseUrl`, `esModuleInterop`, `jsx`, `module`, `moduleResolution` (+51 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 94 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `api-service Go JSON API` and `Local dev app config (config.dev.yaml)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Application` connect `Image API Handlers` to `API Bootstrap & S3 Clients`, `Data Models & Config Defaults`, `Error Responses`, `Handler Test Fakes`, `Middleware & Logging`?**
  _High betweenness centrality (0.091) - this node is a cross-community bridge._
- **Why does `ErrorResponse` connect `Error Responses` to `Middleware & Logging`, `Image API Handlers`?**
  _High betweenness centrality (0.054) - this node is a cross-community bridge._
- **Why does `S3Store` connect `API Bootstrap & S3 Clients` to `Middleware & Logging`, `Handler Test Fakes`, `Data Models & Config Defaults`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **What connects `baseUrl`, `esModuleInterop`, `jsx` to the rest of the system?**
  _56 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `API Bootstrap & S3 Clients` be split into smaller, more focused modules?**
  _Cohesion score 0.11822660098522167 - nodes in this community are weakly interconnected._
- **Should `Data Models & Config Defaults` be split into smaller, more focused modules?**
  _Cohesion score 0.10591133004926108 - nodes in this community are weakly interconnected._