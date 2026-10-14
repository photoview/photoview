---
mode: testing
url: http://localhost:8000/
max_steps: 20
tags: [smoke, setup]
---

# Initial setup

## Open a fresh instance
Open http://localhost:8000/.
Verify the URL contains "/initialSetup" and the page shows "Initial Setup" with "Username", "Password" and "Photo path" fields.

## Complete the setup
Type "admin" into the "Username" field, "admin12345" into the "Password" field and "/photos" into the "Photo path" field, then click "Setup Photoview".
Verify the URL contains "/timeline" and the sidebar shows "Timeline", "Albums", "People" and "Settings".
