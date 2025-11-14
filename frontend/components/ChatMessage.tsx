"use client";

import { Message } from "@/types";
import { User, Bot, ExternalLink, ChevronDown, ChevronUp } from "lucide-react";
import { useState } from "react";

interface ChatMessageProps {
  message: Message;
  mode?: string;
}

export default function ChatMessage({ message, mode }: ChatMessageProps) {
  const [showReasoning, setShowReasoning] = useState(false);
  const isUser = message.role === "user";

  return (
    <div
      className={`flex gap-3 p-4 rounded-xl ${
        isUser
          ? "bg-gradient-to-r from-blue-900/50 to-blue-800/50 ml-12"
          : "bg-gray-800/50 backdrop-blur-sm mr-12"
      }`}
    >
      {/* Avatar */}
      <div
        className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
          isUser
            ? "bg-gradient-to-br from-blue-500 to-blue-600"
            : "bg-gradient-to-br from-purple-500 to-pink-500"
        }`}
      >
        {isUser ? (
          <User size={18} className="text-white" />
        ) : (
          <Bot size={18} className="text-white" />
        )}
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-2">
          <span className="font-semibold text-gray-100">
            {isUser ? "Вы" : "Ассистент"}
          </span>
          {!isUser && mode && (
            <span
              className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                mode.startsWith("pro-social")
                  ? "bg-pink-500/20 text-pink-300 border border-pink-500/30"
                  : mode.startsWith("pro-academic")
                  ? "bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
                  : mode.startsWith("pro-finance")
                  ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                  : mode.startsWith("pro") || mode.includes("→ pro")
                  ? "bg-purple-500/20 text-purple-300 border border-purple-500/30"
                  : mode === "simple"
                  ? "bg-green-500/20 text-green-300 border border-green-500/30"
                  : "bg-blue-500/20 text-blue-300 border border-blue-500/30"
              }`}
            >
              {mode.startsWith("pro-social")
                ? "👥 Social"
                : mode.startsWith("pro-academic")
                ? "🎓 Academic"
                : mode.startsWith("pro-finance")
                ? "💰 Finance"
                : mode.startsWith("pro") || mode.includes("→ pro")
                ? "🧠 Pro"
                : mode === "simple"
                ? "⚡ Simple"
                : "✨ Auto"}
            </span>
          )}
        </div>

        <div className="text-gray-200 whitespace-pre-wrap wrap-break-words leading-relaxed">
          {message.content}
        </div>

        {/* Reasoning */}
        {message.reasoning && (
          <div className="mt-3">
            <button
              onClick={() => setShowReasoning(!showReasoning)}
              className="flex items-center gap-2 text-sm text-gray-400 hover:text-gray-300 transition-colors"
            >
              {showReasoning ? (
                <ChevronUp size={16} />
              ) : (
                <ChevronDown size={16} />
              )}
              <span>Ход рассуждений</span>
            </button>
            {showReasoning && (
              <div className="mt-2 p-3 bg-gray-900/50 rounded-lg text-sm text-gray-300 whitespace-pre-wrap border border-gray-700">
                {message.reasoning}
              </div>
            )}
          </div>
        )}

        {/* Sources */}
        {message.sources && message.sources.length > 0 && (
          <div className="mt-3 space-y-2">
            <div className="text-sm font-semibold text-gray-300">
              📚 Источники:
            </div>
            {message.sources.map((source, idx) => (
              <a
                key={idx}
                href={source.url}
                target="_blank"
                rel="noopener noreferrer"
                className="block p-3 bg-gray-900/50 rounded-lg hover:bg-gray-900/70 transition-colors border border-gray-700/50 hover:border-gray-600"
              >
                <div className="flex items-start gap-2">
                  <ExternalLink
                    size={14}
                    className="text-gray-500 mt-1 flex-shrink-0"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-sm text-gray-200 truncate">
                      {source.title}
                    </div>
                    <div className="text-xs text-gray-500 mt-1 line-clamp-2">
                      {source.snippet}
                    </div>
                  </div>
                </div>
              </a>
            ))}
          </div>
        )}

        {/* Timestamp */}
        <div className="text-xs text-gray-600 mt-2">
          {new Date(message.timestamp * 1000).toLocaleTimeString("ru-RU", {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </div>
      </div>
    </div>
  );
}
