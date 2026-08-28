# Raft Consensus Simulation

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-Omjaiswal26%2Fraft--consensus-181717?logo=github)](https://github.com/Omjaiswal26/raft-consensus)

A hands-on Go implementation of the [Raft consensus algorithm](https://raft.github.io/raft.pdf), built to understand the protocol in depth by constructing it piece by piece—not by wrapping an existing library.

**Repository:** [github.com/Omjaiswal26/raft-consensus](https://github.com/Omjaiswal26/raft-consensus)

Monorepo layout:

| Path | Role |
|------|------|
| [`backend/`](./backend) | Raft cluster + Gin HTTP/WebSocket API |
| [`frontend/`](./frontend) | Next.js realtime simulation UI |

---

## Why Raft?

Raft keeps a replicated log consistent across machines so that a majority agree on the same sequence of commands, even when nodes crash or the network partitions.

Compared to Paxos, Raft emphasizes **understandability**: clear roles (follower / candidate / leader), explicit terms, and two primary RPCs (`RequestVote`, `AppendEntries`).

| Concept | Where in the code |
|---------|-------------------|
| Randomized election timeouts | `backend/internal/raft/timer.go` |
| Leader election | `BecomeCandidate` → `RequestVote` → majority → `BecomeLeader` |
| Heartbeats | Empty `AppendEntries` on a ticker |
| Log replication | `SubmitCommand` + non-empty `AppendEntries` |
| State machine apply | `backend/internal/raft/apply.go` (log → KV) |
| Safe reads | `Node.Snapshot()` under lock |

---

## What's implemented

### Backend (`backend/`)

- In-process 3-node Raft cluster (election, heartbeats, replication, commit, apply → KV)
- REST + WebSocket APIs for inspection and client writes
- Layered API: router → handler → service → raft

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/cluster` | JSON snapshot (`success` / `message` / `data`) |
| `POST` | `/api/command` | Submit `{"command":"SET k=v"}` to the current leader |
| `GET` | `/ws` | Stream `{"type":"cluster","nodes":[...]}` ~every 300ms |

### Frontend (`frontend/`)

- Next.js app that connects to `/ws` and renders node role, term, log, and KV
- Command form that posts to `/api/command`

### Still open

- Per-follower `nextIndex` / `matchIndex` maps
- Event-driven WS (heartbeat/election pulses)
- Fault injection (crash / partition)
- Durable Raft log via GORM (scaffolding exists)

---

## Quick start

### 1. Backend

```bash
git clone https://github.com/Omjaiswal26/raft-consensus.git
cd raft-consensus/backend

go mod tidy
go run ./cmd/main.go
```

Listens on **`:8080`**. On boot it elects a leader and submits a sample `SET x=1`.

```bash
curl http://localhost:8080/api/cluster
curl -X POST http://localhost:8080/api/command \
  -H "Content-Type: application/json" \
  -d '{"command":"SET y=2"}'
npx wscat -c ws://localhost:8080/ws
```

### 2. Frontend

```bash
cd raft-consensus/frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). Optional env:

```bash
# frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

CORS allows `http://localhost:3000` for REST; WebSocket uses `CheckOrigin: true` in development.

---

## Project layout

```
backend/
  cmd/main.go                 # cluster + HTTP entry
  models/raft.go              # GORM RaftNode / LogEntry
  internal/
    raft/                     # consensus engine
    api/                      # Gin routes + WS
    services/                 # cluster use-cases
    dto/                      # API DTOs
    response/                 # success/message/data helpers
    store/                    # GORM helpers
  database/db.go
  go.mod
frontend/
  src/app/                    # Next.js App Router
  src/components/             # ClusterDashboard
  src/lib/types.ts
```

---

## Protocol timings (simulation defaults)

| Constant | Value | Role |
|----------|-------|------|
| Election timeout | 150–300 ms (random) | Detect leader failure |
| Heartbeat interval | 50 ms | Prevent spurious elections |

Rule of thumb: **heartbeat ≪ election timeout**.

---

## References

- Ongaro & Ousterhout, *[In Search of an Understandable Consensus Algorithm (Raft)](https://raft.github.io/raft.pdf)*
- [raft.github.io](https://raft.github.io/)

---

## License

[MIT License](./LICENSE)
