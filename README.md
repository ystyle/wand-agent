# wand-agent

WebSocket PTY agent for [wand](https://github.com/ystyle/wand) — a HarmonyOS terminal emulator.

Creates a PTY session for each WebSocket connection, forwarding terminal I/O between the HarmonyOS app and a shell running in a openEuler container.

## Quick Start

```bash
go build -o wand-agent .
./wand-agent --token harmonyterm
```

Default listen address: `:8765`

## Usage

```
./wand-agent [--addr :8765] [--token <token>]
```

- `--addr` — listen address (default `:8765`)
- `--token` — auth token (auto-generated if empty, printed to stderr on start)

## Protocol

WebSocket endpoint: `/ws?token=<token>&cols=80&rows=24&cwd=/path`

- **Binary frames** (client→server): terminal input (keyboard data)
- **Binary frames** (server→client): PTY output (ANSI/VT sequence)
- **Text frames** (client→server): JSON control messages

### Control Messages

| Type | Direction | Description |
|------|-----------|-------------|
| `{"type":"resize","cols":80,"rows":24}` | client→server | Resize PTY |
| `{"type":"cwd"}` | client→server | Query working directory |
| `{"type":"cwd","dir":"/path"}` | server→client | Working directory response |
| `{"type":"fork","cwd":"/path"}` | client→server | Fork new session at path |
| `{"type":"forked","id":"..."}` | server→client | New session ID |
| `{"type":"ping","ts":123}` | bidirectional | Heartbeat |
| `{"type":"error","error":"..."}` | server→client | Error notification |

## Build

```bash
go build -buildvcs=false -o wand-agent .
```

## License

MIT
