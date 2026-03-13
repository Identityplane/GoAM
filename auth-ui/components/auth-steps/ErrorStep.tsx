'use client';

import { AlertCircle, RotateCcw } from 'lucide-react';
import { Button } from '../ui/button';
import { AuthStepProps } from './types';

export function ErrorStep({ error, onContinue, accentColor }: AuthStepProps) {
  return (
    <div className="text-center space-y-6 animate-in fade-in zoom-in duration-300">
      <div className="flex justify-center">
        <div 
          className="p-4 rounded-full bg-destructive/10 text-destructive"
          style={{ color: accentColor || 'var(--destructive)' }}
        >
          <AlertCircle className="w-12 h-12" />
        </div>
      </div>
      
      <div className="space-y-2">
        <h2 className="text-2xl font-bold tracking-tight text-foreground">Something went wrong</h2>
        <p className="text-muted-foreground">
          {error || "An unexpected error occurred during the authentication process. Please try again or contact support if the problem persists."}
        </p>
      </div>

      <Button
        className="w-full flex items-center justify-center gap-2 group"
        onClick={() => onContinue({}, 'login')}
        style={{ backgroundColor: accentColor }}
      >
        <RotateCcw className="w-4 h-4 transition-transform group-hover:rotate-[-45deg]" />
        Restart Flow
      </Button>
    </div>
  );
}
