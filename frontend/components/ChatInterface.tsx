"use client";

import { useState, useEffect, useRef } from "react";
import { searchAPI } from "@/lib/api";
import type { ChatSession, SearchMode, Message } from "@/types";
import ChatList from "./ChatList";
import ChatMessage from "./ChatMessage";
import CompactModeSelector from "./CompactModeSelector";
import { Send, Loader2, Menu, X } from "lucide-react";

export default function ChatInterface() {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(
    null
  );
  const [mode, setMode] = useState<SearchMode>("auto");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Load sessions from localStorage on mount
  useEffect(() => {
    const savedSessions = localStorage.getItem("chat_sessions");
    if (savedSessions) {
      const parsed = JSON.parse(savedSessions);
      setSessions(parsed);
      if (parsed.length > 0) {
        loadSession(parsed[0].id);
      }
    }
  }, []);

  // Save sessions to localStorage
  useEffect(() => {
    if (sessions.length > 0) {
      localStorage.setItem("chat_sessions", JSON.stringify(sessions));
    }
  }, [sessions]);

  // Scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [currentSession?.messages]);

  const loadSession = async (sessionId: string) => {
    try {
      const session = await searchAPI.getSession(sessionId);
      setCurrentSession(session);
      setMode(session.mode);
    } catch (error) {
      console.error("Failed to load session:", error);
      const session = sessions.find((s) => s.id === sessionId);
      if (session) {
        setCurrentSession(session);
        setMode(session.mode);
      }
    }
  };

  const createNewChat = async () => {
    try {
      const newSession = await searchAPI.createSession(mode);
      setSessions([newSession, ...sessions]);
      setCurrentSession(newSession);
    } catch (error) {
      console.error("Failed to create session:", error);
      const newSession: ChatSession = {
        id: `local-${Date.now()}`,
        mode,
        created_at: Math.floor(Date.now() / 1000),
        updated_at: Math.floor(Date.now() / 1000),
        messages: [],
      };
      setSessions([newSession, ...sessions]);
      setCurrentSession(newSession);
    }
  };

  const deleteSession = async (sessionId: string) => {
    try {
      await searchAPI.deleteSession(sessionId);
    } catch (error) {
      console.error("Failed to delete session:", error);
    }

    const newSessions = sessions.filter((s) => s.id !== sessionId);
    setSessions(newSessions);

    if (currentSession?.id === sessionId) {
      if (newSessions.length > 0) {
        loadSession(newSessions[0].id);
      } else {
        setCurrentSession(null);
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim() || loading) return;

    if (!currentSession) {
      await createNewChat();
      // После создания сессии нужно отправить сообщение
      setTimeout(() => {
        handleSubmit(e);
      }, 100);
      return;
    }

    const userMessage: Message = {
      id: `msg-${Date.now()}`,
      role: "user",
      content: query.trim(),
      timestamp: Math.floor(Date.now() / 1000),
    };

    const updatedSession = {
      ...currentSession,
      messages: [...currentSession.messages, userMessage],
      updated_at: Math.floor(Date.now() / 1000),
    };
    setCurrentSession(updatedSession);
    setQuery("");
    setLoading(true);

    try {
      const response = await searchAPI.sendMessage(
        currentSession.id,
        userMessage.content,
        mode
      );

      const assistantMessage: Message = {
        id: `msg-${Date.now()}-assistant`,
        role: "assistant",
        content: response.answer,
        timestamp: Math.floor(Date.now() / 1000),
        sources: response.sources,
        reasoning: response.reasoning,
      };

      const finalSession = {
        ...updatedSession,
        messages: [...updatedSession.messages, assistantMessage],
        // НЕ перезаписываем mode из ответа - сохраняем выбор пользователя
        mode: mode,
        updated_at: Math.floor(Date.now() / 1000),
      };

      setCurrentSession(finalSession);
      setSessions(
        sessions.map((s) => (s.id === finalSession.id ? finalSession : s))
      );
    } catch (error) {
      console.error("Failed to send message:", error);
      setCurrentSession(currentSession);
    } finally {
      setLoading(false);
      // Возвращаем фокус в input после отправки
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  };

  // Обработчик Enter в input
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (!loading && query.trim()) {
        handleSubmit(e as any);
      }
    }
  };

  return (
    <div className="flex h-screen bg-gray-950">
      {/* Sidebar */}
      <div
        className={`${
          sidebarOpen ? "w-80" : "w-0"
        } transition-all duration-300 overflow-hidden`}
      >
        <ChatList
          sessions={sessions}
          currentSessionId={currentSession?.id}
          onSelectSession={loadSession}
          onDeleteSession={deleteSession}
          onNewChat={createNewChat}
        />
      </div>

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="bg-gray-900/50 backdrop-blur-sm border-b border-gray-800 p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="p-2 hover:bg-gray-800 rounded-lg transition-colors text-gray-400 hover:text-gray-200"
              >
                {sidebarOpen ? <X size={20} /> : <Menu size={20} />}
              </button>
              <h1 className="text-xl font-bold bg-linear-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                🔍 Research Pro
              </h1>
            </div>
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4">
          {!currentSession ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <div className="text-6xl mb-4">🔍</div>
                <h2 className="text-2xl font-bold text-gray-200 mb-2">
                  Начните новый поиск
                </h2>
                <p className="text-gray-500">
                  Создайте чат или задайте вопрос
                </p>
              </div>
            </div>
          ) : currentSession.messages.length === 0 ? (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <div className="text-6xl mb-4">💬</div>
                <h2 className="text-2xl font-bold text-gray-200 mb-2">
                  Чат создан
                </h2>
                <p className="text-gray-500">Задайте свой первый вопрос</p>
              </div>
            </div>
          ) : (
            currentSession.messages.map((message, idx) => (
              <ChatMessage
                key={message.id || idx}
                message={message}
                mode={
                  message.role === "assistant"
                    ? currentSession.mode
                    : undefined
                }
              />
            ))
          )}
          {loading && (
            <div className="flex items-center gap-3 p-4 bg-gray-800/50 backdrop-blur-sm rounded-xl mr-12">
              <div className="w-8 h-8 rounded-full bg-linear-to-br from-purple-500 to-pink-500 flex items-center justify-center">
                <Loader2 size={18} className="text-white animate-spin" />
              </div>
              <div className="text-gray-400">Обрабатываю запрос...</div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="border-t border-gray-800 bg-gray-900/50 backdrop-blur-sm p-4">
          <form onSubmit={handleSubmit} className="max-w-4xl mx-auto">
            <div className="flex gap-2">
              <CompactModeSelector
                mode={mode}
                onModeChange={setMode}
                disabled={loading}
              />
              <div className="flex-1 flex gap-2 bg-gray-800 rounded-xl p-1">
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Задайте вопрос..."
                  disabled={loading}
                  className="flex-1 px-3 bg-transparent border-none focus:outline-none disabled:cursor-not-allowed text-gray-100 placeholder-gray-500"
                  autoFocus
                />
                <button
                  type="submit"
                  disabled={loading || !query.trim()}
                  className="h-10 w-10 bg-linear-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-500 hover:to-purple-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center shrink-0"
                >
                  {loading ? (
                    <Loader2 size={18} className="animate-spin" />
                  ) : (
                    <Send size={18} />
                  )}
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}