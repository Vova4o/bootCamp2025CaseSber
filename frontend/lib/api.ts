import axios from "axios";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const api = axios.create({
  baseURL: API_URL,
  timeout: 60000,
  headers: {
    "Content-Type": "application/json",
  },
});

export interface SearchRequest {
  query: string;
  mode: "auto" | "simple" | "pro";
}

export interface Source {
  url: string;
  title: string;
  content: string;
  semantic_score?: number;
}

export interface SearchResponse {
  mode: string;
  query: string;
  answer: string;
  sources: Source[];
  reasoning_steps?: string[];
  search_queries?: string[];
  response_time: number;
}

export const searchAPI = {
  search: async (data: SearchRequest): Promise<SearchResponse> => {
    const response = await api.post<SearchResponse>("/api/search", data);
    return response.data;
  },

  getHistory: async (limit: number = 20) => {
    const response = await api.get("/api/history", { params: { limit } });
    return response.data;
  },
};

export default api;
