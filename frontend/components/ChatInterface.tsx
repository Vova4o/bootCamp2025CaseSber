"use client";

import { useState, useEffect, useRef } from "react";
import {
  Search,
  Loader2,
  Clock,
  ExternalLink,
  MessageSquare,
  RotateCcw,
} from "lucide-react";
import { searchAPI } from "@/lib/api";
import type { Message, SearchMode, Source } from "@/types";

interface ChatInterfaceProps {
  mode: SearchMode;
}

export default function ChatInterface({ mode }: ChatInterfaceProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Create new session
  const createNewSession = async () => {
    try {
      const session = await searchAPI.createSession(mode);
      setSessionId(session.id);
      setMessages([]);
    } catch (error) {
      console.error("Failed to create session:", error);
    }
  };

  // Initialize session on mount
  useEffect(() => {
    createNewSession();
  }, [mode]);

  const handleSendMessage = async () => {
    if (!query.trim() || !sessionId) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: "user",
      content: query,
      timestamp: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setQuery("");
    setLoading(true);

    try {
      const response = await searchAPI.sendMessage(sessionId, query, mode);

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: "assistant",
        content: response.answer,
        timestamp: response.timestamp || new Date().toISOString(),
        sources: response.sources,
        reasoning: response.reasoning,
      };

      setMessages((prev) => [...prev, assistantMessage]);
    } catch (error: any) {
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: "assistant",
        content: `Ошибка: ${
          error.response?.data?.detail || "Не удалось получить ответ"
        }`,
        timestamp: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-3xl shadow-2xl flex flex-col h-[600px]">
      {/* Header */}
      <div className="p-6 border-b border-gray-200 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <MessageSquare className="text-blue-600" size={24} />
          <div>
            <h2 className="text-xl font-bold text-gray-900">
              Chat Mode {mode === "pro" ? "🧠 Pro" : "⚡ Simple"}
            </h2>
            <p className="text-sm text-gray-500">
              {messages.length} сообщений в этой сессии
            </p>
          </div>
        </div>
        <button
          onClick={createNewSession}
          className="px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-gray-700 font-medium flex items-center gap-2 transition-colors"
        >
          <RotateCcw size={16} />
          Новая сессия
        </button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {messages.length === 0 && (
          <div className="text-center text-gray-500 mt-12">
            <MessageSquare className="mx-auto mb-4 text-gray-300" size={48} />
            <p className="text-lg">Начните диалог с вашего вопроса</p>
            <p className="text-sm mt-2">
              {mode === "pro"
                ? "Pro Mode будет учитывать контекст предыдущих сообщений"
                : "Simple Mode для быстрых ответов"}
            </p>
          </div>
        )}

        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex ${
              message.role === "user" ? "justify-end" : "justify-start"
            }`}
          >
            <div
              className={`max-w-[80%] rounded-2xl p-4 ${
                message.role === "user"
                  ? "bg-blue-600 text-white"
                  : "bg-gray-100 text-gray-900"
              }`}
            >
              <p className="whitespace-pre-wrap leading-relaxed">
                {message.content}
              </p>

              {/* Sources for assistant messages */}
              {message.role === "assistant" &&
                message.sources &&
                message.sources.length > 0 && (
                  <div className="mt-4 space-y-2">
                    <p className="text-xs font-semibold text-gray-600 mb-2">
                      📚 Источники:
                    </p>
                    {message.sources.map((source, idx) => (
                      <a
                        key={idx}
                        href={source.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block p-3 bg-white rounded-lg border border-gray-200 hover:border-blue-400 transition-colors text-sm"
                      >
                        <p className="font-medium text-gray-900 mb-1">
                          {source.title}
                        </p>
                        <p className="text-xs text-gray-500 truncate">
                          {source.url}
                        </p>
                      </a>
                    ))}
                  </div>
                )}

              {/* Reasoning for pro mode */}
              {message.role === "assistant" && message.reasoning && (
                <div className="mt-3 pt-3 border-t border-gray-300">
                  <p className="text-xs font-semibold text-gray-600 mb-1">
                    💡 Рассуждения:
                  </p>
                  <p className="text-xs text-gray-700">{message.reasoning}</p>
                </div>
              )}

              <p className="text-xs opacity-60 mt-2">
                {new Date(message.timestamp).toLocaleTimeString()}
              </p>
            </div>
          </div>
        ))}

        {loading && (
          <div className="flex justify-start">
            <div className="bg-gray-100 rounded-2xl p-4 max-w-[80%]">
              <Loader2 className="animate-spin text-blue-600" size={20} />
              <p className="text-sm text-gray-600 mt-2">
                {mode === "pro"
                  ? "Анализирую с учётом контекста..."
                  : "Ищу ответ..."}
              </p>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-6 border-t border-gray-200">
        <div className="flex gap-4">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyPress={(e) =>
              e.key === "Enter" && !loading && handleSendMessage()
            }
            placeholder="Задайте следующий вопрос..."
            disabled={loading}
            className="flex-1 px-6 py-4 text-lg border-2 border-gray-200 rounded-xl focus:outline-none focus:border-blue-500 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-400"
          />
          <button
            onClick={handleSendMessage}
            disabled={loading || !query.trim()}
            className="px-8 py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-700 hover:to-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2 shadow-lg"
          >
            {loading ? (
              <Loader2 className="animate-spin" size={20} />
            ) : (
              <Search size={20} />
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
