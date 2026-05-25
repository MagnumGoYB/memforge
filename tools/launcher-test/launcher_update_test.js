'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const childProcess = require('node:child_process');
const os = require('node:os');
const path = require('node:path');
const claudeLauncher = require('../../plugins/claude-code/memforge/bin/memforge-mcp-launcher-lib.js');
const codexLauncher = require('../../plugins/codex/memforge/bin/memforge-mcp-launcher-lib.js');

async function testWarnsWhenLatestIsNewer() {
  const stderr = [];
  const spawned = [];
  const exitCode = await claudeLauncher.main({
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
  await claudeLauncher.main({
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

async function testCodexDownloadsMissingRuntime() {
  const pluginRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'memforge-codex-launcher-'));
  fs.mkdirSync(path.join(pluginRoot, '.codex-plugin'), { recursive: true });
  fs.writeFileSync(path.join(pluginRoot, '.codex-plugin', 'plugin.json'), JSON.stringify({ version: '1.2.3' }));
  const spawned = [];
  const exitCode = await codexLauncher.main({
    env: {
      MEMFORGE_PLUGIN_ROOT: pluginRoot,
      MEMFORGE_NO_VERSION_CHECK: '1',
    },
    platform: 'linux',
    arch: 'x64',
    stderr: { write: () => {} },
    fetch: async (url) => {
      assert.match(url, /\/releases\/download\/v1\.2\.3\/memforge-linux-amd64$/);
      return {
        ok: true,
        arrayBuffer: async () => Buffer.from('#!/bin/sh\nexit 0\n'),
      };
    },
    spawn: (binary, args) => {
      spawned.push({ binary, args });
      assert.equal(fs.existsSync(binary), true);
      assert.equal(fs.readFileSync(binary, 'utf8'), '#!/bin/sh\nexit 0\n');
      return { on(event, callback) { if (event === 'exit') callback(0, null); } };
    },
  });
  assert.equal(exitCode, 0);
  assert.equal(spawned.length, 1);
  assert.match(spawned[0].binary, /bin\/linux-amd64\/memforge$/);
}

function testShellWrapperFindsExplicitNodePath() {
  const pluginRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'memforge-wrapper-'));
  const launcherPath = path.join(pluginRoot, 'memforge-mcp-launcher.js');
  const wrapperPath = path.join(pluginRoot, 'run-memforge-mcp-launcher.sh');
  fs.writeFileSync(launcherPath, 'process.stdout.write("wrapper-ok\\n");\n');
  fs.writeFileSync(wrapperPath, fs.readFileSync(path.join('plugins', 'codex', 'memforge', 'bin', 'run-memforge-mcp-launcher.sh'), 'utf8'), { mode: 0o755 });
  const result = childProcess.spawnSync(wrapperPath, {
    cwd: pluginRoot,
    env: { PATH: '/usr/bin:/bin:/usr/sbin:/sbin' },
    encoding: 'utf8',
  });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, 'wrapper-ok\n');
}

(async () => {
  await testWarnsWhenLatestIsNewer();
  await testNoWarnWhenDisabled();
  await testCodexDownloadsMissingRuntime();
  testShellWrapperFindsExplicitNodePath();
})();
