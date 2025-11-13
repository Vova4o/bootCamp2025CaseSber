'use client';

import { ChatSession } from '@/types';
import { MessageSquare, Trash2, Plus } from 'lucide-react';

interface ChatListProps {
  sessions: ChatSession[];
  currentSessionId?: string;
  onSelectSession: (sessionId: string) => void;
  onDeleteSession: (sessionId: string) => void;
  onNewChat: () => void;
}

export default function ChatList({
  sessions,
  currentSessionId,
  onSelectSession,
  onDeleteSession,
  onNewChat,
}: ChatListProps) {
  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) return 'Сегодня';
    if (days === 1) return 'Вчера';
    if (days < 7) return `${days} дней назад`;
    return date.toLocaleDateString('ru-RU');
  };

  const getPreview = (session: ChatSession) => {
    const lastMessage = session.messages[session.messages.length - 1];
    if (!lastMessage) return 'Новый чат';
    return lastMessage.role === 'user'
      ? lastMessage.content
      : session.messages[session.messages.length - 2]?.content || 'Новый чат';
  };

  return (
    <div className="flex flex-col h-full bg-gray-900 border-r border-gray-800">
      {/* Header */}
      <div className="p-4 border-b border-gray-800">
        <button
          onClick={onNewChat}
          className="w-full flex items-center justify-center gap-2 p-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl hover:from-blue-500 hover:to-indigo-500 transition-all shadow-lg hover:shadow-xl"
        >
          <Plus size={20} />
          <span className="font-semibold">Новый чат</span>
        </button>
      </div>

      {/* Chat List */}
      <div className="flex-1 overflow-y-auto p-2">
        {sessions.length === 0 ? (
          <div className="text-center text-gray-500 mt-8">
            <MessageSquare size={48} className="mx-auto mb-3 opacity-30" />
            <p className="text-sm">Нет чатов</p>
          </div>
        ) : (
          <div className="space-y-1">
            {sessions
              .sort((a, b) => b.updated_at - a.updated_at)
              .map((session) => (
                <div
                  key={session.id}
                  className={`group relative p-3 rounded-lg cursor-pointer transition-all ${
                    currentSessionId === session.id
                      ? 'bg-gray-800 shadow-md border border-blue-500/50'
                      : 'hover:bg-gray-800/50'
                  }`}
                  onClick={() => onSelectSession(session.id)}
                >
                  <div className="flex items-start gap-2">
                    <MessageSquare
                      size={16}
                      className={`mt-1 flex-shrink-0 ${
                        currentSessionId === session.id
                          ? 'text-blue-400'
                          : 'text-gray-500'
                      }`}
                    />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium text-gray-200 truncate">
                        {getPreview(session)}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {formatDate(session.updated_at)} •{' '}
                        <span className="capitalize">{session.mode}</span>
                      </div>
                    </div>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onDeleteSession(session.id);
                      }}
                      className="opacity-0 group-hover:opacity-100 p-1.5 hover:bg-red-500/20 rounded transition-opacity"
                    >
                      <Trash2 size={14} className="text-red-400" />
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