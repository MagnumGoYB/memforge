'use strict';

const { spawn: defaultSpawn } = require('node:child_process');
const fs = require('node:fs');
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

function normalizeVersion(version) {
  return String(version || '').trim().replace(/^v/, '');
}

function isVersionCheckDisabled(env) {
  const value = String(env.MEMFORGE_NO_VERSION_CHECK || '').trim().toLowerCase();
  return value === '1' || value === 'true' || value === 'yes';
}

function readPluginVersion(pluginRoot) {
  for (const manifestPath of [
    path.join(pluginRoot, '.claude-plugin', 'plugin.json'),
    path.join(pluginRoot, '.codex-plugin', 'plugin.json'),
  ]) {
    try {
      const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
      if (manifest.version) return normalizeVersion(manifest.version);
    } catch (_) {
      // Try the next host manifest shape.
    }
  }
  return '';
}

async function latestVersion(env) {
  if (env.MEMFORGE_VERSION_CHECK_LATEST) {
    return normalizeVersion(env.MEMFORGE_VERSION_CHECK_LATEST);
  }
  if (typeof fetch !== 'function') return '';
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 750);
  try {
    const url = env.MEMFORGE_VERSION_CHECK_URL || 'https://api.github.com/repos/MagnumGOYB/memforge/releases/latest';
    const response = await fetch(url, { signal: controller.signal });
    if (!response.ok) return '';
    const payload = await response.json();
    return normalizeVersion(payload.tag_name);
  } catch (_) {
    return '';
  } finally {
    clearTimeout(timeout);
  }
}

async function maybeWarnUpdate({ env, pluginRoot, stderr }) {
  if (isVersionCheckDisabled(env)) return;
  const current = readPluginVersion(pluginRoot);
  const latest = await latestVersion(env);
  if (!current || !latest || current === latest) return;
  stderr.write(`MemForge ${latest} is available (current ${current}). Reinstall the plugin package and reload plugins: https://github.com/MagnumGOYB/memforge/releases/tag/v${latest}\n`);
}

function resolveRuntime({ env, platform, arch, dirname }) {
  const os = osMap[platform];
  const mappedArch = archMap[arch];
  if (!os) throw new Error(`unsupported platform ${platform}`);
  if (!mappedArch) throw new Error(`unsupported architecture ${arch}`);
  const pluginRoot = env.MEMFORGE_PLUGIN_ROOT || path.resolve(dirname, '..');
  const binaryName = os === 'windows' ? 'memforge.exe' : 'memforge';
  const binaryPath = path.join(pluginRoot, 'bin', `${os}-${mappedArch}`, binaryName);
  const inheritedPWD = env.PWD;
  const projectRoot = env.MEMFORGE_PROJECT_ROOT ||
    (inheritedPWD && path.isAbsolute(inheritedPWD) && path.resolve(inheritedPWD) !== path.resolve(pluginRoot)
      ? inheritedPWD
      : '');
  const args = ['--no-version-check', 'mcp'];
  if (projectRoot) args.push('--root', projectRoot);
  return { pluginRoot, binaryPath, args };
}

async function main(options = {}) {
  const env = options.env || process.env;
  const platform = options.platform || process.platform;
  const arch = options.arch || process.arch;
  const dirname = options.dirname || __dirname;
  const stderr = options.stderr || process.stderr;
  const spawn = options.spawn || defaultSpawn;

  let runtime;
  try {
    runtime = resolveRuntime({ env, platform, arch, dirname });
  } catch (error) {
    stderr.write(`memforge MCP launcher: ${error.message}\n`);
    return 1;
  }

  await maybeWarnUpdate({ env, pluginRoot: runtime.pluginRoot, stderr });

  return await new Promise((resolve) => {
    const child = spawn(runtime.binaryPath, runtime.args, {
      stdio: 'inherit',
      env,
    });
    child.on('error', (error) => {
      if (error.code === 'ENOENT') {
        stderr.write(`memforge MCP launcher: bundled memforge runtime not found at ${runtime.binaryPath}\n`);
      } else {
        stderr.write(`memforge MCP launcher: failed to launch bundled memforge runtime at ${runtime.binaryPath}: ${error.message}\n`);
      }
      resolve(1);
    });
    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
        return;
      }
      resolve(code ?? 1);
    });
  });
}

module.exports = { main, maybeWarnUpdate, normalizeVersion, resolveRuntime };
