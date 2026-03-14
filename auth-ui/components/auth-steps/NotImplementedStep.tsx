'use client';

import { Button } from '@/components/ui/button';
import { AlertTriangle } from 'lucide-react';
import { AuthStepProps } from './types';

export function NotImplementedStep({ onRestart, currentStep, currentNode, currentNodeType }: AuthStepProps) {
  return (
    <div className="space-y-6 text-center">
      <div className="space-y-2 text-center mb-4">
        <h2 className="text-3xl text-foreground">Step Not Found</h2>
        <p className="text-muted-foreground">
          The requested authentication step node <code className="bg-muted px-1 rounded">{currentNode || currentStep}</code> 
          (Type: <code className="bg-muted px-1 rounded">{currentNodeType || 'unknown'}</code>) 
          is not yet implemented in this UI.
        </p>
      </div>

      <div className="flex justify-center">
        <div className="bg-amber-100 p-4 rounded-full">
          <AlertTriangle className="w-16 h-16 text-amber-600" />
        </div>
      </div>

      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Oops!
        </h2>
        <p className="text-gray-600 mb-4">
          It looks like the backend is requesting a step that we don't support yet.
        </p>
      </div>

      <Button onClick={onRestart} className="w-full">
        Return to Login
      </Button>
    </div>
  );
}
