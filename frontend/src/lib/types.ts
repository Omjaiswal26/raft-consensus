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

export type ClusterMessage = {
  type: "cluster";
  nodes: RaftNode[];
};

export type ApiResponse<T> = {
  success: boolean;
  message: string;
  data: T;
};
