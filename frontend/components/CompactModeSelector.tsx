"use client";

import { SearchMode } from "@/types";
import { Sparkles, Zap, Brain } from "lucide-react";

interface CompactModeSelectorProps {
  mode: SearchMode;
  onModeChange: (mode: SearchMode) => void;
  disabled?: boolean;
}

export default function CompactModeSelector({
  mode,
  onModeChange,
  disabled = false,
}: CompactModeSelectorProps) {
  const modes = [
    {
      value: "auto" as SearchMode,
      icon: Sparkles,
      label: "Auto",
      description: "Автоматический выбор",
    },
    {
      value: "simple" as SearchMode,
      icon: Zap,
      label: "Simple",
      description: "Быстрый поиск",
    },
    {
      value: "pro" as SearchMode,
      icon: Brain,
      label: "Pro",
      description: "Глубокий анализ",
    },
  ];

  return (
    <div className="flex gap-1.5 bg-neutral-900 rounded-lg p-1 border border-neutral-800">
      {modes.map((m) => {
        const Icon = m.icon;
        const isActive = mode === m.value;

        return (
          <button
            key={m.value}
            onClick={() => onModeChange(m.value)}
            disabled={disabled}
            title={`${m.label}: ${m.description}`}
            className={`relative flex-1 px-3 py-2 rounded-md transition-all flex items-center justify-center gap-2 ${
              isActive
                ? "bg-neutral-800 text-neutral-100"
                : "text-neutral-500 hover:text-neutral-300 hover:bg-neutral-800/50"
            } ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}`}
          >
            <Icon size={16} />
            <span className="text-sm font-medium hidden sm:inline">
              {m.label}
            </span>
          </button>
        );
      })}
    </div>
  );
}
