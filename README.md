# DiagnosticToolkit

`netcheck` is a configurable Go command-line tool for checking basic network connectivity and DNS behavior. It can run a one-off snapshot or repeat checks on an interval, then report results in the terminal and optionally write machine-readable output.

## Checks and reports

- TCP connection checks against configured targets.
- ICMP ping checks with latency and packet-loss thresholds.
- DNS lookups against system or explicitly configured resolvers, including expected-address and expected-NXDOMAIN checks.
- Automatic discovery of the machine's gateway and DNS resolvers for relevant checks.
- One-shot and continuous monitoring modes, with console summaries, optional JSON Lines results, and a JSON run summary.

## Run

Requires Go 1.25 or newer. Review the targets in `netcheck.example.yaml` and run:

```sh
go run ./cmd/netcheck -config ./netcheck.example.yaml
```

Run continuously until interrupted with:

```sh
go run ./cmd/netcheck -config ./netcheck.example.yaml -monitor
```

Useful flags:

- `-config <path>` — YAML configuration file.
- `-output <path>` — append individual results as JSON Lines.
- `-summary <path>` — write the run summary as JSON.

The configuration file documents available targets, resolver and query options, timeouts, concurrency, and thresholds. On Linux, unprivileged ICMP may depend on the host's ping socket policy; netcheck reports a permission hint if a check cannot run.

## Development

```sh
go test ./...
```

The separate `netcheck-web` entry point is experimental and still under development; the CLI is the usable interface at this stage.
