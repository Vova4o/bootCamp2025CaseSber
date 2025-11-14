"use client";

import { useState, useEffect, useRef } from "react";
import { searchAPI } from "@/lib/api";
import type { ChatSession, SearchMode, Message } from "@/types";
import ChatList from "./ChatList";
import ChatMessage from "./ChatMessage";
import CompactModeSelector from "./CompactModeSelector";
import { Send, Loader2, Menu, X, Search } from "lucide-react";

export default function ChatInterface() {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(
    null
  );
  const [mode, setMode] = useState<SearchMode>("auto");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

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

  useEffect(() => {
    if (sessions.length > 0) {
      localStorage.setItem("chat_sessions", JSON.stringify(sessions));
    }
  }, [sessions]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [currentSession?.messages]);

  const loadSession = async (sessionId: string) => {
    try {
      const session = await searchAPI.getSession(sessionId);
      setCurrentSession(session);
      setMode(session.mode);
      setSidebarOpen(false);
    } catch (error) {
      console.error("Failed to load session:", error);
      const session = sessions.find((s) => s.id === sessionId);
      if (session) {
        setCurrentSession(session);
        setMode(session.mode);
        setSidebarOpen(false);
      }
    }
  };

  const createNewChat = async () => {
    try {
      const newSession = await searchAPI.createSession(mode);
      setSessions([newSession, ...sessions]);
      setCurrentSession(newSession);
      setSidebarOpen(false);
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
      setSidebarOpen(false);
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
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (!loading && query.trim()) {
        handleSubmit(e as any);
      }
    }
  };

  return (
    <div className="flex h-screen bg-neutral-950 overflow-hidden">
      {/* Mobile Overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={`fixed lg:relative inset-y-0 left-0 z-50 w-80 transform transition-transform duration-300 ease-in-out ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        }`}
      >
        <ChatList
          sessions={sessions}
          currentSessionId={currentSession?.id}
          onSelectSession={loadSession}
          onDeleteSession={deleteSession}
          onNewChat={createNewChat}
          onClose={() => setSidebarOpen(false)}
        />
      </div>

      {/* Main Chat Area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header */}
        <div className="flex-shrink-0 border-b border-neutral-800 bg-neutral-900/80 backdrop-blur-xl">
          <div className="flex items-center justify-between px-4 py-3 lg:px-6 lg:py-4">
            <div className="flex items-center gap-3">
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className="p-2 hover:bg-neutral-800 rounded-lg transition-colors text-neutral-400 hover:text-neutral-200"
                aria-label="Toggle sidebar"
              >
                {sidebarOpen ? <X size={20} /> : <Menu size={20} />}
              </button>
              <div className="flex items-center gap-2">
                <Search size={20} className="text-neutral-500" />
                <h1 className="text-lg lg:text-xl font-semibold text-neutral-100">
                  Research Pro
                </h1>
              </div>
            </div>
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto">
          <div className="max-w-4xl mx-auto px-4 py-6 lg:px-6 lg:py-8">
            {!currentSession ? (
              <div className="flex items-center justify-center h-full min-h-[60vh]">
                <div className="text-center">
                  <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-neutral-800 flex items-center justify-center">
                    <Search size={32} className="text-neutral-500" />
                  </div>
                  <h2 className="text-2xl font-semibold text-neutral-200 mb-2">
                    Начните новый поиск
                  </h2>
                  <p className="text-neutral-500">
                    Создайте чат или задайте вопрос
                  </p>
                </div>
              </div>
            ) : currentSession.messages.length === 0 ? (
              <div className="flex items-center justify-center h-full min-h-[60vh]">
                <div className="text-center">
                  <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-neutral-800 flex items-center justify-center">
                    <Send size={32} className="text-neutral-500" />
                  </div>
                  <h2 className="text-2xl font-semibold text-neutral-200 mb-2">
                    Чат создан
                  </h2>
                  <p className="text-neutral-500">Задайте свой первый вопрос</p>
                </div>
              </div>
            ) : (
              <div className="space-y-6">
                {currentSession.messages.map((message, idx) => (
                  <ChatMessage
                    key={message.id || idx}
                    message={message}
                    mode={
                      message.role === "assistant"
                        ? currentSession.mode
                        : undefined
                    }
                  />
                ))}
              </div>
            )}
            {loading && (
              <div className="flex items-center gap-3 p-4 bg-neutral-900/50 backdrop-blur-sm rounded-xl border border-neutral-800 mt-6">
                <div className="w-8 h-8 rounded-full bg-neutral-800 flex items-center justify-center">
                  <Loader2 size={18} className="text-neutral-400 animate-spin" />
                </div>
                <div className="text-neutral-400">Обрабатываю запрос...</div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>
        </div>

        {/* Input */}
        <div className="flex-shrink-0 border-t border-neutral-800 bg-neutral-900/80 backdrop-blur-xl">
          <div className="max-w-4xl mx-auto px-4 py-4 lg:px-6">
            <form onSubmit={handleSubmit} className="flex flex-col gap-3">
              <CompactModeSelector
                mode={mode}
                onModeChange={setMode}
                disabled={loading}
              />
              <div className="flex gap-2 bg-neutral-900 rounded-xl p-1.5 border border-neutral-800 focus-within:border-neutral-700 transition-colors">
                <input
                  ref={inputRef}
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Задайте вопрос..."
                  disabled={loading}
                  className="flex-1 px-3 py-2 bg-transparent border-none focus:outline-none disabled:cursor-not-allowed text-neutral-100 placeholder-neutral-500"
                  autoFocus
                />
                <button
                  type="submit"
                  disabled={loading || !query.trim()}
                  className="h-10 w-10 bg-neutral-100 text-neutral-900 rounded-lg hover:bg-white disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center shrink-0"
                  aria-label="Send message"
                >
                  {loading ? (
                    <Loader2 size={18} className="animate-spin" />
                  ) : (
                    <Send size={18} />
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}