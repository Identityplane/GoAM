'use client';

import { Button } from '@/components/ui/button';
import { CheckCircle2 } from 'lucide-react';
import { useEffect } from 'react';
import { AuthStepProps } from './types';
import { getRedirectUrl } from '@/lib/success-redirect';

export function SuccessStep({ formData, onRestart, result, executionId }: AuthStepProps) {
  const identifier = result?.user_id || formData.email || '';
  const redirectUrl = getRedirectUrl(result);

  useEffect(() => {
    if (!redirectUrl) return;
    window.location.assign(redirectUrl);
  }, [redirectUrl]);

  if (redirectUrl) {
    return (
      <div className="space-y-6 text-center">
        <div className="space-y-2 text-center mb-4">
          <h2 className="text-3xl text-foreground">Redirecting…</h2>
          <p className="text-muted-foreground">Finishing sign-in.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 text-center">
      <div className="space-y-2 text-center mb-4">
        <h2 className="text-3xl text-foreground">Welcome!</h2>
        <p className="text-muted-foreground">You are all set!</p>
      </div>

      <div className="flex justify-center">
        <CheckCircle2 className="w-16 h-16 text-green-500" />
      </div>

      <div className="space-y-4">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">
            Success!
          </h2>
          <p className="text-gray-600">
            You have successfully authenticated.
          </p>
        </div>
        
        <div className="bg-gray-50 rounded-lg p-4 space-y-2 text-left border border-gray-100">
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">User ID:</span>
            <span className="font-mono font-medium text-gray-700 break-all ml-4">{identifier}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-500">Execution ID:</span>
            <span className="font-mono font-medium text-gray-700 break-all ml-4">{executionId}</span>
          </div>
        </div>
      </div>

      <Button onClick={onRestart} className="w-full">
        Start Over
      </Button>
    </div>
  );
}

