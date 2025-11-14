"use client";

import { ChatSession } from "@/types";
import { MessageSquare, Trash2, Plus, X } from "lucide-react";

interface ChatListProps {
  sessions: ChatSession[];
  currentSessionId?: string;
  onSelectSession: (sessionId: string) => void;
  onDeleteSession: (sessionId: string) => void;
  onNewChat: () => void;
  onClose: () => void;
}

export default function ChatList({
  sessions,
  currentSessionId,
  onSelectSession,
  onDeleteSession,
  onNewChat,
  onClose,
}: ChatListProps) {
  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) return "Сегодня";
    if (days === 1) return "Вчера";
    if (days < 7) return `${days} дней назад`;
    return date.toLocaleDateString("ru-RU");
  };

  const getPreview = (session: ChatSession) => {
    const lastMessage = session.messages[session.messages.length - 1];
    if (!lastMessage) return "Новый чат";
    return lastMessage.role === "user"
      ? lastMessage.content
      : session.messages[session.messages.length - 2]?.content || "Новый чат";
  };

  return (
    <div className="flex flex-col h-full bg-neutral-900 border-r border-neutral-800">
      {/* Header */}
      <div className="flex-shrink-0 p-4 border-b border-neutral-800">
        <div className="flex items-center justify-between mb-3 lg:hidden">
          <h2 className="text-lg font-semibold text-neutral-100">Чаты</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-neutral-800 rounded-lg transition-colors text-neutral-400"
            aria-label="Close sidebar"
          >
            <X size={20} />
          </button>
        </div>
        <button
          onClick={onNewChat}
          className="w-full flex items-center justify-center gap-2 p-3 bg-neutral-100 text-neutral-900 rounded-xl hover:bg-white transition-all font-medium"
        >
          <Plus size={20} />
          <span>Новый чат</span>
        </button>
      </div>

      {/* Chat List */}
      <div className="flex-1 overflow-y-auto p-3">
        {sessions.length === 0 ? (
          <div className="text-center text-neutral-500 mt-8 px-4">
            <div className="w-12 h-12 mx-auto mb-3 rounded-xl bg-neutral-800 flex items-center justify-center">
              <MessageSquare size={24} className="text-neutral-600" />
            </div>
            <p className="text-sm">Нет чатов</p>
          </div>
        ) : (
          <div className="space-y-1.5">
            {sessions
              .sort((a, b) => b.updated_at - a.updated_at)
              .map((session) => (
                <div
                  key={session.id}
                  className={`group relative p-3 rounded-xl cursor-pointer transition-all ${
                    currentSessionId === session.id
                      ? "bg-neutral-800 border border-neutral-700"
                      : "hover:bg-neutral-800/50 border border-transparent"
                  }`}
                  onClick={() => onSelectSession(session.id)}
                >
                  <div className="flex items-start gap-3">
                    <div
                      className={`w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 ${
                        currentSessionId === session.id
                          ? "bg-neutral-700"
                          : "bg-neutral-800"
                      }`}
                    >
                      <MessageSquare
                        size={16}
                        className={
                          currentSessionId === session.id
                            ? "text-neutral-300"
                            : "text-neutral-500"
                        }
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium text-neutral-200 truncate">
                        {getPreview(session)}
                      </div>
                      <div className="text-xs text-neutral-500 mt-1 flex items-center gap-1.5">
                        <span>{formatDate(session.updated_at)}</span>
                        <span>•</span>
                        <span className="capitalize">{session.mode}</span>
                      </div>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onDeleteSession(session.id);
                      }}
                      className="opacity-0 group-hover:opacity-100 p-1.5 hover:bg-neutral-700 rounded-lg transition-all"
                      aria-label="Delete chat"
                    >
                      <Trash2 size={14} className="text-neutral-400" />
                    </button>
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
}