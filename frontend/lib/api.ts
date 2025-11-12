import axios from "axios";
import type { SearchRequest, SearchResponse } from "@/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

export const searchAPI = {
  search: async (request: SearchRequest): Promise<SearchResponse> => {
    const response = await api.post<any>("/api/search", request);

    // Transform the response to match SearchResponse interface
    return {
      query: response.data.query || request.query,
      mode: response.data.mode || request.mode,
      answer: response.data.answer || "",
      sources: response.data.sources || [],
      reasoning: response.data.reasoning,
      processingTime:
        response.data.processing_time || response.data.processingTime || 0,
      timestamp: response.data.timestamp || new Date().toISOString(),
    };
  },

  health: async () => {
    const response = await api.get("/api/health");
    return response.data;
  },
};
