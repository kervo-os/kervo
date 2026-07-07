#!/usr/bin/env node
// kervo npx wrapper — downloads the GoReleaser binary for this platform on
// first run, verifies it against the release's checksums.txt, caches it per
// version, then execs it. The npm version tracks the binary release tag.
const { spawnSync, execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const crypto = require('node:crypto');

const version = require('../package.json').version;

const goos = { darwin: 'darwin', linux: 'linux', win32: 'windows' }[process.platform];
const goarch = { x64: 'amd64', arm64: 'arm64' }[process.arch];
if (!goos || !goarch || (goos === 'windows' && goarch === 'arm64')) {
  console.error(
    `kervo: no prebuilt binary for ${process.platform}/${process.arch} — ` +
      'see https://github.com/kervo-os/kervo/releases'
  );
  process.exit(1);
}
const exe = goos === 'windows' ? '.exe' : '';

const cacheRoot =
  process.env.KERVO_NPM_CACHE ||
  (goos === 'windows' && process.env.LOCALAPPDATA
    ? path.join(process.env.LOCALAPPDATA, 'kervo-npm')
    : path.join(os.homedir(), '.cache', 'kervo-npm'));
const binDir = path.join(cacheRoot, version);
const bin = path.join(binDir, 'kervo' + exe);

async function download() {
  const archive = `kervo_${version}_${goos}_${goarch}.${goos === 'windows' ? 'zip' : 'tar.gz'}`;
  const base = `https://github.com/kervo-os/kervo/releases/download/v${version}`;
  console.error(`kervo: first run — fetching ${archive} ...`);
  const [res, sums] = await Promise.all([
    fetch(`${base}/${archive}`),
    fetch(`${base}/checksums.txt`),
  ]);
  if (!res.ok) throw new Error(`download failed (${res.status}): ${base}/${archive}`);
  if (!sums.ok) throw new Error(`checksums.txt download failed (${sums.status})`);
  const buf = Buffer.from(await res.arrayBuffer());
  const line = (await sums.text()).split('\n').find((l) => l.trim().endsWith(archive));
  const want = line && line.trim().split(/\s+/)[0];
  const got = crypto.createHash('sha256').update(buf).digest('hex');
  if (!want || got !== want) throw new Error(`checksum mismatch for ${archive}`);

  fs.mkdirSync(binDir, { recursive: true });
  const tmp = path.join(binDir, archive);
  fs.writeFileSync(tmp, buf);
  // bsdtar (macOS, Windows 10+) reads both tar.gz and zip; GNU tar (Linux)
  // gets tar.gz. Extract only the binary member.
  execFileSync('tar', ['-xf', tmp, '-C', binDir, 'kervo' + exe]);
  fs.rmSync(tmp);
  fs.chmodSync(bin, 0o755);
}

(async () => {
  if (!fs.existsSync(bin)) await download();
  const r = spawnSync(bin, process.argv.slice(2), { stdio: 'inherit' });
  process.exit(r.status ?? 1);
})().catch((e) => {
  console.error(`kervo: ${e.message}`);
  process.exit(1);
});
