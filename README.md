# monitoring-agent

A small, standalone Go agent that reports CPU, memory, disk, and Docker
container statistics from a Linux host to a remote HTTP endpoint once a
minute. It has no opinion about what happens to that data — it only collects
and reports it; what's done with that data is entirely up to whatever's on
the other end of the report URL.

Metric collection itself (reading `/proc/stat`, `/proc/meminfo`, disk usage,
and `docker stats`) lives in the separate
[`servermetrics`](https://github.com/aliasproject/servermetrics) package,
which this agent just calls on a timer and reports.

## Usage

```bash
alios-monitor <report-url>
```

Every minute, the agent collects a fresh snapshot and sends it as a JSON POST
to `<report-url>`. A non-200 response is logged; the agent keeps running and
tries again on the next tick.

Check the installed version:

```bash
alios-monitor --version
# or
alios-monitor -v
```

### Example payload

```json
{
  "version": "v1.1.0",
  "cpu": { "usedpct": 12.4, "...": "..." },
  "memory": {
    "total": 2048000,
    "available": 512000,
    "used": 1536000,
    "usedpct": 75.0,
    "swap_total": 4194304,
    "swap_used": 0,
    "swap_usedpct": 0
  },
  "disk": { "total": 20000000000, "free": 8000000000, "used": 12000000000, "usedpct": 60.0 },
  "containers": [ { "container_name": "web", "cpu_pct": 3.2, "...": "..." } ],
  "timestamp": 1735689600
}
```

See [`servermetrics`](https://github.com/aliasproject/servermetrics#api-reference)
for the full field reference on `cpu`, `memory`, `disk`, and `containers`.

## Building

```bash
go build -o alios-monitor .
```

To bake in a version string (what `--version` reports and what's sent in
every metric payload) at build time:

```bash
go build -ldflags "-X main.Version=v1.1.0" -o alios-monitor .
```

Without that flag, the binary reports its version as `dev`.

## Requirements

- Linux (reads `/proc/stat` and `/proc/meminfo` directly)
- Docker installed and on `PATH`, if you want container stats — the agent
  runs fine without it, just reports an empty container list
- Go 1.23 or later, to build from source

## Releases

Tagged versions (`vX.Y.Z`) are built and published automatically — see
[`.github/workflows/release.yml`](.github/workflows/release.yml).

## License

MIT — see [LICENSE](LICENSE).
