# Egern Compatibility Status

**Status: Beta — initial iPhone client testing passed (2026-11-04).**

## Verified

- 29 policy groups and 30 remote rule references generated from the shared policy plan
- 30 referenced rule files found locally; both IPv6 variants passed static validation
- Go unit/integration tests and GitHub Actions auto-generation succeeded
- An Egern-native YAML subscription containing 38 nodes was successfully examined
- Initial real-device import and node selection appeared normal
- The `其他节点` (Other Nodes) group was visible and a node could be selected on the tested client

## Still to verify

- Full regional exclusion accuracy for every actual node name, and whether the selected Other Nodes proxy passes traffic
- Remote rule download/update reliability in Egern and correct policy matching
- ADBlock performance with the 206,634-entry rule set
- 42 upstream `PROCESS-NAME` rules excluded during conversion
- Long-running stability, DNS behavior and Wi-Fi/cellular handoff

## Public templates

- `Config/Egern/Egern_config.yaml` — IPv6 enabled
- `Config/Egern/Egern_config_no_ipv6.yaml` — IPv6 disabled

The published files contain the **nonfunctional** placeholder
`https://example.invalid/REPLACE_WITH_YOUR_EGERN_SUBSCRIPTION`.
Replace that URL **locally** with your own Egern-compatible HTTPS subscription;
do not commit a personalized configuration. The template contains no nodes and
cannot connect until its placeholder is replaced.

`其他节点` currently uses a negative-lookahead regular expression. The tested
Egern client displayed the group and allowed selection. This is not a guarantee
that every Egern release or every provider naming scheme supports the same filter.
