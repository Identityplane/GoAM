'use client';

import { Button } from '@/components/ui/button';
import { Key } from 'lucide-react';

interface LoginPasskeyStepProps {
  onLogin: () => void;
  onAlternativeLogin: () => void;
  isLoading: boolean;
  email?: string;
}

export function LoginPasskeyStep({
  onLogin,
  onAlternativeLogin,
  isLoading,
  email,
}: LoginPasskeyStepProps) {
  return (
    <div className="space-y-6 text-center">
      <div className="flex justify-center mb-4">
        <div className="p-4 bg-primary/10 rounded-full">
          <Key className="w-12 h-12 text-primary" />
        </div>
      </div>
      
      <div className="space-y-2">
        <h3 className="text-xl font-semibold">Sign in with Passkey</h3>
        <p className="text-sm text-muted-foreground">
          {email ? `Use your passkey for ${email}` : 'Use your biometric authentication to sign in securely.'}
        </p>
      </div>

      <div className="space-y-3 pt-4">
        <Button
          onClick={onLogin}
          disabled={isLoading}
          className="w-full gap-2"
        >
          <Key className="w-4 h-4" />
          {isLoading ? 'Waiting for authentication...' : 'Sign in with Passkey'}
        </Button>
        <Button
          variant="ghost"
          onClick={onAlternativeLogin}
          disabled={isLoading}
          className="w-full"
        >
          Use password instead
        </Button>
      </div>
    </div>
  );
}
