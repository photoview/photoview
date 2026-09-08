# Handoff: open CodeRabbit findings on PR #1473

Status as of 2026-09-08. PR #1473 (`feature/configurable-search-result-limit` →
`photoview/photoview:master`) has had three CodeRabbit review passes today (11:25, 11:48, 12:51),
producing 23 findings. 20 were fixed and deployed the same day (16 commits, `530aaba`..`41f53bd`,
image tag `coderabbit-fixes-batch2`). The 3 below were deliberately deferred — either because they
need a real design decision, or because they're a deployment-policy question rather than a code
fix — and are still open.

## 1. Grant provenance: `PropagateAlbumLevel`/`RevokeAlbumLevel` still have no memory of source

**Where:** `api/graphql/models/user.go:52` (the `UserAlbums.GrantedByUserID` field comment),
`api/graphql/models/album.go:91` (`PropagateAlbumLevel`) and `:115` (`RevokeAlbumLevel`).

**What's already fixed:** `actions.GrantAlbumAccess`/`RevokeAlbumAccess`
(`api/graphql/models/actions/sharing_actions.go`) now refuse to act, for non-admin actors, when the
target user already holds a grant anywhere in the album's subtree that traces to a different
grantor (see `targetHasForeignGrantInSubtree`, added 2026-09-08, commit `44bc3db`). That closes the
sharing-UI path: a peer share can no longer silently overwrite an admin's direct grant, or another
owner's independent share of a nested folder.

**What's still open:** the check above only guards the two `sharing_actions.go` entry points.
`PropagateAlbumLevel`/`RevokeAlbumLevel` themselves are unconditional bulk upsert/delete over
`UserAlbums`, and are also called from:

- The scanner's copy-parent-grants-onto-new-children step (new albums discovered under an existing
  grant inherit it via `PropagateAlbumLevel`).
- `NewRootAlbum` (`api/scanner/scanner_album.go`) when a new root path is created.
- An **admin** granting/revoking access — admins bypass the foreign-grant check entirely (by
  design, since they're allowed to override), so an admin action can still clobber a peer share the
  same way.

Because `UserAlbums` has exactly one row per `(user_id, album_id)` — `Level` plus a single nullable
`GrantedByUserID` — there is fundamentally no way to represent "this user has independent access to
this album from two different sources, keep the higher/most-specific one" or "only remove the part
of this grant that came from album X". Two sources always collapse into one row, and whichever
write happens last wins.

**Why this hasn't been fixed today:** CodeRabbit tagged this "Heavy lift" and it's a real
data-model change, not a quick patch. It needs a design decision before implementation, e.g.:

- Track grant provenance per source (one row per `(user, album, source_album)`, take the highest
  level across sources, only delete the row for the matching source on revoke) — most correct, but
  changes the `UserAlbums` schema and every place that reads a user's access level.
- Or: make `PropagateAlbumLevel`/`RevokeAlbumLevel` themselves subtree-scan for foreign grants the
  same way `sharing_actions.go` now does, and skip (rather than clobber) any row that doesn't trace
  back to the actor — cheaper, but changes propagation semantics (a subtree with a mix of grant
  sources would end up with genuinely inconsistent levels across siblings, which may be the more
  honest outcome anyway).

**Suggested next step:** bring this to the user as a design question (which of the two directions
above, or something else) before writing code — this is exactly the kind of decision the session
convention has been pausing on rather than silently picking one.

## 2. Album-tree filtering fans out one GraphQL request per visible node

**Where:** `ui/src/components/albumTree/AlbumTreeNode.tsx:73-94`, and the sibling
`ui/src/components/albumTree/AlbumTree.tsx:52-58` (the search/filter query that produces
`visibleIds`/`matchedIds`).

**What's happening:** when a filter is active, `isFiltering` forces `isExpanded = true` for every
rendered node (line 74), and each node's own `useEffect` (line ~90-94) calls `fetchSubAlbums()` the
first time it renders expanded. Since `limitAlbums: 0` on the search means "return every match",
searching a large tree can render (and therefore fetch-for) every matched album *and* every one of
its ancestors simultaneously — one `albumTreeSubAlbumsQuery` GraphQL request each, all fired within
the same render pass.

**Why this hasn't been fixed today:** also tagged "Heavy lift". A real fix isn't a per-node
change — it's giving the tree enough data upfront (either the search result itself returning the
tree shape, or a dedicated batched "give me children for these N album IDs" query) so filtered
nodes don't each need their own round trip. That's a shape change to either the search resolver or
a new tree-batching query, not a one-file patch.

**Suggested next step:** worth asking the user how large their album trees / search result sets
typically get in practice — if it's a handful of matches this may not be worth the redesign; if
someone has thousands of albums it's a real scaling problem worth prioritizing.

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
