'use client';

import { AlertCircle, RotateCcw } from 'lucide-react';
import { Button } from '../ui/button';
import { AuthStepProps } from './types';

export function ErrorStep({ error, onRestart, accentColor, executionId }: AuthStepProps) {
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
      
      <div className="space-y-4">
        <div className="space-y-2">
          <h2 className="text-2xl font-bold tracking-tight text-foreground">Something went wrong</h2>
          <p className="text-muted-foreground">
            {error || "An unexpected error occurred during the authentication process. Please try again or contact support if the problem persists."}
          </p>
        </div>

        {executionId && (
          <div className="bg-destructive/5 rounded-lg p-3 text-left border border-destructive/10">
            <div className="flex justify-between text-xs sm:text-sm items-center">
              <span className="text-muted-foreground">Execution ID:</span>
              <span className="font-mono font-medium text-destructive/80 break-all ml-4 selection:bg-destructive/20">{executionId}</span>
            </div>
          </div>
        )}
      </div>

      <Button
        className="w-full flex items-center justify-center gap-2 group"
        onClick={onRestart}
        style={{ backgroundColor: accentColor }}
      >
        <RotateCcw className="w-4 h-4 transition-transform group-hover:rotate-[-45deg]" />
        Restart Flow
      </Button>
    </div>
  );
}
