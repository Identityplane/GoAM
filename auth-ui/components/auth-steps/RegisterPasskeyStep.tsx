'use client';

import { Button } from '@/components/ui/button';
import { Fingerprint, AlertCircle } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';

interface RegisterPasskeyStepProps {
  onRegister: () => void;
  onSkip: () => void;
  isLoading: boolean;
  error?: string | null;
}

export function RegisterPasskeyStep({
  onRegister,
  onSkip,
  isLoading,
  error,
}: RegisterPasskeyStepProps) {
  return (
    <div className="space-y-6 text-center">
      <div className="flex justify-center mb-4">
        <div className="p-4 bg-primary/10 rounded-full">
          <Fingerprint className="w-12 h-12 text-primary" />
        </div>
      </div>
      
      <div className="space-y-2">
        <h3 className="text-xl font-semibold">Setup a Passkey</h3>
        <p className="text-sm text-muted-foreground">
          Sign in faster and more securely using your device's fingerprint, face
          scan, or screen lock.
        </p>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-3 pt-4">
        <Button
          onClick={onRegister}
          disabled={isLoading}
          className="w-full gap-2"
        >
          {isLoading ? 'Setting up...' : 'Create Passkey'}
        </Button>
        <Button
          variant="outline"
          onClick={onSkip}
          disabled={isLoading}
          className="w-full"
        >
          Not now
        </Button>
      </div>
    </div>
  );
}
