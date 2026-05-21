#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const path = require('node:path');

const osMap = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
};

const archMap = {
  x64: 'amd64',
  arm64: 'arm64',
};

const os = osMap[process.platform];
const arch = archMap[process.arch];

function fail(message) {
  console.error(`memforge MCP launcher: ${message}`);
  process.exit(1);
}

if (!os) {
  fail(`unsupported platform ${process.platform}`);
}

if (!arch) {
  fail(`unsupported architecture ${process.arch}`);
}

const pluginRoot = process.env.MEMFORGE_PLUGIN_ROOT || path.resolve(__dirname, '..');
const binaryName = os === 'windows' ? 'memforge.exe' : 'memforge';
const binaryPath = path.join(pluginRoot, 'bin', `${os}-${arch}`, binaryName);
const inheritedPWD = process.env.PWD;
const projectRoot = process.env.MEMFORGE_PROJECT_ROOT ||
  (inheritedPWD && path.isAbsolute(inheritedPWD) && path.resolve(inheritedPWD) !== path.resolve(pluginRoot)
    ? inheritedPWD
    : '');
const args = ['--no-version-check', 'mcp'];

if (projectRoot) {
  args.push('--root', projectRoot);
}

const child = spawn(binaryPath, args, {
  stdio: 'inherit',
  env: process.env,
});

child.on('error', (error) => {
  if (error.code === 'ENOENT') {
    fail(`bundled memforge runtime not found at ${binaryPath}`);
  }
  fail(`failed to launch bundled memforge runtime at ${binaryPath}: ${error.message}`);
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 1);
});
