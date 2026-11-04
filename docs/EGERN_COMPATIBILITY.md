# Egern Compatibility Status

## Verified

- 29 policy groups preserved
- 30 remote rule references generated
- 30 local rule files found
- IPv4 and IPv6 profiles generated
- Basic DNS configuration generated
- Go unit and integration tests passed

## Pending client verification

- Other Nodes filter uses negative lookahead `(?!...)`
- Actual Egern subscription parsing and node expansion
- Regional automatic selection behavior
- Native rule-set loading
- ADBlock performance with 206,634 entries
- 42 excluded PROCESS-NAME rules
- Full DNS behavior on Wi-Fi and cellular networks

Do not mark the Egern profile production-ready until
the pending client checks have been completed.
