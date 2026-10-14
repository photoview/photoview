---
mode: testing
url: http://localhost:8000/
max_steps: 35
tags: [smoke, albums]
---

# Scan the library and browse albums

## Complete the initial setup
Open http://localhost:8000/.
Type "admin" into the "Username" field, "admin12345" into the "Password" field and "/photos" into the "Photo path" field, then click "Setup Photoview".
Verify the URL contains "/timeline".

## Scan the photo library
Click "Settings" in the sidebar.
Click the "Scan all users" button, then wait 10 seconds for the scan to finish.
Verify the Settings page is still shown.

## Open the Vacation album
Click "Albums" in the sidebar. If the "Vacation" album is not listed yet, open the "photos" album.
Click the "Vacation" album.
Verify the album page shows "Vacation" and 3 photos.
