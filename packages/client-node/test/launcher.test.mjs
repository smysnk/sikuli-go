import test from "node:test";
import assert from "node:assert/strict";

import { buildServerArgs } from "../dist/src/launcher.js";

test("buildServerArgs encodes empty admin listen as a single flag assignment", () => {
  const args = buildServerArgs({
    address: "127.0.0.1:50051",
    adminListen: "",
    authToken: "token",
    sqlitePath: "sikuli-go.db"
  });
  assert.deepEqual(args, [
    "-listen",
    "127.0.0.1:50051",
    "-admin-listen=",
    "-auth-token",
    "token",
    "-enable-reflection=false",
    "-sqlite-path",
    "sikuli-go.db"
  ]);
  assert.equal(args.includes(""), false);
});
