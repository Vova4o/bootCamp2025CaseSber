"use client";

import { useState } from "react";
import ChatInterface from "@/components/ChatInterface";
import { MessageSquare, Zap } from "lucide-react";
import type { SearchMode } from "@/types";

export default function Home() {
  const [mode, setMode] = useState<SearchMode>("auto");
  const [chatMode, setChatMode] = useState(false);

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="container mx-auto px-4 py-12 max-w-6xl">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-6xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent mb-4">
            🔍 Research Pro Mode
          </h1>
          <p className="text-xl text-gray-600">
            Умный поисковый ассистент с проверкой фактов
          </p>
        </div>

        {/* Mode Toggle */}
        <div className="bg-white rounded-3xl shadow-2xl p-8 mb-8">
          <div className="flex gap-4 mb-6 flex-wrap">
            {[
              { value: "auto", label: "🤖 Auto", desc: "Автоматический выбор" },
              { value: "simple", label: "⚡ Simple", desc: "~2-3 сек" },
              { value: "pro", label: "🧠 Pro", desc: "Глубокий анализ" },
            ].map((m) => (
              <button
                key={m.value}
                onClick={() => setMode(m.value as any)}
                className={`flex-1 min-w-[150px] p-4 rounded-xl border-2 transition-all ${
                  mode === m.value
                    ? "border-blue-500 bg-blue-50 shadow-md"
                    : "border-gray-200 bg-white hover:border-gray-300 hover:bg-gray-50"
                }`}
              >
                <div className="font-semibold text-gray-900 mb-1">
                  {m.label}
                </div>
                <div className="text-sm text-gray-600">{m.desc}</div>
              </button>
            ))}
          </div>

          {/* Interface Type Toggle */}
          <div className="flex gap-4">
            <button
              onClick={() => setChatMode(false)}
              className={`flex-1 p-4 rounded-xl border-2 transition-all ${
                !chatMode
                  ? "border-blue-500 bg-blue-50"
                  : "border-gray-200 hover:border-gray-300"
              }`}
            >
              <div className="flex items-center justify-center gap-2 text-gray-900">
                <Zap size={20} />
                <span className="font-semibold">Разовый поиск</span>
              </div>
            </button>
            <button
              onClick={() => setChatMode(true)}
              className={`flex-1 p-4 rounded-xl border-2 transition-all ${
                chatMode
                  ? "border-blue-500 bg-blue-50"
                  : "border-gray-200 hover:border-gray-300"
              }`}
            >
              <div className="flex items-center justify-center gap-2 text-gray-900">
                <MessageSquare size={20} />
                <span className="font-semibold">Чат с контекстом</span>
              </div>
            </button>
          </div>
        </div>

        {/* Content */}
        {chatMode ? (
          <ChatInterface mode={mode} />
        ) : (
          <div className="grid md:grid-cols-2 gap-6">
            <div className="bg-white rounded-3xl shadow-xl p-8 border-2 border-green-100">
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                ⚡ Разовый поиск
              </h3>
              <p className="text-gray-700 leading-relaxed">
                Быстрый поиск для получения ответа на один вопрос без сохранения
                контекста.
              </p>
            </div>
            <div className="bg-white rounded-3xl shadow-xl p-8 border-2 border-purple-100">
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                💬 Чат с контекстом
              </h3>
              <p className="text-gray-700 leading-relaxed">
                Диалог с сохранением истории. Система учитывает предыдущие
                вопросы и ответы.
              </p>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
