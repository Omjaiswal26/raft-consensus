"use client";

import { useEffect, useRef, useState } from "react";
import type { RaftEvent, RaftEventType, RaftNode } from "@/lib/types";

const NODE_POS: Record<number, { x: number; y: number }> = {
  1: { x: 50, y: 18 },
  2: { x: 18, y: 82 },
  3: { x: 82, y: 82 },
};

const EDGES: [number, number][] = [
  [1, 2],
  [1, 3],
  [2, 3],
];

const EVENT_STYLE: Record<
  RaftEventType,
  { color: string; label: string; duration: number }
> = {
  heartbeat: { color: "#38bdf8", label: "heartbeat", duration: 400 },
  request_vote: { color: "#fbbf24", label: "vote req", duration: 550 },
  vote_granted: { color: "#34d399", label: "vote", duration: 550 },
  append_entries: { color: "#a78bfa", label: "append", duration: 650 },
  election_start: { color: "#fb923c", label: "election", duration: 700 },
  became_leader: { color: "#facc15", label: "leader", duration: 900 },
};

type Pulse = {
  id: number;
  type: RaftEventType;
  from: number;
  to: number;
  startedAt: number;
};

type ReplicationGraphProps = {
  nodes: RaftNode[];
  events: RaftEvent[];
};

function nodeStateClass(state: string) {
  switch (state) {
    case "leader":
      return "fill-emerald-500/30 stroke-emerald-400";
    case "candidate":
      return "fill-amber-500/25 stroke-amber-400";
    default:
      return "fill-slate-800/80 stroke-slate-500";
  }
}

export default function ReplicationGraph({ nodes, events }: ReplicationGraphProps) {
  const [pulses, setPulses] = useState<Pulse[]>([]);
  const [flashes, setFlashes] = useState<Record<number, RaftEventType | null>>({});
  const [, setFrame] = useState(0);
  const nextId = useRef(0);
  const lastHeartbeat = useRef<Record<string, number>>({});

  useEffect(() => {
    if (events.length === 0) return;
    const latest = events[events.length - 1];
    const style = EVENT_STYLE[latest.type];
    if (!style) return;

    if (latest.type === "heartbeat") {
      const key = `${latest.from}-${latest.to}`;
      const now = Date.now();
      if (now - (lastHeartbeat.current[key] ?? 0) < 180) return;
      lastHeartbeat.current[key] = now;
    }

    const id = ++nextId.current;
    const startedAt = Date.now();

    if (latest.to === 0) {
      setFlashes((prev) => ({ ...prev, [latest.from]: latest.type }));
      window.setTimeout(() => {
        setFlashes((prev) => {
          if (prev[latest.from] !== latest.type) return prev;
          const next = { ...prev };
          delete next[latest.from];
          return next;
        });
      }, style.duration);
      return;
    }

    setPulses((prev) => [...prev.slice(-24), { id, type: latest.type, from: latest.from, to: latest.to, startedAt }]);
  }, [events]);

  useEffect(() => {
    if (pulses.length === 0) return;
    let raf = 0;
    const loop = () => {
      setFrame((f) => f + 1);
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, [pulses.length]);

  const nodeById = Object.fromEntries(nodes.map((n) => [n.id, n]));

  return (
    <section className="rounded-2xl border border-slate-800 bg-slate-950/60 p-4 backdrop-blur">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h2 className="font-mono text-xs uppercase tracking-[0.2em] text-slate-400">
          Replication flow
        </h2>
        <div className="flex flex-wrap gap-2 font-mono text-[10px] text-slate-500">
          {Object.entries(EVENT_STYLE).map(([type, s]) => (
            <span key={type} className="inline-flex items-center gap-1.5">
              <span
                className="h-2 w-2 rounded-full"
                style={{ backgroundColor: s.color }}
              />
              {s.label}
            </span>
          ))}
        </div>
      </div>

      <svg viewBox="0 0 100 100" className="mx-auto h-auto w-full max-w-2xl">
        {EDGES.map(([a, b]) => {
          const p1 = NODE_POS[a];
          const p2 = NODE_POS[b];
          return (
            <line
              key={`${a}-${b}`}
              x1={p1.x}
              y1={p1.y}
              x2={p2.x}
              y2={p2.y}
              stroke="rgba(148,163,184,0.25)"
              strokeWidth="0.35"
            />
          );
        })}

        {pulses.map((pulse) => {
          const from = NODE_POS[pulse.from];
          const to = NODE_POS[pulse.to];
          if (!from || !to) return null;
          const style = EVENT_STYLE[pulse.type];
          const elapsed = Date.now() - pulse.startedAt;
          if (elapsed >= style.duration) return null;
          const t = elapsed / style.duration;
          const x = from.x + (to.x - from.x) * t;
          const y = from.y + (to.y - from.y) * t;
          return (
            <circle
              key={pulse.id}
              cx={x}
              cy={y}
              r="1.6"
              fill={style.color}
              opacity={1 - t * 0.35}
            />
          );
        })}

        {nodes.map((node) => {
          const pos = NODE_POS[node.id];
          if (!pos) return null;
          const flash = flashes[node.id];
          const flashColor = flash ? EVENT_STYLE[flash].color : null;
          return (
            <g key={node.id}>
              {flashColor && (
                <circle
                  cx={pos.x}
                  cy={pos.y}
                  r="9"
                  fill="none"
                  stroke={flashColor}
                  strokeWidth="0.6"
                  opacity="0.85"
                />
              )}
              <circle
                cx={pos.x}
                cy={pos.y}
                r="7"
                className={nodeStateClass(node.state)}
                strokeWidth="0.5"
              />
              <text
                x={pos.x}
                y={pos.y + 0.8}
                textAnchor="middle"
                className="fill-white font-mono text-[3.2px]"
              >
                N{node.id}
              </text>
              <text
                x={pos.x}
                y={pos.y + 12}
                textAnchor="middle"
                className="fill-slate-500 font-mono text-[2.6px] uppercase"
              >
                {nodeById[node.id]?.state ?? "?"}
              </text>
            </g>
          );
        })}
      </svg>
    </section>
  );
}
