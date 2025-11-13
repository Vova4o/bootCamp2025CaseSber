export type SearchMode = "auto" | "simple" | "pro";

export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: string;
  sources?: Source[];
  reasoning?: string;
}

export interface ChatSession {
  id: string;
  messages: Message[];
  mode: SearchMode;
  created_at: string;
  updated_at: string;
}

export interface SearchRequest {
  query: string;
  mode: SearchMode;
  session_id?: string; // For conversation context
  previous_messages?: Message[]; // Context from previous turns
}

export interface Source {
  title: string;
  url: string;
  snippet: string;
  credibility?: number;
}

export interface SearchResponse {
  query: string;
  mode: SearchMode;
  answer: string;
  sources: Source[];
  reasoning?: string;
  processingTime?: number;
  timestamp?: string;
  session_id?: string; // Session tracking
  context_used?: boolean; // Whether context was used
}

export interface StreamChunk {
  type: "thinking" | "searching" | "analyzing" | "result";
  content: string;
  agent?: string;
}
