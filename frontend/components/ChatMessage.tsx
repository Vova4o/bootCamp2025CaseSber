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

  const getModeLabel = (mode: string) => {
    if (mode.startsWith("pro-social")) return "Social";
    if (mode.startsWith("pro-academic")) return "Academic";
    if (mode.startsWith("pro-finance")) return "Finance";
    if (mode.startsWith("pro") || mode.includes("→ pro")) return "Pro";
    if (mode === "simple") return "Simple";
    return "Auto";
  };

  return (
    <div className={`flex gap-3 ${isUser ? "flex-row-reverse" : ""}`}>
      {/* Avatar */}
      <div
        className={`flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center ${
          isUser ? "bg-neutral-800" : "bg-neutral-800"
        }`}
      >
        {isUser ? (
          <User size={16} className="text-neutral-400" />
        ) : (
          <Bot size={16} className="text-neutral-400" />
        )}
      </div>

      {/* Content */}
      <div className={`flex-1 min-w-0 ${isUser ? "max-w-[85%]" : ""}`}>
        <div
          className={`p-4 rounded-xl ${
            isUser
              ? "bg-neutral-800 border border-neutral-700"
              : "bg-neutral-900/50 border border-neutral-800"
          }`}
        >
          {/* Header */}
          {!isUser && mode && (
            <div className="flex items-center gap-2 mb-3 pb-3 border-b border-neutral-800">
              <span className="text-xs px-2 py-1 rounded-md bg-neutral-800 text-neutral-400 font-medium">
                {getModeLabel(mode)}
              </span>
            </div>
          )}

          {/* Message Content */}
          <div className="text-neutral-200 whitespace-pre-wrap break-words leading-relaxed">
            {message.content}
          </div>

          {/* Reasoning */}
          {message.reasoning && (
            <div className="mt-4">
              <button
                onClick={() => setShowReasoning(!showReasoning)}
                className="flex items-center gap-2 text-sm text-neutral-400 hover:text-neutral-300 transition-colors"
              >
                {showReasoning ? (
                  <ChevronUp size={16} />
                ) : (
                  <ChevronDown size={16} />
                )}
                <span>Ход рассуждений</span>
              </button>
              {showReasoning && (
                <div className="mt-3 p-3 bg-neutral-800/50 rounded-lg text-sm text-neutral-300 whitespace-pre-wrap border border-neutral-700">
                  {message.reasoning}
                </div>
              )}
            </div>
          )}

          {/* Sources */}
          {message.sources && message.sources.length > 0 && (
            <div className="mt-4 space-y-2">
              <div className="text-sm font-medium text-neutral-400">
                Источники:
              </div>
              {message.sources.map((source, idx) => (
                <a
                  key={idx}
                  href={source.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block p-3 bg-neutral-800/50 rounded-lg hover:bg-neutral-800 transition-colors border border-neutral-700/50 hover:border-neutral-600"
                >
                  <div className="flex items-start gap-2">
                    <ExternalLink
                      size={14}
                      className="text-neutral-500 mt-1 flex-shrink-0"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm text-neutral-200 truncate">
                        {source.title}
                      </div>
                      <div className="text-xs text-neutral-500 mt-1 line-clamp-2">
                        {source.snippet}
                      </div>
                    </div>
                  </div>
                </a>
              ))}
            </div>
          )}

          {/* Timestamp */}
          <div className="text-xs text-neutral-600 mt-3">
            {new Date(message.timestamp * 1000).toLocaleTimeString("ru-RU", {
              hour: "2-digit",
              minute: "2-digit",
            })}
          </div>
        </div>
      </div>
    </div>
  );
}