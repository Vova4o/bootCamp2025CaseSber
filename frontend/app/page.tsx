"use client";

import { useState } from "react";
import { searchAPI, SearchResponse } from "@/lib/api";
import { Search, Loader2, Clock, ExternalLink } from "lucide-react";

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
        <div className="bg-white rounded-3xl shadow-2xl p-8 mb-8 backdrop-blur-sm bg-opacity-95">
          {/* Mode Selector */}
          <div className="flex gap-4 mb-6 flex-wrap">
            {[
              { value: "auto", label: "🤖 Auto", desc: "Автоматический выбор" },
              { value: "simple", label: "⚡ Simple", desc: "~2-3 сек" },
              { value: "pro", label: "🧠 Pro", desc: "Глубокий анализ" },
            ].map((m) => (
              <label
                key={m.value}
                className={`flex-1 min-w-[150px] cursor-pointer \${
                  mode === m.value ? 'ring-2 ring-blue-500' : ''
                }`}
              >
                <div className="border-2 border-gray-200 rounded-xl p-4 hover:border-blue-300 transition-all">
                  <input
                    type="radio"
                    name="mode"
                    value={m.value}
                    checked={mode === m.value}
                    onChange={(e) => setMode(e.target.value as any)}
                    className="sr-only"
                  />
                  <div className="font-semibold text-center mb-1">
                    {m.label}
                  </div>
                  <div className="text-xs text-gray-500 text-center">
                    {m.desc}
                  </div>
                </div>
              </label>
            ))}
          </div>

          {/* Search Input */}
          <div className="flex gap-4">
            <div className="flex-1 relative">
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyPress={(e) =>
                  e.key === "Enter" && !loading && handleSearch()
                }
                placeholder="Введите ваш вопрос..."
                className="w-full px-6 py-4 text-lg border-2 border-gray-200 rounded-2xl focus:border-blue-500 focus:outline-none focus:ring-4 focus:ring-blue-100 transition-all"
                disabled={loading}
              />
              <Search
                className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400"
                size={24}
              />
            </div>
            <button
              onClick={handleSearch}
              disabled={loading || !query.trim()}
              className="px-8 py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-semibold rounded-2xl hover:from-blue-700 hover:to-indigo-700 disabled:from-gray-400 disabled:to-gray-400 disabled:cursor-not-allowed transition-all shadow-lg hover:shadow-xl flex items-center gap-2"
            >
              {loading ? (
                <>
                  <Loader2 className="animate-spin" size={20} />
                  Ищу...
                </>
              ) : (
                "Поиск"
              )}
            </button>
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="bg-red-50 border-2 border-red-200 rounded-2xl p-6 mb-8">
            <p className="text-red-800 font-medium">❌ {error}</p>
          </div>
        )}

        {/* Results */}
        {result && (
          <div className="space-y-6">
            {/* Answer */}
            <div className="bg-white rounded-3xl shadow-2xl p-8 backdrop-blur-sm bg-opacity-95">
              <div className="flex items-center justify-between mb-6 flex-wrap gap-4">
                <h2 className="text-3xl font-bold text-gray-900">
                  {result.mode === "simple" ? "⚡ Simple Mode" : "🧠 Pro Mode"}
                </h2>
                <div className="flex items-center gap-2 text-sm text-gray-500 bg-gray-100 px-4 py-2 rounded-full">
                  <Clock size={16} />
                  {result.response_time.toFixed(2)}с
                </div>
              </div>

              <div className="prose max-w-none">
                <p className="text-gray-800 whitespace-pre-wrap leading-relaxed text-lg">
                  {result.answer}
                </p>
              </div>
            </div>

            {/* Reasoning Steps */}
            {result.reasoning_steps && result.reasoning_steps.length > 0 && (
              <div className="bg-gradient-to-br from-purple-50 to-pink-50 rounded-3xl shadow-xl p-8">
                <h3 className="text-2xl font-bold text-gray-900 mb-6">
                  🧠 Логика рассуждения
                </h3>
                <ol className="space-y-3">
                  {result.reasoning_steps.map((step, idx) => (
                    <li key={idx} className="flex gap-4 items-start">
                      <span className="flex-shrink-0 w-8 h-8 bg-purple-500 text-white rounded-full flex items-center justify-center font-bold">
                        {idx + 1}
                      </span>
                      <span className="text-gray-700 pt-1">{step}</span>
                    </li>
                  ))}
                </ol>
              </div>
            )}

            {/* Search Queries */}
            {result.search_queries && result.search_queries.length > 1 && (
              <div className="bg-blue-50 rounded-3xl shadow-xl p-8">
                <h3 className="text-2xl font-bold text-gray-900 mb-6">
                  🔎 Поисковые запросы
                </h3>
                <div className="flex flex-wrap gap-3">
                  {result.search_queries.map((q, idx) => (
                    <span
                      key={idx}
                      className="px-5 py-3 bg-white rounded-xl shadow text-sm font-medium text-gray-700 border border-blue-200"
                    >
                      {q}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Sources */}
            <div className="bg-white rounded-3xl shadow-2xl p-8">
              <h3 className="text-2xl font-bold text-gray-900 mb-6">
                📚 Источники
              </h3>
              <div className="space-y-4">
                {result.sources.map((source, idx) => (
                  <div
                    key={idx}
                    className="border-2 border-gray-200 rounded-2xl p-6 hover:shadow-lg hover:border-blue-300 transition-all"
                  >
                    <div className="flex items-start gap-4">
                      <span className="flex-shrink-0 w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center font-bold text-blue-600">
                        {idx + 1}
                      </span>
                      <div className="flex-1 min-w-0">
                        <a
                          href={source.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="font-semibold text-blue-600 hover:text-blue-800 hover:underline flex items-center gap-2 mb-2 text-lg"
                        >
                          {source.title}
                          <ExternalLink size={16} />
                        </a>
                        <p className="text-gray-600 mb-3 line-clamp-3">
                          {source.content}
                        </p>
                        <div className="flex items-center gap-4 text-sm text-gray-500">
                          <span className="truncate flex-1">{source.url}</span>
                          {source.semantic_score && (
                            <span className="flex-shrink-0 bg-green-100 text-green-700 px-3 py-1 rounded-full font-medium">
                              Релевантность:{" "}
                              {(source.semantic_score * 100).toFixed(0)}%
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}
