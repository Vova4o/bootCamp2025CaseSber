'use client';

import { SearchMode } from '@/types';
import { Sparkles, Zap, Brain, Users, GraduationCap, DollarSign } from 'lucide-react';

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
      description: 'Автоматический выбор (Pro при контексте)',
    },
    {
      value: 'simple' as SearchMode,
      icon: Zap,
      label: 'Simple',
      color: 'green',
      description: 'Быстрый поиск',
    },
    {
      value: 'pro' as SearchMode,
      icon: Brain,
      label: 'Pro',
      color: 'purple',
      description: 'Глубокий анализ',
    },
    {
      value: 'pro-social' as SearchMode,
      icon: Users,
      label: 'Social',
      color: 'pink',
      description: 'Мнения из соцсетей',
    },
    {
      value: 'pro-academic' as SearchMode,
      icon: GraduationCap,
      label: 'Academic',
      color: 'indigo',
      description: 'Научные статьи',
    },
    {
      value: 'pro-finance' as SearchMode,
      icon: DollarSign,
      label: 'Finance',
      color: 'emerald',
      description: 'Финансовые данные',
    },
  ];

  return (
    <div className="flex gap-1 bg-gray-800 rounded-xl p-1">
      {modes.map((m) => {
        const Icon = m.icon;
        const isActive = mode === m.value;
        
        const colorClasses = {
          blue: isActive ? 'bg-blue-600 text-white' : 'text-gray-400',
          green: isActive ? 'bg-green-600 text-white' : 'text-gray-400',
          purple: isActive ? 'bg-purple-600 text-white' : 'text-gray-400',
          pink: isActive ? 'bg-pink-600 text-white' : 'text-gray-400',
          indigo: isActive ? 'bg-indigo-600 text-white' : 'text-gray-400',
          emerald: isActive ? 'bg-emerald-600 text-white' : 'text-gray-400',
        };
        
        return (
          <button
            key={m.value}
            onClick={() => onModeChange(m.value)}
            disabled={disabled}
            title={`${m.label}: ${m.description}`}
            className={`relative p-2.5 rounded-lg transition-all flex items-center justify-center ${
              colorClasses[m.color as keyof typeof colorClasses]
            } ${
              !isActive ? 'hover:text-gray-200 hover:bg-gray-700' : ''
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