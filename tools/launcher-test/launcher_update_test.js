'use strict';

const assert = require('node:assert/strict');
const path = require('node:path');
const launcher = require('../../plugins/claude-code/memforge/bin/memforge-mcp-launcher-lib.js');

async function testWarnsWhenLatestIsNewer() {
  const stderr = [];
  const spawned = [];
  const exitCode = await launcher.main({
    env: {
      MEMFORGE_PLUGIN_ROOT: path.resolve('plugins/claude-code/memforge'),
      MEMFORGE_VERSION_CHECK_LATEST: 'v9.9.9',
    },
    platform: 'linux',
    arch: 'x64',
    stderr: { write: (text) => stderr.push(text) },
    spawn: (binary, args) => {
      spawned.push({ binary, args });
      return { on(event, callback) { if (event === 'exit') callback(0, null); } };
    },
  });
  assert.equal(exitCode, 0);
  assert.equal(spawned.length, 1);
  assert.match(stderr.join(''), /MemForge 9\.9\.9 is available/);
}

async function testNoWarnWhenDisabled() {
  const stderr = [];
  await launcher.main({
    env: {
      MEMFORGE_PLUGIN_ROOT: path.resolve('plugins/claude-code/memforge'),
      MEMFORGE_VERSION_CHECK_LATEST: 'v9.9.9',
      MEMFORGE_NO_VERSION_CHECK: '1',
    },
    platform: 'linux',
    arch: 'x64',
    stderr: { write: (text) => stderr.push(text) },
    spawn: () => ({ on(event, callback) { if (event === 'exit') callback(0, null); } }),
  });
  assert.equal(stderr.join(''), '');
}

(async () => {
  await testWarnsWhenLatestIsNewer();
  await testNoWarnWhenDisabled();
})();
