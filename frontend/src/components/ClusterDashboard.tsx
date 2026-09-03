"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import ReplicationGraph from "@/components/ReplicationGraph";
import type {
  ApiResponse,
  RaftEvent,
  RaftNode,
  WsMessage,
} from "@/lib/types";
import { isClusterMessage, isRaftEvent } from "@/lib/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/ws";

function stateTone(state: string) {
  switch (state) {
    case "leader":
      return "border-emerald-500/60 bg-emerald-500/10 text-emerald-200";
    case "candidate":
      return "border-amber-500/60 bg-amber-500/10 text-amber-100";
    default:
      return "border-slate-600 bg-slate-900/80 text-slate-200";
  }
}

export default function ClusterDashboard() {
  const [nodes, setNodes] = useState<RaftNode[]>([]);
  const [raftEvents, setRaftEvents] = useState<RaftEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const [command, setCommand] = useState("SET z=3");
  const [status, setStatus] = useState("Connecting…");
  const [busy, setBusy] = useState(false);

  const leader = useMemo(
    () => nodes.find((n) => n.state === "leader") ?? null,
    [nodes]
  );

  useEffect(() => {
    let ws: WebSocket | null = null;
    let closed = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    const connect = () => {
      if (closed) return;
      ws = new WebSocket(WS_URL);
      ws.onopen = () => {
        setConnected(true);
        setStatus("Live");
      };
      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data) as WsMessage;
          if (isClusterMessage(msg) && Array.isArray(msg.nodes)) {
            setNodes(msg.nodes);
          } else if (isRaftEvent(msg)) {
            setRaftEvents((prev) => [...prev.slice(-49), msg]);
          }
        } catch {
          /* ignore malformed frames */
        }
      };
      ws.onclose = () => {
        setConnected(false);
        setStatus("Reconnecting…");
        retryTimer = setTimeout(connect, 1500);
      };
      ws.onerror = () => ws?.close();
    };

    connect();
    return () => {
      closed = true;
      if (retryTimer) clearTimeout(retryTimer);
      ws?.close();
    };
  }, []);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!command.trim()) return;
    setBusy(true);
    setStatus("Submitting…");
    try {
      const res = await fetch(`${API_BASE}/api/command`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ command: command.trim() }),
      });
      const body = (await res.json()) as ApiResponse<unknown>;
      setStatus(body.success ? body.message || "Submitted" : body.message || "Failed");
    } catch {
      setStatus("Request failed — is the backend running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-[#0b1220] text-slate-100">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-70"
        style={{
          background:
            "radial-gradient(ellipse 80% 50% at 10% -10%, rgba(16,185,129,0.18), transparent), radial-gradient(ellipse 60% 40% at 90% 10%, rgba(56,189,248,0.12), transparent), linear-gradient(180deg, #0b1220 0%, #071018 100%)",
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-[0.07]"
        style={{
          backgroundImage:
            "linear-gradient(rgba(148,163,184,0.35) 1px, transparent 1px), linear-gradient(90deg, rgba(148,163,184,0.35) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
        }}
      />

      <main className="relative mx-auto flex w-full max-w-6xl flex-col gap-10 px-6 py-12 sm:px-10">
        <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.25em] text-emerald-400/80">
              Simulation
            </p>
            <h1 className="mt-2 font-serif text-4xl tracking-tight text-white sm:text-5xl">
              Raft Consensus
            </h1>
            <p className="mt-3 max-w-xl text-sm leading-relaxed text-slate-400">
              Live Raft cluster — replication graph, roles, log, and KV state.
            </p>
          </div>
          <div className="flex items-center gap-3 font-mono text-xs">
            <span
              className={`inline-flex items-center gap-2 rounded-full border px-3 py-1.5 ${
                connected
                  ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-300"
                  : "border-amber-500/40 bg-amber-500/10 text-amber-200"
              }`}
            >
              <span
                className={`h-1.5 w-1.5 rounded-full ${
                  connected ? "bg-emerald-400" : "bg-amber-400 animate-pulse"
                }`}
              />
              {status}
            </span>
            {leader && (
              <span className="rounded-full border border-slate-700 bg-slate-900/80 px-3 py-1.5 text-slate-300">
                leader N{leader.id} · term {leader.current_term}
              </span>
            )}
          </div>
        </header>

        <form
          onSubmit={onSubmit}
          className="flex flex-col gap-3 rounded-2xl border border-slate-800 bg-slate-950/60 p-4 backdrop-blur sm:flex-row sm:items-center"
        >
          <label className="sr-only" htmlFor="command">
            Command
          </label>
          <input
            id="command"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder='SET key=value'
            className="flex-1 rounded-xl border border-slate-700 bg-slate-900 px-4 py-3 font-mono text-sm text-slate-100 outline-none ring-emerald-500/40 placeholder:text-slate-600 focus:ring-2"
          />
          <button
            type="submit"
            disabled={busy || !connected}
            className="rounded-xl bg-emerald-500 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Submit to leader
          </button>
        </form>

        <ReplicationGraph nodes={nodes} events={raftEvents} />

        <section className="grid gap-4 md:grid-cols-3">
          {nodes.length === 0 && (
            <p className="col-span-full rounded-2xl border border-dashed border-slate-700 bg-slate-950/40 px-6 py-16 text-center text-sm text-slate-500">
              Waiting for cluster snapshots from {WS_URL}
            </p>
          )}
          {nodes.map((node) => (
            <article
              key={node.id}
              className={`flex flex-col gap-4 rounded-2xl border p-5 shadow-[0_0_0_1px_rgba(15,23,42,0.6)] transition ${stateTone(
                node.state
              )}`}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h2 className="font-serif text-2xl text-white">Node {node.id}</h2>
                  <p className="mt-1 font-mono text-[11px] uppercase tracking-wider opacity-80">
                    {node.state}
                  </p>
                </div>
                <div className="text-right font-mono text-[11px] leading-5 opacity-80">
                  <div>term {node.current_term}</div>
                  <div>commit {node.commit_index}</div>
                  <div>applied {node.last_applied}</div>
                </div>
              </div>

              <div>
                <h3 className="mb-2 font-mono text-[10px] uppercase tracking-[0.2em] opacity-60">
                  Log
                </h3>
                <ul className="max-h-28 space-y-1 overflow-y-auto rounded-xl border border-white/5 bg-black/20 p-2 font-mono text-[11px]">
                  {(node.log ?? []).length === 0 && (
                    <li className="opacity-50">empty</li>
                  )}
                  {(node.log ?? []).map((entry, i) => (
                    <li key={`${node.id}-${i}`}>
                      <span className="opacity-50">[{i + 1} t{entry.term}]</span>{" "}
                      {entry.command}
                    </li>
                  ))}
                </ul>
              </div>

              <div>
                <h3 className="mb-2 font-mono text-[10px] uppercase tracking-[0.2em] opacity-60">
                  KV
                </h3>
                <pre className="overflow-x-auto rounded-xl border border-white/5 bg-black/20 p-2 font-mono text-[11px]">
                  {JSON.stringify(node.kv ?? {}, null, 2)}
                </pre>
              </div>
            </article>
          ))}
        </section>
      </main>
    </div>
  );
}
