#!/usr/bin/env node
// kervo npx wrapper — v0.0.x ships instructions; binary download lands with
// the first tagged release (GoReleaser artifacts).
const msg = `
kervo — deterministic context for non-deterministic agents

This npm package is the future npx wrapper for the kervo binary.
Until the first tagged release, install from source:

  git clone https://github.com/kervo-os/kervo && cd kervo
  make build          # requires Go
  ./kervo init

Docs & source: https://github.com/kervo-os/kervo
`;
console.log(msg.trim());
process.exit(0);
