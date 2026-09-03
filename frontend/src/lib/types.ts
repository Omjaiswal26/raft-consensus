export type LogEntry = {
  term: number;
  command: string;
};

export type RaftNode = {
  id: number;
  state: string;
  current_term: number;
  voted_for: number | null;
  leader_id: number | null;
  commit_index: number;
  last_applied: number;
  peers: number[];
  log: LogEntry[];
  kv: Record<string, string>;
};

export type RaftEventType =
  | "election_start"
  | "request_vote"
  | "vote_granted"
  | "became_leader"
  | "heartbeat"
  | "append_entries";

export type RaftEvent = {
  type: RaftEventType;
  from: number;
  to: number;
  term: number;
  at: string;
};

export type ClusterMessage = {
  type: "cluster";
  nodes: RaftNode[];
};

export type WsMessage = ClusterMessage | RaftEvent;

export type ApiResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};

export function isClusterMessage(msg: WsMessage): msg is ClusterMessage {
  return msg.type === "cluster";
}

export function isRaftEvent(msg: WsMessage): msg is RaftEvent {
  return msg.type !== "cluster";
}
