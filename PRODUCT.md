# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Primary user: a local operator running network diagnostics on their own machine. They use the web UI on localhost to start or stop continuous monitoring, run one-off checks, review recent runs and results, inspect the machine's network environment, and edit configuration.

## Product Purpose

DiagnosticToolkit helps an operator understand network connectivity and diagnose issues by running configured TCP, DNS, and ping checks and presenting their results alongside discovered network environment details. Success means the operator can quickly start a diagnostic run, understand its results, and inspect relevant network context.

## Positioning

The application compares configured checks and resolvers with network details discovered from the host, including its gateway, interfaces, and DNS configuration. The distinct value and broader positioning are open decisions.

## Operating Context

- The web version is a separate Go executable serving a local browser UI, defaulting to localhost.
- Monitoring is started explicitly by the user; the server itself starts idle.
- The existing CLI remains a separate way to run checks.
- Configuration is YAML and covers timeouts, concurrency, thresholds, ping, targets, DNS resolvers/queries, and output paths.

## Capabilities and Constraints

- Preserve the current workflows: start/stop monitoring, run once, inspect live/recent results, review prior runs and their summaries, inspect environment data, and edit/save configuration.
- The dashboard should prioritize starting a diagnostic run.
- The Environment page should make network topology (interfaces, addresses, gateway, resolvers/upstreams, routes) easy to understand.
- The visual theme and page structure are open to redesign.
- The app is intended for localhost use and keeps run history in memory.
- Do not imply that a check succeeded merely because a configured expected outcome was met without making that expectation understandable.

## Evidence on Hand

- Go implementation and YAML configuration in this repository.
- Existing configured TCP, ping, resolver, and DNS-query examples in `netcheck.yaml`.
- No customer, benchmark, or testimonial evidence is established; do not fabricate any.

## Product Principles

- Make the next useful diagnostic action obvious.
- Present results with enough expectation and network context to interpret them correctly.
- Preserve the existing diagnostic capabilities while improving their discoverability and usability.
- Distinguish observed host facts from check results and warnings.
