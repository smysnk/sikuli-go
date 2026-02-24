# sikuligo (Node.js)

SikuliGO is a GoLang implementation of Sikuli visual automation. This package provides the Node.js SDK for launching `sikuligo` locally and executing automation with a small API surface.

## Setup

```bash
cd clients/node
npm install
npm run build
```

## Quickstart

Run:

```bash
cd clients/node
npm run example:workflow:connect
```

`npm run example:workflow:connect` runs:

```js
import { Screen, Pattern } from "../src";

async function main() {
  const screen = await Screen();
  try {
    const match = await screen.click(Pattern("assets/pattern.png").exact());
    console.log(`clicked match target at (${match.targetX}, ${match.targetY})`);
  } finally {
    await screen.close();
  }
}
```

`npm run example:workflow:auto` uses the same primary constructor pattern (`connect -> spawn` fallback handled by `Screen()`):

```bash
cd clients/node
npm run example:workflow:auto
```

`npm run example:workflow:auto` runs:

```js
import { Screen, Pattern } from "../src";

async function main() {
  const screen = await Screen();
  try {
    const match = await screen.click(Pattern("assets/pattern.png").exact());
    console.log(`clicked match target at (${match.targetX}, ${match.targetY})`);
  } finally {
    await screen.close();
  }
}
```

## Run Examples

```bash
cd clients/node
npm run example:workflow:auto
npm run example:workflow:connect
npm run example:find
npm run example:click
npm run example:ocr
npm run example:input
npm run example:app
npm run example:user-flow
npm run doctor
```

## Environment
- `SIKULIGO_BINARY_PATH` (optional explicit path to `sikuligo`)
- `SIKULI_GRPC_ADDR` (optional address used by `start/launch` probe/connect; default probe `127.0.0.1:50051`)
- `SIKULI_GRPC_AUTH_TOKEN` (optional; sent as `x-api-key` for spawned/connected sessions)
- `SIKULI_DEBUG` (optional; set to `1` to log launcher and per-RPC timing details; spawned `sikuligo` logs are shown too)
- `SIKULIGO_SQLITE_PATH` (optional sqlite path for spawned server sessions; default `sikuligo.db`)
- `SIKULI_APP_NAME` (optional; used by `examples/app.js`)

Primary constructors:
- `Screen()` / `Sikuli()` = connect to default address first (1s), else spawn
- `Screen.connect()` / `Sikuli.connect()` = connect only
- `Screen.spawn()` / `Sikuli.spawn()` = spawn only
