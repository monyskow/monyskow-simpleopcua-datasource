# Plan: fix(e2e): lift node-opcua connection limit blocking parallel datasource tests (#32)

## Kontekst

`test-servers/node-opcua-server/server.js` runs with node-opcua's default per-endpoint
connection limit (10). Grafana provisions 14 datasources at startup; each fires a health
check on boot and saturates the pool, producing `PREVENTING DDOS ATTACK => maxConnection =10`
in the server log. Per-test datasource isolation from #31 surfaced this: `Save & Test` in
`tests/datasource-config.spec.ts:89` is currently `test.skip` because its connection attempt
is refused. A prior attempt used `maxConnections` as a top-level `OPCUAServer` option — that
key does not exist on `OPCUAServer` (it lives on `OPCUAServerEndPoint`), so it was a no-op.

Issue: https://github.com/MonyskowM/monyskow-simpleopcua-datasource/issues/32

## Podejście

Source-of-truth confirmed by grepping `node-opcua-server@2.123.x` in `node_modules`:

- `node_modules/node-opcua-server/source/opcua_server.ts:185` — `const default_maxConnectionsPerEndpoint = 10;`
- `:837` — documented constructor option `maxConnectionsPerEndpoint?: number` on `OPCUAServerOptions`.
- `:1335` — `this.maxConnectionsPerEndpoint = options.maxConnectionsPerEndpoint || default;`
- `:3847` — value is forwarded to each endpoint as `maxConnections` when the endpoint is built.

The previous attempt used `maxConnections` (wrong key). The correct, documented top-level
option is **`maxConnectionsPerEndpoint`**. One-line change in `server.js`.

## Zmiany w plikach

- `test-servers/node-opcua-server/server.js` — add `maxConnectionsPerEndpoint: 64` inside
  the `new OPCUAServer({ ... })` options block (around line 55-106; near `port`/`hostname`).
  Value 64 gives generous headroom over the AC threshold of 32 and over the 14 provisioned
  datasources × retry bursts, without being absurd. Short comment: `// 14 provisioned DS + per-test isolation; default 10 is too low`.

No other files change. The Dockerfile rebuilds the image on next `docker compose up --build`
and re-runs `npm ci`, so no version bump needed.

## Kolejność

1. Edit `server.js` — add `maxConnectionsPerEndpoint: 64` to the `OPCUAServer` constructor options.
2. Un-skip the test: in `tests/datasource-config.spec.ts:89` change `test.skip(` to `test(`
   and remove the explanatory comment block on lines 84-88.
3. Rebuild + run e2e locally: `docker compose -f docker-compose.e2e.yaml build opcua-server`
   then `npm run e2e`. Grep server logs (`docker compose logs opcua-server`) — must NOT
   contain `PREVENTING DDOS ATTACK`. The server's `MAX CONNECTIONS = 64` debug line (line 964
   in upstream) confirms the option took effect when `DEBUG=*` is set; not required for AC.

## Test plan

- [ ] AC1 (server config ≥ 32 concurrent) — covered by code change; the constant value (64) is the test artifact.
- [ ] AC2 (no DDOS warnings during `npm run e2e`) — manual `docker compose logs opcua-server | grep -c 'PREVENTING DDOS'` must return `0` after a full run.
- [ ] AC3 (un-skip `should test connection successfully`) — Playwright test passes against the patched server.
- [ ] AC4 (`npm run e2e` 52/52 with workers=2) — full local run green; CI matrix already covers anon/userpass/cert variants.
- [ ] Regression guard: the un-skipped `should test connection successfully` test IS the regression guard — if the limit ever regresses to 10, this test fails first.

## Ryzyka / open questions

- **Risk: limit raised but DDOS still emitted.** Means option name is still wrong on the
  installed version. Mitigation: check `package-lock.json` for the resolved `node-opcua-server`
  version and re-grep that exact version's source. The grep above was done against the locally
  installed tree pulled by `npm install`, which matches the Dockerfile's `npm ci` deterministically.
- **Risk: bottleneck is not the per-endpoint limit but lingering Grafana health-check retries.**
  Verification is direct: server logs. If `PREVENTING DDOS` is gone but the un-skipped test still
  flakes, look at Grafana's `datasource health check` interval/timeout rather than the server.
- **Risk: `setInterval` retries from earlier failed health checks keep firing.** Unlikely —
  Grafana health checks are on-demand from the config page in this test, not periodic.
  If observed, increase the value to 128 (still trivial for a test server).
- **ADR needed?** No — single config option on an existing dependency, no new deps, no API change.

## Resolved option name (no ambiguity for `developer`)

```js
const server = new OPCUAServer({
  port: PORT,
  hostname: HOSTNAME,
  maxConnectionsPerEndpoint: 64, // 14 provisioned DS + per-test isolation; default 10 is too low
  // ...rest unchanged
});
```

Source: `node_modules/node-opcua-server/source/opcua_server.ts:837` (option declaration),
`:1335` (assignment), `:3847` (propagation to `OPCUAServerEndPoint.maxConnections`).
