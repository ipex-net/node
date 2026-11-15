# IPEX node

IPEX (Internet Number Resource Exchange) is a protocol for leasing IPv4 address space between organizations, built as a sovereign Cosmos SDK chain. This repository is the node, `ipexd` (the project's earlier working name was IPEX, which is still used in the code).

**Status: research prototype.** It runs as a local development chain and has broad test coverage, but it is unfinished, has not been audited, and has never run as a public network.

## The problem

IPv4 space ran out years ago. Buying or leasing it now goes through brokers: slow, manual, with little price transparency and no shared record of who holds what or how a range has been used. A lessee can't easily tell whether a subnet is blacklisted, and a lessor can't easily tell whether a lessee will abuse it.

IPEX puts the lease itself on chain, ties every participant to a verified organization, and keeps a public reputation history for both subnets and organizations, fed by oracles that watch routing (RPKI, RIR data) and abuse blocklists.

## Modules

| Module | What it does |
| --- | --- |
| [`x/org`](x/org) | Organizations as on-chain identities: legal details, abuse and contact addresses, member addresses with roles, domain verification. |
| [`x/subnet`](x/subnet) | IPv4 subnets as on-chain assets: owner org, assigned user, geolocation, RIR data, RPKI origin status, active/inactive state. |
| [`x/lease`](x/lease) | Lease contracts and the marketplace: listing, bids, enrolment, renewal, cancellation, safety deposits, revocation periods. |
| [`x/payments`](x/payments) | A single ledger of every value transfer (lease payments, deposits, refunds), written by other modules. |
| [`x/trust`](x/trust) | Reputation scores for subnets, lessees and lessors, built from an append-only blacklist log. |
| [`x/abuse`](x/abuse) | On-chain abuse complaints and their resolution workflow between lessors and lessees. |
| [`x/oracle`](x/oracle) | Brings off-chain data on chain through ABCI++ vote extensions ([`x/oracle/abci`](x/oracle/abci)), so validators agree on it as part of consensus. |
| [`internal/netintel`](internal/netintel) | The off-chain side of the oracle: downloads and caches FireHOL blocklists, country netsets and ASN-DROP, and answers reputation, ASN and geolocation lookups. |

## Running it

Requirements: Go 1.23 and [Ignite CLI](https://docs.ignite.com).

```bash
ignite chain serve
```

This builds `ipexd`, creates a local chain from [`config.yml`](config.yml) with two development accounts (`alice`, `bob`) and starts it. The mnemonics in `config.yml` are development-only keys and control nothing.

The netintel service downloads its blocklists and country data on first start into the node's home directory (`~/.ipex/data/netintel`). Geolocation can also use a MaxMind GeoLite2 Country database at `data/maxmind/GeoLite2-Country.mmdb`, which you supply yourself under MaxMind's licence.

Example transactions:

```bash
ipexd tx org register-org "Example Org" "Example Holdings Ltd" "Hong Kong" "REG1" "HK" "TAX1" \
  "https://example.com" "abuse@example.com" "contact@example.com" --from alice --chain-id ipex -y
ipexd q org list-org
ipexd q lease list-contract
```

## Tests

```bash
go test ./...
```

Timing-sensitive performance tests are behind a build tag, because their limits depend on the machine:

```bash
go test -tags perf ./internal/netintel/ ./x/oracle/keeper/ ./x/payments/keeper/
```

## Licence

[Apache 2.0](LICENSE).
