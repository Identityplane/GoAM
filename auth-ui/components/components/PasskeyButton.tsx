'use client';

import { Button } from '@/components/ui/button';

interface PasskeyButtonProps {
  settings: {
    enable_passkey?: boolean;
    accentColor?: string;
  } | null;
  onAction?: (action: string) => void;
  isLoading?: boolean;
}

export function PasskeyButton({ settings, onAction, isLoading }: PasskeyButtonProps) {
  if (!settings || !settings.enable_passkey) return null;

  return (
    <Button
      variant="outline"
      disabled={isLoading}
      onClick={() => onAction?.('passkeys')}
      className="h-12 border-gray-200 hover:bg-gray-50 hover:text-gray-900 rounded-lg bg-white shadow-none cursor-pointer w-full text-base font-medium flex items-center justify-center gap-3"
    >
      <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 2l-2 2" />
        <circle cx="7" cy="13" r="5" />
        <path d="M11 9l10 10l-2 2l-2-2l-2 2l-2-2l-2 2l-4-4" />
      </svg>
      Sign in with Passkey
    </Button>
  );
}
