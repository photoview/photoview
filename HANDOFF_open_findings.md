# Handoff: open CodeRabbit findings on PR #1473

Status as of 2026-09-08. PR #1473 (`feature/configurable-search-result-limit` →
`photoview/photoview:master`) has had three CodeRabbit review passes today (11:25, 11:48, 12:51),
producing 23 findings. 20 were fixed and deployed the same day (16 commits, `530aaba`..`41f53bd`,
image tag `coderabbit-fixes-batch2`). Two more (grant provenance, album-tree filter fanout) were
deliberately deferred pending a design decision, then implemented later the same day once the user
picked a direction — see below. Item 3 (HTTPS enforcement) is a deployment-policy question, not a
code fix, and is still open.

## 1. Grant provenance: `PropagateAlbumLevel`/`RevokeAlbumLevel` had no memory of source — RESOLVED

**Where:** `api/graphql/models/user.go` (`UserAlbumGrant`), `api/graphql/models/album.go`
(`PropagateAlbumLevel`, `RevokeAlbumLevel`, `recomputeUserAlbums`), `api/database/database.go`
(`backfillUserAlbumGrants`).

**What was wrong:** `UserAlbums` had exactly one row per `(user_id, album_id)`, so two independent
grants reaching the same album from different sources (e.g. an admin's root grant and a peer's
share of a nested folder) collapsed into one row — whichever `PropagateAlbumLevel` call ran last
silently won, and `RevokeAlbumLevel` could delete access that traced to an unrelated source.
`actions.GrantAlbumAccess`/`RevokeAlbumAccess` had a same-day stopgap (`targetHasForeignGrantInSubtree`)
that refused the operation instead of clobbering, but admin actions and the scanner's
copy-parent-grants-onto-new-children step still went through the unconditional `PropagateAlbumLevel`/
`RevokeAlbumLevel` path.

**Fix:** the user picked "track provenance per source". Added `UserAlbumGrant`, one row per
`(user, album, source_album)`; `UserAlbums` is now a materialized cache recomputed from it (max
level across sources, owner-rooted if any source is) after every propagate/revoke. Two independent
grants on the same album now coexist instead of colliding, so the `targetHasForeignGrantInSubtree`
stopgap was removed — a peer share into a subtree that already has an unrelated grant now succeeds
(coexists) rather than being refused, which is the intended outcome. Existing `UserAlbums` rows are
backfilled as self-sourced `UserAlbumGrant` rows on migration; this is forward-looking only — grants
already clobbered before today can't be un-clobbered, since the old single-row model never recorded
which source it lost.

## 2. Album-tree filtering fanned out one GraphQL request per visible node — RESOLVED

**Where:** `ui/src/components/albumTree/AlbumTree.tsx`, `AlbumTreeNode.tsx`,
`api/graphql/resolvers/album.graphql`/`album.go` (`albumTreeChildren`).

**What was wrong:** when a filter was active, every visible node (every match *and* every one of
its ancestors) rendered expanded and fired its own `subAlbums` request via `useEffect` - a broad
match against a large tree could queue hundreds of simultaneous GraphQL requests.

**Fix:** the user confirmed a real library (~50,000 photos over 20 years of album folders), i.e.
this was a real scaling problem. Added a batched `albumTreeChildren(albumIds, showHidden)` query
that returns direct children for a whole list of album ids in one round trip (reusing `subAlbums`'s
own authorization/hidden-filtering logic). `AlbumTree.tsx` now fires this once for the full
`visibleIds` set when filtering; `AlbumTreeNode` reads from the batched result instead of issuing
its own request while filtering. Normal (non-filtering) single-node expand/collapse browsing is
unchanged.

## 3. Credentialed uploads don't enforce HTTPS

**Where:** `ui/src/components/sidebar/SidebarAlbumUpload.tsx:51` (`xhr.withCredentials = true`),
`ui/src/helpers/authentication.ts` (`saveTokenCookie`, sets the `auth-token` cookie without a
`Secure` attribute).

**What's happening:** the upload request sends the `auth-token` cookie via `withCredentials`, and
neither `API_ENDPOINT` nor the cookie itself are restricted to HTTPS. On an HTTP-only deployment
this means the token could be sniffed in transit.

**Why this hasn't been fixed today:** this cookie is sent over HTTP on *every* authenticated
request already (not just uploads) — `SidebarAlbumUpload.tsx` isn't a special case, it's just where
CodeRabbit happened to flag it this pass. Forcing `Secure` on the cookie would break any deployment
still running plain HTTP (a real, currently-supported configuration per the README), so this is a
deployment-policy question, not a pure code fix: does the project want to (a) require HTTPS in
production and document it, (b) make `Secure` conditional on `location.protocol === 'https:'`, or
(c) leave it as-is and treat this as accepted risk for self-hosted deployments behind a trusted
network/reverse proxy.

**Suggested next step:** ask the user/maintainers which posture they want before touching this —
it's a decision about the project's deployment story, not something to pick unilaterally.
