# Handoff: open findings on PR #1473

Findings from the CodeRabbit review passes on PR #1473
(`feature/configurable-search-result-limit` → `photoview/photoview:master`) that are _not_ code
fixes and therefore stay open until someone makes a call on them. Everything that was fixed lives
in the branch history and, where it shaped the architecture, in `CLAUDE.md` — it is deliberately
not repeated here, so this file doesn't drift out of date as the code moves.

## Credentialed uploads don't enforce HTTPS

**Where:** `ui/src/components/sidebar/SidebarAlbumUpload.tsx` (`xhr.withCredentials = true`),
`ui/src/helpers/authentication.ts` (`saveTokenCookie`, sets the `auth-token` cookie without a
`Secure` attribute).

**What's happening:** the upload request sends the `auth-token` cookie via `withCredentials`, and
neither `API_ENDPOINT` nor the cookie itself are restricted to HTTPS. On an HTTP-only deployment
this means the token could be sniffed in transit.

**Why it isn't simply fixed:** this cookie is sent over HTTP on _every_ authenticated request
already, not just uploads — `SidebarAlbumUpload.tsx` isn't a special case, it's just where the
review happened to flag it. Forcing `Secure` on the cookie would break any deployment still running
plain HTTP, a real and currently-supported configuration per the README. So this is a
deployment-policy question rather than a pure code fix: does the project want to (a) require HTTPS
in production and document it, (b) make `Secure` conditional on `location.protocol === 'https:'`,
or (c) leave it as-is and treat it as accepted risk for self-hosted deployments behind a trusted
network or reverse proxy.

**Suggested next step:** ask the maintainers which posture they want before touching this — it's a
decision about the project's deployment story, not something to pick unilaterally.
