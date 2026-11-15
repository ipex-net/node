<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/assets/ipex-logo-dark.png">
    <img alt="IPEX Network" src=".github/assets/ipex-logo.png" width="320">
  </picture>
</p>

<p align="center"><b>An open protocol for exchanging Internet Number Resources, starting with IPv4 leasing.</b></p>

---

This repository is the IPEX node, `ipexd`: an application-specific blockchain built with the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk). IPEX was the project's working name, and it's still what the code calls itself.

> **Status: research prototype.** The node runs as a local development chain and every module has tests, but nothing here has been audited or run as a public network. The protocol design, the consensus setup and the economics are still changing.

## Why

IANA's IPv4 pool ran out in 2011, and every Regional Internet Registry (RIR) has run out since. Organizations that need addresses now mostly lease them, and leasing is still handled by brokers and private agreements:

- **Trust runs both ways.** The lessor has to trust that the lessee won't burn the range with spam or attacks, which gets it blacklisted and lowers its value. The lessee has to trust that the lessor will authorise the route and won't pull it early.
- **Routing is set up by hand.** A lessee can only announce a prefix if the holder publishes an RPKI Route Origin Authorization (ROA) or signs a Letter of Authorization. That's a manual step and it often goes wrong.
- **Nobody checks geolocation.** A lessee can say a subnet is deployed in one country and use it in another. Checking that claim takes traceroute work most lessors can't do.
- **Prices are opaque.** There is no shared order book, and the intermediaries decide who gets access and on what terms.

IPEX replaces the broker with a protocol. Leases, deposits, reputations and abuse reports live on chain. Validators act as an oracle network: they check routing, ownership, blocklists and location against the real internet, and record only results that a majority of them agree on.

## What the node does

| Area | Module | Implemented in this prototype |
| --- | --- | --- |
| **Identity** | [`x/org`](x/org) | Organizations register with legal details and abuse and contact addresses, and their members have roles. Validators verify the org's website domain by DNS. They confirm ownership of a subnet by matching the RIR abuse mailbox against that domain. |
| **Assets** | [`x/subnet`](x/subnet) | IPv4 subnets as on-chain assets. Each one is synced with its RIR record over RDAP and tracks owner org, assigned user, origin ASN, RPKI validity and geolocation. |
| **Marketplace** | [`x/lease`](x/lease) | Lessors list subnets with a price, safety deposit, revocation period and their own rules. A lease is either fixed price, where the first qualifying bidder wins, or by bid, where the lessor picks. Contracts can be renewed, cancelled, terminated or revoked. Safety deposits build up across renewals instead of being paid upfront. |
| **Settlement** | [`x/payments`](x/payments) | One ledger records every value transfer (lease payments, deposits, refunds, compensation), written by the other modules. |
| **Reputation** | [`x/trust`](x/trust) | Separate reputation scores for subnets, leases, lessees and lessors, built from an append-only blacklist log and the abuse history. |
| **Abuse** | [`x/abuse`](x/abuse) | Structured complaints by abuse type, a threaded on-chain conversation between lessor and lessee, evidence kept on IPFS with only its hash on chain, and resolution deadlines enforced at the end of each block. The deadlines are set to minutes for development. |
| **Oracle** | [`x/oracle`](x/oracle) | Validators run the checks below off chain and agree on the results through ABCI++ vote extensions ([`x/oracle/abci`](x/oracle/abci)). The proposer tallies the votes by voting power and commits only the results a majority agrees on. |
| **Network intelligence** | [`internal/netintel`](internal/netintel) | The service each validator runs. It loads the [FireHOL](https://github.com/firehol/blocklist-ipsets) level 1–4 blocklists, the ASN-DROP list and several country datasets into an in-memory trie ([cidranger](https://github.com/yl2chen/cidranger)). When new data arrives, it builds a fresh set of lookup tables and swaps them in atomically, so lookups never wait on an update. |

The oracle checks:

- **RPKI:** is this ASN authorised to originate this prefix? Asked of a local [Routinator](https://github.com/NLnetLabs/routinator) instance, with RIPE's validator as a fallback.
- **RIR data:** owner, allocation and abuse contact, fetched over RDAP from the responsible registry.
- **Reputation:** is any address in the subnet on a tracked blocklist, and is its origin ASN on ASN-DROP?
- **Geolocation:** does the claimed country agree with the geolocation datasets? When a lessee moves a subnet, a traceroute checks where the second- or third-to-last hop sits.

### Designed, not built yet

The design also covers:

- a public [RFC 8805](https://www.rfc-editor.org/rfc/rfc8805) geofeed of verified locations
- having the lessor's automation issue ROAs when a lease starts
- governance over protocol parameters
- support for IPv6 and for ASN transfers
- the network's economics, still being designed

## Repository layout

```
app/              chain wiring: modules, ABCI++ handlers, netintel startup
cmd/ipexd/       node binary
proto/ipex/      protobuf definitions for every module
x/<module>/       keepers, message handlers, queries, EndBlock logic, tests
internal/netintel the off-chain intelligence service the validators run
```

## Running it

You need Go 1.23 and [Ignite CLI](https://docs.ignite.com).

```bash
ignite chain serve
```

This builds `ipexd`, starts a local single-validator chain from [`config.yml`](config.yml) and creates two development accounts, `alice` and `bob`. Their mnemonics are development keys and hold nothing of value.

The node downloads its blocklists and country data on first start into `~/.ipex/data/netintel`. The RPKI check expects a Routinator instance on `127.0.0.1:1337` and falls back to RIPE's public validator. Geolocation checks run the system `traceroute`.

A few things to try:

```bash
ipexd tx org register-org "Example Org" "Example Holdings Ltd" "Hong Kong" "REG1" "HK" "TAX1" \
  "https://example.com" "abuse@example.com" "contact@example.com" --from alice --chain-id ipex -y
ipexd q org list-org
ipexd q lease list-contract
ipexd tx lease --help
```

## Tests

```bash
go test ./...
```

The `internal/netintel` tests download live data, so that package takes several minutes. Tests with hard timing limits sit behind a build tag:

```bash
go test -tags perf ./internal/netintel/ ./x/oracle/keeper/ ./x/payments/keeper/
```

## Licence

[Apache 2.0](LICENSE)
