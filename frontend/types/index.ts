export type SearchMode = 
  | 'auto' 
  | 'simple' 
  | 'pro' 
  | 'pro-social' 
  | 'pro-academic' 
  | 'pro-finance';

export interface Source {
  title: string;
  url: string;
  snippet: string;
  credibility?: number;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: number;
  sources?: Source[];
  reasoning?: string;
}

export interface ChatSession {
  id: string;
  mode: SearchMode;
  created_at: number;
  updated_at: number;
  messages: Message[];
}

export interface SearchRequest {
  query: string;
  mode: SearchMode;
  session_id?: string;
}

export interface SearchResponse {
  query: string;
  mode: SearchMode;
  answer: string;
  sources: Source[];
  reasoning?: string;
  processing_time: number;
  timestamp: number;
  session_id?: string;
  context_used?: boolean;
}