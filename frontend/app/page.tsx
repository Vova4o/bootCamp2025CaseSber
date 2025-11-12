"use client";

import { useState } from "react";
import { Search, Loader2, Clock, ExternalLink } from "lucide-react";
import { searchAPI } from "@/lib/api";
import type { SearchResponse } from "@/types";

export default function Home() {
  const [query, setQuery] = useState("");
  const [mode, setMode] = useState<"auto" | "simple" | "pro">("auto");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SearchResponse | null>(null);
  const [error, setError] = useState("");

  const handleSearch = async () => {
    if (!query.trim()) return;

    setLoading(true);
    setError("");
    setResult(null);

    try {
      const data = await searchAPI.search({ query, mode });
      setResult(data);
    } catch (err: any) {
      setError(err.response?.data?.detail || "Ошибка при выполнении запроса");
    } finally {
      setLoading(false);
    }
  };

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

        {/* Search Box */}
        <div className="bg-white rounded-3xl shadow-2xl p-8 mb-8">
          {/* Mode Selector */}
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

          {/* Search Input */}
          <div className="flex gap-4">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyPress={(e) => e.key === "Enter" && handleSearch()}
              placeholder="Задайте ваш вопрос..."
              disabled={loading}
              className="flex-1 px-6 py-4 text-lg border-2 border-gray-200 rounded-xl focus:outline-none focus:border-blue-500 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-400"
            />
            <button
              onClick={handleSearch}
              disabled={loading || !query.trim()}
              className="px-8 py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-700 hover:to-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2 shadow-lg"
            >
              {loading ? (
                <>
                  <Loader2 className="animate-spin" size={20} />
                  Поиск...
                </>
              ) : (
                <>
                  <Search size={20} />
                  Поиск
                </>
              )}
            </button>
          </div>
        </div>

        {/* Loading State */}
        {loading && (
          <div className="bg-white rounded-3xl shadow-2xl p-12 text-center">
            <Loader2
              className="animate-spin mx-auto mb-4 text-blue-600"
              size={48}
            />
            <p className="text-xl text-gray-600">
              {mode === "pro" ? "Глубокий анализ..." : "Поиск информации..."}
            </p>
          </div>
        )}

        {/* Error */}
        {error && (
          <div className="bg-red-50 border-2 border-red-200 rounded-2xl p-6 mb-8">
            <p className="text-red-800 font-semibold mb-2">❌ Ошибка</p>
            <p className="text-red-700">{error}</p>
          </div>
        )}

        {/* Results */}
        {result && !loading && (
          <div className="space-y-6">
            {/* Main Answer */}
            <div className="bg-white rounded-3xl shadow-2xl p-8">
              <div className="flex items-center gap-3 mb-6">
                <span
                  className={`px-4 py-2 rounded-full text-sm font-semibold ${
                    result.mode === "simple"
                      ? "bg-green-100 text-green-800"
                      : "bg-purple-100 text-purple-800"
                  }`}
                >
                  {result.mode === "simple" ? "⚡ Simple Mode" : "🧠 Pro Mode"}
                </span>
                <span className="flex items-center gap-2 text-gray-600">
                  <Clock size={16} />
                  <span className="text-sm">
                    {result.processingTime.toFixed(1)}s
                  </span>
                </span>
              </div>

              <h2 className="text-2xl font-bold text-gray-900 mb-4">
                {result.query}
              </h2>

              <div className="prose prose-lg max-w-none">
                <p className="text-gray-800 leading-relaxed whitespace-pre-wrap">
                  {result.answer}
                </p>
              </div>
            </div>

            {/* Reasoning Steps */}
            {result.reasoning && (
              <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-3xl shadow-xl p-8 border-2 border-blue-100">
                <h3 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                  💡 Логика рассуждений
                </h3>
                <p className="text-gray-800 leading-relaxed whitespace-pre-wrap">
                  {result.reasoning}
                </p>
              </div>
            )}

            {/* Sources */}
            {result.sources && result.sources.length > 0 && (
              <div className="bg-white rounded-3xl shadow-2xl p-8">
                <h3 className="text-xl font-bold text-gray-900 mb-6">
                  📚 Источники ({result.sources.length})
                </h3>
                <div className="space-y-4">
                  {result.sources.map((source, idx) => (
                    <a
                      key={idx}
                      href={source.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="block p-6 border-2 border-gray-200 rounded-2xl hover:border-blue-400 hover:bg-blue-50 transition-all group"
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex-1">
                          <h4 className="font-semibold text-gray-900 group-hover:text-blue-600 mb-2 text-lg">
                            {source.title}
                          </h4>
                          <p className="text-gray-700 text-sm mb-2 line-clamp-2">
                            {source.snippet}
                          </p>
                          <p className="text-gray-500 text-xs truncate">
                            {source.url}
                          </p>
                        </div>
                        <div className="flex flex-col items-end gap-2">
                          <ExternalLink
                            size={20}
                            className="text-gray-400 group-hover:text-blue-600"
                          />
                          {source.credibility && (
                            <span className="text-xs font-semibold text-green-600 bg-green-50 px-2 py-1 rounded">
                              {(source.credibility * 100).toFixed(0)}%
                            </span>
                          )}
                        </div>
                      </div>
                    </a>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Info Cards - Show when no results */}
        {!result && !loading && !error && (
          <div className="grid md:grid-cols-2 gap-6">
            <div className="bg-white rounded-3xl shadow-xl p-8 border-2 border-green-100">
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                ⚡ Simple Mode
              </h3>
              <p className="text-gray-700 leading-relaxed">
                Быстрый поиск для простых вопросов. Идеален для получения
                актуальной информации за секунды.
              </p>
            </div>
            <div className="bg-white rounded-3xl shadow-xl p-8 border-2 border-purple-100">
              <h3 className="text-2xl font-bold text-gray-900 mb-4">
                🧠 Pro Mode
              </h3>
              <p className="text-gray-700 leading-relaxed">
                Глубокий анализ с проверкой фактов. Сравнивает источники и
                выдаёт детальный ответ с обоснованием.
              </p>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
