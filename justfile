# netcheck — common tasks
#
# `just` with no arguments lists everything.

binary := "netcheck"
outdir := "bin"
distdir := "dist"
pkg := "./cmd/netcheck"
pkgweb := "./cmd/netcheck-web"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

default:
    @just --list

# --- build ---------------------------------------------------------------

# Build for the current platform.
build:
    go build -o {{ outdir }}/{{ binary }} {{ pkg }}

# Build for Linux (amd64).
build-linux:
    GOOS=linux GOARCH=amd64 go build -o {{ outdir }}/{{ binary }}-linux-amd64 {{ pkg }}

# Build for Windows (amd64).
build-windows:
    GOOS=windows GOARCH=amd64 go build -o {{ outdir }}/{{ binary }}-windows-amd64.exe {{ pkg }}

# Build for macOS (arm64).
build-darwin:
    GOOS=darwin GOARCH=arm64 go build -o {{ outdir }}/{{ binary }}-darwin-arm64 {{ pkg }}

# Build every target.
build-all: build-linux build-windows build-darwin

# Build with the race detector. Use this while working on the runner.
build-race:
    go build -race -o {{ outdir }}/{{ binary }}-race {{ pkg }}

# Build every platform and package each with the example config.
package: clean-dist build-all
    #!/usr/bin/env bash
    set -euo pipefail

    for target in linux-amd64 darwin-arm64 windows-amd64; do
        name="{{ binary }}-{{ version }}-${target}"
        staging="{{ distdir }}/${name}"
        mkdir -p "${staging}"

        if [ "${target}" = "windows-amd64" ]; then
            cp "{{ outdir }}/{{ binary }}-${target}.exe" "${staging}/{{ binary }}.exe"
        else
            cp "{{ outdir }}/{{ binary }}-${target}" "${staging}/{{ binary }}"
        fi

        cp netcheck.example.yaml "${staging}/"
        [ -f README.md ] && cp README.md "${staging}/" || true

        (cd {{ distdir }} && zip -qr "${name}.zip" "${name}")
        rm -rf "${staging}"
        echo "packaged ${name}.zip"
    done

clean-dist:
    rm -rf {{ distdir }}

# --- run -----------------------------------------------------------------
# Run without building a binary. Note: -config defaults to a path next to

# the executable, which under `go run` is a temp dir — so pass it explicitly.
run *ARGS:
    go run {{ pkg }} -config ./netcheck.yaml {{ ARGS }}

run-web *ARGS:
    go run {{ pkgweb }} -config ./netcheck.yaml {{ ARGS }}

# Run continuously until interrupted.
run-monitor *ARGS:
    go run {{ pkg }} -config ./netcheck.yaml -monitor {{ ARGS }}

# Build, then run the built binary (picks up netcheck.yaml beside it).
start *ARGS: build
    ./{{ outdir }}/{{ binary }} {{ ARGS }}

# Build, then run in monitor mode.
start-monitor *ARGS: build
    ./{{ outdir }}/{{ binary }} -monitor {{ ARGS }}

# Run with results and summary written to ./out/.
run-logged *ARGS:
    go run {{ pkg }} -config ./netcheck.yaml \
        -output ./out/results.jsonl \
        -summary ./out/summary.json {{ ARGS }}

# Monitor with file output. The usual way to capture an intermittent fault.
monitor-logged *ARGS:
    go run {{ pkg }} -config ./netcheck.yaml -monitor \
        -output ./out/results.jsonl \
        -summary ./out/summary.json {{ ARGS }}

# Run under the race detector.
run-race *ARGS: build-race
    ./{{ outdir }}/{{ binary }}-race -config ./netcheck.yaml {{ ARGS }}

# Monitor under the race detector — the shutdown path has the most moving

# parts in the codebase, so interrupt this at a few different points.
monitor-race *ARGS: build-race
    ./{{ outdir }}/{{ binary }}-race -config ./netcheck.yaml -monitor {{ ARGS }}

# --- checks --------------------------------------------------------------

vet:
    go vet ./...

test:
    go test ./...

fmt:
    go fmt ./...

check-cross:
    GOOS=linux   go build ./...
    GOOS=windows go build ./...
    GOOS=darwin  go build ./...

# Everything CI would run.
ci: fmt vet test check-cross

# --- housekeeping --------------------------------------------------------

clean:
    rm -rf {{ outdir }} {{ distdir }} out

tidy:
    go mod tidy

# Create a local netcheck.yaml from the example, if one does not exist.
init-config:
    #!/usr/bin/env bash
    if [ -f netcheck.yaml ]; then
        echo "netcheck.yaml already exists — leaving it alone"
    else
        cp netcheck.example.yaml netcheck.yaml
        echo "created netcheck.yaml"
    fi
