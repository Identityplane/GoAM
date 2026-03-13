'use client';

import { Button } from '@/components/ui/button';
import { CheckCircle2 } from 'lucide-react';
import { AuthStepProps } from './types';

export function SuccessStep({ formData, onRestart }: AuthStepProps) {
  const email = formData.email || '';

  return (
    <div className="space-y-6 text-center">
      <div className="space-y-2 text-center mb-4">
        <h2 className="text-3xl text-foreground">Welcome!</h2>
        <p className="text-muted-foreground">You are all set!</p>
      </div>

      <div className="flex justify-center">
        <CheckCircle2 className="w-16 h-16 text-green-500" />
      </div>

      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Success!
        </h2>
        <p className="text-gray-600 mb-4">
          You have successfully authenticated.
        </p>
        <p className="text-sm text-gray-500">
          Email: <span className="font-medium text-gray-700">{email}</span>
        </p>
      </div>

      <Button onClick={onRestart} className="w-full">
        Start Over
      </Button>
    </div>
  );
}

