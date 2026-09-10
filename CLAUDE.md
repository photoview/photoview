# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Photoview is a self-hosted photo gallery. It's a monorepo with two independent apps that talk over
GraphQL:

- `api/` — Go backend (GraphQL server via gqlgen, filesystem scanner, face detection, media
  encoding/thumbnailing).
- `ui/` — React + TypeScript frontend (Vite, Apollo Client, styled-components + Tailwind, i18next).

## Development environment

Docker is the recommended dev setup (works uniformly across platforms; local setup requires native
deps like libheif, dlib, libmagic and is best-effort only — see README for the full native
dependency list if Docker isn't an option).

```console
$ docker compose -f dev-compose.yaml build   # first-time / after dependency changes
$ docker compose -f dev-compose.yaml up      # launches both servers, restarts on code changes
```

- UI: http://localhost:1234
- GraphQL playground: http://localhost:4001
- Default DB is SQLite. For MySQL/Postgres: `docker compose -f dev-compose.yaml --profile mysql up`
  (or `--profile postgres`), after updating `PHOTOVIEW_DATABASE_DRIVER` in `dev-compose.yaml`.

Local (non-Docker) setup, once native deps are installed and `api/.env` / `ui/.env` are created
from the `example.env` files:

```console
$ cd api && go run .                 # or: reflex -g '*.go' -s -- go run .
$ cd ui && npm install && npm start  # or: npm run mon
```

## Common commands

### API (Go, in `api/`)

```console
$ go build ./...
$ go generate ./...                          # regenerate gqlgen code after editing any *.graphql file
$ go test ./...                              # unit tests only
$ go test ./... -database -filesystem -p 1   # include DB/filesystem integration tests (serialized: -p 1)
$ go test ./graphql/models/actions/... -run TestSearch -v   # single package/test
```

- **Always run `go generate ./...` after touching a `.graphql` schema file or `gqlgen.yml`**, and
  commit the result. `graphql/generated.go` and `graphql/models/generated.go` are generated —
  don't hand-edit them. `scripts/test_is_generated_code_in_sync.sh` (part of CI) fails the build if
  generated output doesn't match a fresh `go generate ./...`.
- `-database` and `-filesystem` are custom test flags (`api/test_utils/flags`) that gate
  integration tests requiring a real DB/filesystem; CI runs with both enabled against sqlite,
  mysql, and postgres.

### UI (in `ui/`)

```console
$ npm run lint          # eslint (type-aware; also catches most TS type errors — no separate tsc script)
$ npm run format:check  # prettier --check
$ npm test               # vitest, watch mode
$ npx vitest run path/to/file.test.tsx   # single test file
$ npm run build
$ npm run genSchemaTypes   # regenerate GraphQL TS types from src/**/*.graphql queries; needs a running API for introspection
```

- `genSchemaTypes` (Apollo codegen) must be re-run after adding/changing any `gql` query/mutation
  in the UI — it regenerates the corresponding `__generated__/*.ts` files, which are committed and
  should not be hand-edited.
- Translations: source strings live inline via `t('key', 'Default text')` calls; there is no local
  `en.json` to edit by hand for new strings — `npm run extractTranslations` (i18next-parser) pulls
  them into `src/extractedTranslations/<locale>/translation.json`. Actual translated bundles are
  managed by an external translation service, not authored by hand in this repo.

## Architecture

### API (`api/`)

- `main.go` / `server.go` wire everything together: load `.env`, connect the DB, start the
  scanner's background workers, mount the GraphQL endpoint and REST-ish media routes, start the
  periodic scanner.
- `graphql/resolvers/*.graphql` are the GraphQL schema (split by domain: album, media, user,
  search, faces, scanner, share_token, timeline, ...). `gqlgen.yml` maps schema types to Go models
  in `graphql/models/`. Resolvers follow gqlgen's "follow-schema" layout: one `<name>.go` per
  schema file in `graphql/resolvers/`.
- `graphql/auth/` — request authentication/authorization middleware for the GraphQL endpoint.
- **Album permission model**: `AlbumPermissionLevel` (`READ`/`UPLOAD`/`DELETE`, ordered) is stored
  per `(user, album)` in `UserAlbums.Level`, materialized down the album tree at grant/creation
  time rather than computed live — `PropagateAlbumLevel`/`RevokeAlbumLevel`
  (`graphql/models/album.go`) stamp a grant onto a whole existing subtree, and the scanner /
  `CreateAlbumFolder` copy a parent's grants onto new children as they're discovered. `User.
  HasAlbumLevel`/`EffectiveGrant`/`IsAlbumOwner` (`graphql/models/user.go`) are the standard
  authorization checks — admins bypass, everyone else is a single indexed lookup, no ancestor walk.
  A grant's `GrantedByUserID` is nil for a root/admin-configured grant (an "owner") or the granting
  owner's user ID for a peer share (`actions.GrantAlbumAccess`/`RevokeAlbumAccess`,
  `graphql/models/actions/sharing_actions.go`) — only an owner may share further, capped at their
  own level. `UserAlbums` is a **materialized cache**: the actual source of truth for *why* a user
  has access is `UserAlbumGrant` (one row per `(user, album, source_album)`), and
  `PropagateAlbumLevel`/`RevokeAlbumLevel` write/delete only the rows for the source they were
  called with, then recompute `UserAlbums.Level` as the max across every remaining source for that
  `(user, album)` pair (`recomputeUserAlbums`, `graphql/models/album.go`) — so two independent
  grants reaching the same descendant (e.g. an admin's root grant and a peer share of a nested
  folder) coexist instead of one silently overwriting the other, and revoking one source never
  touches another source's rows. Existing installs get their pre-existing `UserAlbums` rows
  backfilled as self-sourced `UserAlbumGrant` rows on first migration
  (`backfillUserAlbumGrants`, `database/database.go`) — this can't recover provenance for grants
  already clobbered before this existed, only prevents new clobbers going forward. `MoveAlbum`
  (`graphql/resolvers/upload.go`) reparents an album inside the same transaction that rewrites the
  subtree's paths, then calls `RecomputeGrantsAfterMove` (`graphql/models/album.go`) to drop grants
  sourced outside the moved subtree and inherit the new parent's — so access follows the move
  rather than going stale in either direction.
- `dataloader/` — per-request batching/caching (`gen_*.go` are generated dataloaders, via
  `github.com/vektah/dataloaden` — no `go:generate` directive wires it up, so a new loader is
  generated by running the tool manually and hand-writing its `New*Loader(db)` constructor
  alongside, see `dataloader/albumGrantLoader.go`) to avoid N+1 queries when resolving nested
  GraphQL fields (e.g. media URLs, user favorites, per-album permission grants for
  `viewerCanUpload`/`viewerCanDelete`/`viewerIsOwner`).
- `database/` — GORM setup and migrations; `database/drivers/` abstracts sqlite/mysql/postgres.
  Most schema changes apply via GORM `AutoMigrate`, but a few columns are added via hand-written
  raw SQL in `database.go` instead (e.g. `user_albums.level`/`granted_by_user_id`, via driver-
  specific `information_schema`/`PRAGMA` introspection) when `AutoMigrate` can't be trusted to
  detect them reliably across sqlite/mysql/postgres. Any migration step in `MigrateDatabase` that
  guards access control (not just cosmetic/data-quality cleanup) must return its error rather than
  log-and-continue, since `server.go` panics on a non-nil error from it — swallowing the error lets
  the server start with the DB in a half-migrated, possibly under-permissioned state.
- `scanner/` — the filesystem scanning pipeline: walks album directories (`scanner_album.go`),
  processes media (`scanner_media.go`), generates thumbnails/encodes video
  (`media_encoding/executable_worker`), runs face detection (`face_detection/`), and is driven by a
  queue (`scanner_queue/`) plus a periodic trigger (`periodic_scanner/`). `scanner_queue.
  GetQueueStatus`/`CancelJob` (surfaced as the `scannerQueueStatus` query / `cancelScanJob`
  mutation) introspect and cancel individual queued/running jobs; cancellation is cooperative
  (checked between files via a per-job `context.CancelFunc`, not mid-file) so an in-progress
  encode always finishes cleanly rather than leaving a half-written cache file.
- `graphql/models/actions/` — business logic invoked by resolvers that's more than trivial CRUD
  (e.g. search).

### UI (`ui/src`)

- `apolloClient.ts` — Apollo Client setup (HTTP + WebSocket links, auth token attachment).
- `localization.ts` — i18next setup; loads the active user's language as a translation bundle at
  runtime.
- `Pages/` — top-level routed pages (one folder per page, e.g. `SettingsPage`, `SearchPage`,
  `AlbumPage`).
- `components/` — shared feature components (album grid/tree, photo gallery, sidebar, header,
  mapbox integration, timeline gallery, etc.), generally colocated with their own `__generated__`
  GraphQL types where they define queries.
- `primitives/` — low-level, page-agnostic UI building blocks (form inputs, modal, table, loader).
- `helpers/` and `hooks/` — cross-cutting utilities and shared React hooks.
- Styling is a mix of styled-components (older code) and Tailwind utility classes (newer code) —
  match whichever convention the file you're editing already uses.

## Known limitations / future work

- **Scanner discovery blocks the GraphQL request.** `AddUserToQueue` and `AddAlbumToQueue`
  (`api/scanner/scanner_queue/queue.go`) call `FindAlbumsForUser`/`FindAlbumsForAlbum`
  (`api/scanner/scanner_user.go`) synchronously from the resolver. `walkAlbumScanQueue` walks the
  *entire* directory subtree and opens a DB transaction per directory before anything is queued for
  the background scanner workers — so scanning a large album/library can block the request for
  seconds to minutes, risking a client/proxy timeout even though the scan itself would otherwise
  succeed. Both the pre-existing "scan my library" resolver and the "rescan this album" mutation
  share this pattern. A proper fix means introducing a "discover children of this album" job type
  that the scanner workers process asynchronously (queueing further discovery/scan jobs themselves),
  rather than the resolver doing the full walk inline — a real change to the job/queue model
  (`ScannerJob`, job dedup in `queue.go`) affecting both entry points, not a quick patch.
- **Uploads send the credentialed cookie without enforcing HTTPS.** `SidebarAlbumUpload.tsx` sets
  `xhr.withCredentials = true` so the `auth-token` cookie rides along with the upload request, but
  neither the cookie (set client-side via `js-cookie`, not `Secure`-flagged) nor `API_ENDPOINT` are
  checked to require HTTPS. On an HTTP deployment this is consistent with the rest of the app (the
  cookie is already sent over HTTP everywhere else), so the fix is deployment-policy-shaped rather
  than a pure code change — see `HANDOFF_open_findings.md`.

## PR expectations (from CONTRIBUTING.md)

- Target the `master` branch.
- Keep existing users' data/config migratable with no (or minimal, well-documented) manual steps.
- CI (build, tests, lint) must be clean before merge; generated code (gqlgen output, Apollo
  codegen output, extracted translations where applicable) must be committed in sync with source
  changes.
