# API Electron App (macOS)

Desktop shell for the sikuli-go dashboard/session viewer with API process control.

## Run

```bash
yarn workspace @sikuligo/api-electron dev
```

Environment overrides:

- `SIKULI_GO_BINARY_PATH` (default: `../../../sikuli-go` relative to this package)
- `SIKULI_GO_API_LISTEN` (default: `127.0.0.1:50051`)
- `SIKULI_GO_ADMIN_LISTEN` (default: `127.0.0.1:8080`)
- `SIKULI_GO_API_AUTO_START` (default: `1`; set `0` to disable auto-start)
- `SIKULI_GO_DASHBOARD_URL` (default: `http://127.0.0.1:8080/dashboard`)
- `SIKULI_GO_SESSION_VIEW_URL` (default: `${SIKULI_GO_DASHBOARD_URL}?view=session-viewer`)
