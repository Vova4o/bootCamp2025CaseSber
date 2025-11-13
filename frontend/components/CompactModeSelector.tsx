'use client';

import { SearchMode } from '@/types';
import { Sparkles, Zap, Brain } from 'lucide-react';

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
      value: 'auto' as SearchMode,
      icon: Sparkles,
      label: 'Auto',
      color: 'blue',
    },
    {
      value: 'simple' as SearchMode,
      icon: Zap,
      label: 'Simple',
      color: 'green',
    },
    {
      value: 'pro' as SearchMode,
      icon: Brain,
      label: 'Pro',
      color: 'purple',
    },
  ];

  return (
    <div className="flex gap-1 bg-gray-800 rounded-xl p-1 h-12">
      {modes.map((m) => {
        const Icon = m.icon;
        const isActive = mode === m.value;
        
        return (
          <button
            key={m.value}
            onClick={() => onModeChange(m.value)}
            disabled={disabled}
            title={m.label}
            className={`relative p-2.5 rounded-lg transition-all flex items-center justify-center ${
              isActive
                ? m.color === 'blue'
                  ? 'bg-blue-600 text-white'
                  : m.color === 'green'
                    ? 'bg-green-600 text-white'
                    : 'bg-purple-600 text-white'
                : 'text-gray-400 hover:text-gray-200 hover:bg-gray-700'
            } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
          >
            <Icon size={18} />
            {isActive && (
              <span className="absolute -top-0.5 -right-0.5 w-2 h-2 bg-green-400 rounded-full border-2 border-gray-800" />
            )}
          </button>
        );
      })}
    </div>
  );
}