import axios from 'axios';
import type {
  SearchRequest,
  SearchResponse,
  ChatSession,
  SearchMode,
} from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const searchAPI = {
  search: async (request: SearchRequest): Promise<SearchResponse> => {
    const response = await api.post<any>('/api/search', request);

    return {
      query: response.data.query || request.query,
      mode: response.data.mode || request.mode,
      answer: response.data.answer || '',
      sources: response.data.sources || [],
      reasoning: response.data.reasoning,
      processing_time:
        response.data.processing_time || response.data.processingTime || 0,
      timestamp: response.data.timestamp || Date.now(),
      session_id: response.data.session_id,
      context_used: response.data.context_used,
    };
  },

  createSession: async (mode: SearchMode): Promise<ChatSession> => {
    const response = await api.post<ChatSession>('/api/chat/session', {
      mode,
    });
    return response.data;
  },

  listSessions: async (): Promise<ChatSession[]> => {
    // Если на бэкенде нет этого эндпоинта, используем локальное хранилище
    const sessions = localStorage.getItem('chat_sessions');
    return sessions ? JSON.parse(sessions) : [];
  },

  getSession: async (sessionId: string): Promise<ChatSession> => {
    const response = await api.get<ChatSession>(
      `/api/chat/session/${sessionId}`
    );
    return response.data;
  },

  sendMessage: async (
    sessionId: string,
    query: string,
    mode: SearchMode
  ): Promise<SearchResponse> => {
    const response = await api.post<any>(
      `/api/chat/session/${sessionId}/message`,
      {
        query,
        mode,
      }
    );

    return {
      query: response.data.query || query,
      mode: response.data.mode || mode,
      answer: response.data.answer || '',
      sources: response.data.sources || [],
      reasoning: response.data.reasoning,
      processing_time: response.data.processing_time || 0,
      timestamp: response.data.timestamp || Date.now(),
      session_id: response.data.session_id,
      context_used: response.data.context_used,
    };
  },

  deleteSession: async (sessionId: string): Promise<void> => {
    await api.delete(`/api/chat/session/${sessionId}`);
  },

  health: async () => {
    const response = await api.get('/api/health');
    return response.data;
  },
};