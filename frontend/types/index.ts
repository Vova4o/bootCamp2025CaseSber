export type SearchMode = "auto" | "simple" | "pro";

export interface SearchRequest {
  query: string;
  mode: SearchMode;
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
  processingTime: number;
  timestamp: string;
}

export interface StreamChunk {
  type: "thinking" | "searching" | "analyzing" | "result";
  content: string;
  agent?: string;
}
