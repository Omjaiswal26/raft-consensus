# Raft Consensus Simulation

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-Omjaiswal26%2Fraft--consensus-181717?logo=github)](https://github.com/Omjaiswal26/raft-consensus)

A hands-on Go implementation of the [Raft consensus algorithm](https://raft.github.io/raft.pdf), built to understand the protocol in depth by constructing it piece by piece—not by wrapping an existing library.

**Repository:** [github.com/Omjaiswal26/raft-consensus](https://github.com/Omjaiswal26/raft-consensus)

The goal is a **runnable, inspectable cluster**: leader election, heartbeats, log replication, and HTTP/WebSocket APIs for observing cluster state in real time (with a frontend planned next).

---

## Why Raft?

Raft keeps a replicated log consistent across machines so that a majority agree on the same sequence of commands, even when nodes crash or the network partitions.

Compared to Paxos, Raft emphasizes **understandability**: clear roles (follower / candidate / leader), explicit terms, and two primary RPCs (`RequestVote`, `AppendEntries`).

This project walks through those mechanics in code:

| Concept | What you see in this repo |
|---------|---------------------------|
| Randomized election timeouts | `internal/raft/timer.go` |
| Leader election | `BecomeCandidate` → `RequestVote` → majority → `BecomeLeader` |
| Heartbeats | Empty `AppendEntries` on a ticker |
| Log replication | `SubmitCommand` + non-empty `AppendEntries` |
| Safety under concurrency | Per-node mutex; snapshots for external readers |

---

## What's implemented

### Consensus core (`internal/raft`)

- **In-process cluster** — multiple `raft.Node`s in one process (simulation-friendly)
- **State machine roles** — follower → candidate → leader
- **Randomized election timeout** — 150–300 ms
- **RequestVote RPC** — term checks, one vote per term, log up-to-date rule (logs empty-friendly today)
- **AppendEntries RPC** — heartbeats + entry replication + basic commit index advance on followers
- **Leader heartbeats** — ~50 ms interval (must stay well below election timeout)
- **Client write path** — `SubmitCommand` on the leader; majority replication → `commitIndex`
- **Peer topology** — persisted as `PeerIDs []uint`; runtime peers wired via `WireRuntimePeers`
- **Safe reads** — `Node.Snapshot()` under lock for APIs

### HTTP / realtime (`internal/api`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/cluster` | JSON snapshot of all nodes |
| `GET` | `/ws` | WebSocket stream of cluster snapshots (~300 ms) |

Layering: **router → handler → service → raft** (handlers never lock Raft mutexes).

### Persistence scaffolding

- GORM + PostgreSQL models (`models.RaftNode`, `internal/store`) for future durable membership
- Runtime simulation currently runs **in memory** from `cmd/main.go`

### Not done yet (roadmap)

- Per-follower `nextIndex` / `matchIndex` maps (full paper Fig. 2 leader state)
- State machine apply loop (`lastApplied` → KV map)
- Event-driven WebSocket (heartbeat / election events) instead of snapshot polling
- Fault injection (crash, partition) + Next.js visualization UI
- Persistence of the Raft log across restarts

---

## Architecture

```
cmd/main.go
    │
    ├─ create N × raft.Node (in memory)
    ├─ WireRuntimePeers (PeerIDs → []*Node)
    ├─ Start() each node (election + heartbeats)
    ├─ optional: SubmitCommand on leader
    │
    └─ Gin HTTP
           GET /api/cluster  ──► ClusterHandler ──► ClusterService.Snapshot()
           GET /ws           ──► ServeWS        ──► periodic Snapshot() → JSON
                                      │
                                      ▼
                              Node.Snapshot()  (mu.Lock → copy → Unlock)
```

**Separation of concerns**

| Package | Responsibility |
|---------|----------------|
| `models` | DB/JSON shapes (`RaftNode`, `RaftNodeResponse`) |
| `internal/raft` | Consensus algorithm & timers |
| `internal/services` | Cluster use-cases (snapshot, later submit/crash) |
| `internal/api` | Gin routes & WebSocket |
| `internal/store` / `database` | GORM persistence (optional path) |

Runtime engine vs row model:

```go
raft.Node {
    RaftNode *models.RaftNode  // durable-ish state (term, log, peer IDs)
    Peers    []*Node           // live RPC targets
    // timers, channels, mutex — never exported to HTTP
}
```

---

## Requirements

- Go **1.25+** (see `go.mod`)
- Optional: PostgreSQL if you wire the GORM path

---

## Quick start

```bash
git clone https://github.com/Omjaiswal26/raft-consensus.git
cd raft-consensus

go mod tidy
go run cmd/main.go
```

On boot the process:

1. Starts a **3-node** cluster
2. Waits for election, submits `SET x=1` to the leader
3. Prints each node’s log / commit index
4. Listens on **`:8080`** (Gin default)

### Inspect the cluster

```bash
# REST
curl http://localhost:8080/api/cluster

# WebSocket (streams {"type":"cluster","nodes":[...]} )
npx wscat -c ws://localhost:8080/ws
```

Keep `go run cmd/main.go` running while you connect; if the process exits, the socket will fail with a bare connection error.

---

## Project layout

```
cmd/main.go                      # entry: cluster + HTTP
models/
  raft.go                        # RaftNode, LogEntry (GORM)
  node_response.go               # API DTO
internal/
  raft/
    node.go                      # roles, election, heartbeats, SubmitCommand
    rpc.go                       # RequestVote / AppendEntries types
    timer.go                     # election / heartbeat constants
    cluster.go                   # WireRuntimePeers
    snapshot.go                  # locked Snapshot() for readers
  services/
    cluster_service.go           # Snapshot across all nodes
    init_nodes.go                # DB-oriented node init (optional)
  api/
    router.go
    cluster_handler.go           # GET /api/cluster, GET /ws
  store/                         # GORM helpers
database/db.go                   # Postgres connect + AutoMigrate
```

---

## Protocol timings (simulation defaults)

| Constant | Value | Role |
|----------|-------|------|
| Election timeout | 150–300 ms (random) | Detect leader failure; avoid split votes |
| Heartbeat interval | 50 ms | Keep followers from starting elections |

Rule of thumb: **heartbeat ≪ election timeout**.

---

## Learning path (how this repo was built)

1. Follower loop + randomized election timer  
2. Candidate + `RequestVote` + majority → leader  
3. Heartbeats (`AppendEntries` empty) for stability  
4. `SubmitCommand` + log replication + commit  
5. HTTP snapshot API + WebSocket stream  
6. *(next)* UI, fault injection, full leader match/next indexes  

---

## References

- Ongaro & Ousterhout, *[In Search of an Understandable Consensus Algorithm (Raft)](https://raft.github.io/raft.pdf)*  
- [raft.github.io](https://raft.github.io/) — visualizations and links  
- Diego Ongaro’s dissertation for deeper treatment of membership and log compaction  

---

## License

This project is licensed under the [MIT License](./LICENSE) — a simple, permissive license well suited to educational and open-source learning projects.
