---
test: ../initial_setup_test.md
status: passed
started: 2026-10-14T11:21:29.073Z
duration_s: 91
session_id: 6fa91b7e-8788-45f6-9a69-d145be4bc564
---

# Initial setup — Result

## Open a fresh instance ✓ passed (31.1s)
md5: 826c6f3a09f27260ff1943ab0215dfc8
Open http://localhost:8000/.
Verify the URL contains "/initialSetup" and the page shows "Initial Setup" with "Username", "Password" and "Photo path" fields.

## Complete the setup ✓ passed (53.5s)
md5: 63a841e502a8c6a49c15c22281757e15
Type "admin" into the "Username" field, "admin12345" into the "Password" field and "/photos" into the "Photo path" field, then click "Setup Photoview".
Verify the URL contains "/timeline" and the sidebar shows "Timeline", "Albums", "People" and "Settings".
