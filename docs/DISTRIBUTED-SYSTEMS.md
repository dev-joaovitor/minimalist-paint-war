# Distributed-Systems Characteristics — Minimalist Paint War

## What it is

Minimalist Paint War is a real-time multiplayer browser game built as a
**centralized, authoritative client-server system**. Multiple browser clients
connect over the network to a single Go server. The server runs the entire game
simulation and is the single source of truth; the browsers only send player
input and draw what the server tells them. Match results are persisted to a
PostgreSQL database.

## Architecture

```
   [Browser A]  [Browser B]  [Browser C]      <- clients: input + rendering
       |            |            |
       |   WebSocket (JSON, persistent, bidirectional)
       \____________|____________/
                    |
              [ Go Server ]                    <- single room, authoritative state
              - actor game loop (20 Hz)
              - validates input, runs physics
              - broadcasts world snapshots
                    |
              [ PostgreSQL ]                    <- async match/result persistence
```

## Distributed-systems properties

### 1. Authoritative server / single source of truth
All game state (players, projectiles, flags, scores) lives on the server and is
mutated only there. Clients cannot change state directly — they send *intent*
(move/shoot), and the server validates it (cooldowns, collisions, spawn
protection) before applying. This prevents cheating and keeps every client
consistent. *(`internal/hub/hub.go`, `internal/game/world.go`)*

### 2. Communication model
Clients and server communicate over a **persistent WebSocket** connection using
a typed JSON envelope protocol: `{ "type": "...", "data": {...} }`. Messages are
asynchronous and bidirectional — client sends `join`/`input`/`start_game`;
server sends `lobby_state`/`match_start`/`state` (snapshots)/`scoreboard`. This
is a message-passing distributed system, not shared memory.
*(`internal/ws/protocol.go`, `internal/ws/client.go`)*

### 3. Concurrency: the actor model
The server avoids locks entirely. A **single goroutine** (`Hub.run()`) owns all
game state. Every other goroutine — one read pump and one write pump per
connected client — communicates with it by sending messages onto a single
**command channel** (buffered at 256). Because only one goroutine ever touches
the state, there are no data races and no mutexes. This is the classic *actor
pattern*: serialize concurrent access through a single mailbox.
*(`internal/hub/hub.go`, `internal/ws/client.go`)*

### 4. State synchronization
The simulation advances on a **fixed tick of 20 Hz** (one step every 50 ms). On
each tick the server computes a **full world snapshot** and broadcasts the same
encoded bytes to every connected client. Clients store each player's *last
input* and re-apply it every tick, so clients only transmit on change
(edge-triggered), reducing traffic. There is **no client-side prediction or
interpolation** — clients render exactly what the server sends, so the design is
simple and consistent at the cost of showing one round-trip of latency.
*(`internal/game/loop.go`, `internal/game/snapshot.go`)*

### 5. Fault tolerance / partial failure
Distributed systems must survive nodes failing independently. Handling here:
- **Client reconnect:** exponential backoff, 500 ms up to 5 s. *(`client/src/lib/net/socket.svelte.js`)*
- **Join timeout:** a connection that does not complete the join handshake within 10 s is closed. *(`internal/ws/handler.go`)*
- **Slow-client protection:** if a client's send buffer (32) fills, that client is dropped rather than stalling the server. *(`internal/ws/client.go`)*
- **Leader re-election:** the lobby leader role is reassigned if the leader disconnects. *(`internal/hub/hub.go`)*
- **Graceful degradation:** if active players drop below 2 mid-match, the match aborts back to lobby; an empty room resets to a clean state. *(`internal/hub/hub.go`, `internal/hub/room.go`)*
- **Async persistence:** database writes go through a separate background worker (queue of 64, 5 s timeout) so a slow DB never blocks the game loop. *(`internal/persist/writer.go`)*

### 6. Consistency & timing model
The game is a **deterministic, fixed-tick simulation**: all clients in a match
observe the same authoritative timeline. Real-time consistency is bounded by
network RTT (clients lag the server by their latency). Persistence to PostgreSQL
is **eventually consistent** — it happens asynchronously, off the critical path,
after the in-memory authoritative state has already updated.

## Key parameters

| Parameter | Value | Source |
|---|---|---|
| Tick rate (simulation) | 20 Hz (50 ms) | `internal/game/constants.go` |
| Snapshot broadcast rate | 20 Hz | `internal/hub/hub.go` |
| Min players to start | 2 | `internal/hub/room.go` |
| Match duration | 120 s (configurable) | `server/main.go` |
| Spawn protection | 2000 ms | `internal/game/constants.go` |
| Respawn delay | 3000 ms | `internal/game/constants.go` |
| Hub command buffer | 256 | `internal/hub/hub.go` |
| Per-client send buffer | 32 | `internal/ws/client.go` |
| Persistence queue | 64 (5 s write timeout) | `internal/persist/writer.go` |
| Join timeout | 10 s | `internal/ws/handler.go` |
| Reconnect backoff | 500 ms -> 5 s | `client/src/lib/net/socket.svelte.js` |

## Tech stack

- **Server:** Go 1.26, `coder/websocket`, `pgx` (PostgreSQL driver)
- **Client:** Svelte 5 / SvelteKit, HTML5 Canvas 2D
- **Storage:** PostgreSQL 16
- **Deployment:** Docker Compose (server, client, postgres on one bridge network)

## One-line summaries (for quick recall)

1. **Authoritative server** — server owns all state; clients send input only.
2. **Message passing** — persistent WebSocket, typed JSON messages, async/bidirectional.
3. **Actor concurrency** — one goroutine owns state; clients talk to it via a command channel; no locks.
4. **Synchronization** — fixed 20 Hz tick, full snapshot broadcast, no client prediction.
5. **Fault tolerance** — reconnect with backoff, timeouts, slow-client drop, leader re-election, match abort, async DB.
6. **Consistency** — deterministic fixed-tick timeline; latency-bounded; eventual consistency to the database.
