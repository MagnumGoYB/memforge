#!/usr/bin/env node
'use strict';

const launcher = require('./memforge-mcp-launcher-lib.js');

launcher.main().then((code) => {
  process.exit(code);
});
