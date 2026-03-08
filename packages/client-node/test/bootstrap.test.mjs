import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";

import { ensureSikuliGoOnPath } from "../examples/bootstrap.mjs";

function withTempDir(fn) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "sikuli-go-bootstrap-test-"));
  const prevCwd = process.cwd();
  const prevPath = process.env.PATH;
  const prevBinary = process.env.SIKULI_GO_BINARY_PATH;
  try {
    process.chdir(dir);
    return fn(dir, prevPath);
  } finally {
    process.chdir(prevCwd);
    process.env.PATH = prevPath;
    if (prevBinary === undefined) {
      delete process.env.SIKULI_GO_BINARY_PATH;
    } else {
      process.env.SIKULI_GO_BINARY_PATH = prevBinary;
    }
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

function writeExecutable(filePath, contents) {
  fs.writeFileSync(filePath, contents, { mode: 0o755 });
  fs.chmodSync(filePath, 0o755);
}

test("ensureSikuliGoOnPath ignores PATH wrapper shims", () => {
  withTempDir((dir, prevPath) => {
    const fakeBinDir = path.join(dir, "fake-bin");
    fs.mkdirSync(fakeBinDir, { recursive: true });
    const wrapperPath = path.join(fakeBinDir, process.platform === "win32" ? "sikuli-go.exe" : "sikuli-go");
    writeExecutable(wrapperPath, "#!/usr/bin/env node\nconsole.log('wrapper');\n");

    const sourceBinary = path.join(dir, process.platform === "win32" ? "source.exe" : "source");
    fs.writeFileSync(sourceBinary, Buffer.from([0xcf, 0xfa, 0xed, 0xfe, 0x00, 0x00, 0x00, 0x00]), {
      mode: 0o755
    });
    fs.chmodSync(sourceBinary, 0o755);

    process.env.PATH = `${fakeBinDir}${path.delimiter}${prevPath ?? ""}`;
    process.env.SIKULI_GO_BINARY_PATH = sourceBinary;

    const installed = ensureSikuliGoOnPath();
    assert.notEqual(installed, wrapperPath);
    assert.match(
      installed,
      new RegExp(`\\.sikuli-go[\\\\/]bin[\\\\/]${process.platform === "win32" ? "sikuli-go\\.exe" : "sikuli-go"}$`)
    );
    assert.equal(process.env.SIKULI_GO_BINARY_PATH, installed);
    assert.deepEqual(fs.readFileSync(installed), fs.readFileSync(sourceBinary));
  });
});
