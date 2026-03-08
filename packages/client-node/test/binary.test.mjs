import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";

import { resolveSikuliBinary } from "../dist/src/binary.js";

function withTempDir(fn) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "sikuli-go-node-test-"));
  try {
    return fn(dir);
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

function writeExecutable(filePath, contents) {
  fs.writeFileSync(filePath, contents, { mode: 0o755 });
  fs.chmodSync(filePath, 0o755);
}

test("resolveSikuliBinary rejects explicit wrapper scripts", () => {
  withTempDir((dir) => {
    const wrapper = path.join(dir, "sikuli-go");
    writeExecutable(wrapper, "#!/usr/bin/env node\nconsole.log('wrapper');\n");
    assert.throws(
      () => resolveSikuliBinary(wrapper),
      /does not point to a native sikuli-go runtime binary/
    );
  });
});

test("resolveSikuliBinary accepts explicit native-looking runtime binaries", () => {
  withTempDir((dir) => {
    const binary = path.join(dir, "sikuli-go");
    fs.writeFileSync(binary, Buffer.from([0xcf, 0xfa, 0xed, 0xfe, 0x00, 0x00, 0x00, 0x00]), {
      mode: 0o755
    });
    fs.chmodSync(binary, 0o755);
    assert.equal(resolveSikuliBinary(binary), binary);
  });
});
