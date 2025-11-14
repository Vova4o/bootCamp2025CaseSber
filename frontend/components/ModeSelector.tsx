'use client';

import { SearchMode } from '@/types';

interface ModeSelectorProps {
  mode: SearchMode;
  onModeChange: (mode: SearchMode) => void;
  disabled?: boolean;
}

export default function ModeSelector({
  mode,
  onModeChange,
  disabled = false,
}: ModeSelectorProps) {
  const modes = [
    {
      value: 'auto' as SearchMode,
      label: '🤖 Auto',
      desc: 'Автоматический выбор',
      activeClass: 'border-blue-500 bg-blue-50',
    },
    {
      value: 'simple' as SearchMode,
      label: '⚡ Simple',
      desc: '~2-3 сек',
      activeClass: 'border-green-500 bg-green-50',
    },
    {
      value: 'pro' as SearchMode,
      label: '🧠 Pro',
      desc: 'Глубокий анализ',
      activeClass: 'border-purple-500 bg-purple-50',
    },
  ];

  return (
    <div className="flex gap-2 p-3 bg-gray-50 rounded-xl">
      {modes.map((m) => (
        <button
          key={m.value}
          onClick={() => onModeChange(m.value)}
          disabled={disabled}
          className={`flex-1 p-3 rounded-lg border-2 transition-all ${
            mode === m.value
              ? `${m.activeClass} shadow-sm`
              : 'border-gray-200 bg-white hover:border-gray-300 hover:bg-gray-50'
          } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
        >
          <div className="font-semibold text-sm text-gray-900">
            {m.label}
          </div>
          <div className="text-xs text-gray-600 mt-0.5">{m.desc}</div>
        </button>
      ))}
    </div>
  );
}