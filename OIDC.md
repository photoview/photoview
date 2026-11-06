# Optional OpenID Connect login

Photoview can offer OpenID Connect alongside its existing local password sign-in. The OIDC provider authenticates the user; Photoview continues to own gallery access, album roots, and administrator permissions. This is intentionally provider-neutral and has no just-in-time account creation, email matching, or role mapping.

## Configure

1. Create a client at your OIDC provider for authorization-code flow with PKCE S256. Set its exact redirect URI to `https://photos.example.com/api/auth/oidc/callback` (adjust the host and API prefix to your deployment).
2. Create the Photoview users and their album access first. Set administrator permissions in Photoview as usual.
3. Find each provider user's stable `sub` claim (the Keycloak user ID is an example). Put a JSON object mapping each subject to an existing Photoview username in `PHOTOVIEW_OIDC_SUBJECT_MAP`.
4. Set `PHOTOVIEW_OIDC_ISSUER`, `PHOTOVIEW_OIDC_CLIENT_ID`, `PHOTOVIEW_OIDC_REDIRECT_URL`, and `PHOTOVIEW_OIDC_SUCCESS_URL`. Set `PHOTOVIEW_OIDC_CLIENT_SECRET` for a confidential client. See `api/example.env` for an example.
5. Restart Photoview. The login page will offer **Sign in with SSO** only when OIDC is configured.

The issuer and discovered authorization, token, and JWKS URLs must use HTTPS. The callback and success URLs must use HTTPS and the same host, so the existing Photoview session cookie is available to the UI. The callback path must exactly match the API prefix plus `/auth/oidc/callback`. The success URL must be the UI base URL ending in `/` (for example, `https://photos.example.com/` or `https://example.com/photos/`), because failed sign-ins return to its `login` route. It must not contain a query or fragment. OIDC configuration is all-or-nothing: partial configuration prevents startup.

This first integration requires the API and UI to share one browser host. A UI served from a separate hostname is not supported by this cookie flow. If the provider is unreachable during startup, Photoview starts with local password login and hides the SSO button; restart after the provider recovers to retry discovery.

A user's `sub` is scoped to the issuer. If the issuer changes, review every mapping before starting Photoview. Do not map multiple people to one Photoview username. The map is an operator-controlled configuration value; Photoview will reject a login if the mapped local user no longer exists. Removing a subject from the map prevents new OIDC sign-ins for that subject. Existing Photoview sessions remain valid until they expire or are revoked.

## Session behavior and limits

OIDC uses authorization code, PKCE S256, a one-use server-side state record, a browser-bound state cookie, and a nonce checked against the verified ID token. The provider's tokens are discarded after sign-in. Photoview issues its existing 14-day access token and uses the existing cookie-based session, including GraphQL WebSocket authentication. The existing session cookie is accessible to JavaScript because the current UI reads it to authenticate WebSocket connections; a future session redesign could make it HttpOnly. Serve Photoview over HTTPS and keep its normal cross-site scripting protections in place.

Local password sign-in remains available. There is no automatic fallback or silent email linking. OIDC does not alter local admin rights or album permissions. For inspiration, see the optional, explicit-account-linking implementation in [Obsidian MCP](https://github.com/maxkuminov/obsidian-mcp/blob/main/docs/native-oidc.md).

Sign-in errors return to the login page with a short error code. The server logs the reason, including an unmapped OIDC subject when relevant, without logging provider tokens. If a user starts login in two tabs, the second tab replaces the browser state cookie; retry the first tab if it expires.

Login starts are limited to 20 per second per Photoview replica, 300 per minute across replicas, and 1,600 outstanding states across replicas. Excess requests receive HTTP 429 and can be retried later. These shared limits protect database capacity; an attacker can still fill the shared budget, so public deployments should also rate-limit abusive clients at the edge. The shared limits use the Photoview database; if its OIDC state tables are unavailable at startup, local password login remains available and SSO is hidden.
