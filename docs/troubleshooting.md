# Troubleshooting

Known issues, quirks, and workarounds for the Simple OPC-UA datasource.

## Known Limitation: ProSys Simulation Server (Free Edition) + Secure Channels

When connecting to ProSys with Sign or SignAndEncrypt, see issue #68 and upstream gopcua/opcua#698. The workaround for the cert KeyUsage issue is now fixed in main; if you still see `EOF`, ensure your Generated Certificate was created from a version of this plugin >= v1.0.1. For Certificate-as-Auth, ProSys requires manual user mapping in its Users tab.

## Connection EOF — Generic

OPC-UA secure channel failures often surface as `connect: connect: EOF` because gopcua swallows the underlying status code (gopcua/opcua#698). Verify trust list and cert KeyUsage compliance (Part 6 §6.2.2).

## Grafana 13 Quirks

These are non-obvious behaviors discovered during testing against Grafana 13. Do not re-discover them.

### UID-based Resource URLs Only

G13 returns 404 on `/api/datasources/{numericId}/resources/...`. Use `/api/datasources/uid/{uid}/resources/...` (supported on G9+, so no version branching needed).

### "What's New" Splash Modal

Blocks every authenticated page on first load in G13.0.1+. `GF_FEATURE_TOGGLES_SPLASHSCREEN=false` does NOT disable it (`AllowSelfServe:false`). `tests/auth.setup.ts` dismisses it by clicking close with a soft 5s timeout — fine if the modal is not there.

### Dashboard Panel Selectors Changed

Use role-scoped locators, not text-matched OR-chains.
