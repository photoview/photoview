# End-to-end (e2e) tests for web

Browser end-to-end tests for Photoview, authored with [KaneCLI](https://www.testmuai.com/kane-cli/) by TestMu AI. Each test is a plain-English `_test.md` file; KaneCLI runs it in a real Chrome browser and records each step so later runs replay without re-authoring.

| Test | What it covers |
|---|---|
| `initial_setup_test.md` | Completing the initial setup on a fresh instance and landing on the timeline |
| `albums_test.md` | Scanning the library from Settings and opening an album with its photos |

Both tests start from a fresh instance (no database yet). They complement the API and UI unit tests by going through the whole app in a browser.

## Run the tests

The tests expect Photoview at `http://localhost:8000`, using SQLite and a media folder mounted at `/photos` that contains a `Vacation` folder with 3 photos:

```bash
docker build -t photoview-e2e .
docker run -d --name photoview-e2e -p 8000:80 \
  -e PHOTOVIEW_DATABASE_DRIVER=sqlite \
  -e PHOTOVIEW_SQLITE_PATH=/home/photoview/data/photoview.db \
  -e PHOTOVIEW_LISTEN_IP=0.0.0.0 \
  -v "$PWD/e2e-data:/home/photoview/data" \
  -v "/path/to/test/photos:/photos:ro" \
  photoview-e2e
```

Delete `e2e-data/photoview.db` and restart the container before each test so it starts at the initial setup page.

### With KaneCLI

```bash
npm install -g @testmuai/kane-cli
kane-cli login
kane-cli testmd run e2e/kane/initial_setup_test.md --agent
kane-cli testmd run e2e/kane/albums_test.md --agent
```

### With the exported Playwright code

Each test also has an exported Playwright (JavaScript) version under `output-<test>/playwright-javascript-code/`:

```bash
cd e2e/kane/output-albums/playwright-javascript-code
npm install
npx playwright install chromium
node test.js
```

Copy `.env.example` to `.env` and fill in your TestMu AI (LambdaTest) credentials first.

## Editing a test

Edit the step text in the `_test.md` file and run it again with `kane-cli testmd run`. Changed steps (and the steps after them) are re-recorded; unchanged steps replay from `output-<test>/`.
