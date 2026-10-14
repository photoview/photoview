---
test: ../albums_test.md
status: passed
started: 2026-10-14T11:23:52.454Z
duration_s: 183
session_id: af070360-6e76-403e-ac3a-aa621560934a
---

# Scan the library and browse albums — Result

## Complete the initial setup ✓ passed (48.2s)
md5: de6088d750d03f9a8520567be6463c1b
Open http://localhost:8000/.
Type "admin" into the "Username" field, "admin12345" into the "Password" field and "/photos" into the "Photo path" field, then click "Setup Photoview".
Verify the URL contains "/timeline".

## Scan the photo library ✓ passed (61.5s)
md5: e03bf2ffaef3fe9cf1b623f087246227
Click "Settings" in the sidebar.
Click the "Scan all users" button, then wait 10 seconds for the scan to finish.
Verify the Settings page is still shown.

## Open the Vacation album ✓ passed (67.5s)
md5: 65c5c70014cd3d619262929b252d3dc1
Click "Albums" in the sidebar. If the "Vacation" album is not listed yet, open the "photos" album.
Click the "Vacation" album.
Verify the album page shows "Vacation" and 3 photos.
