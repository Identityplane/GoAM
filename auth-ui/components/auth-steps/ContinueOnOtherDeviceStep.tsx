'use client';

import { Button } from '@/components/ui/button';
import { AuthStepProps } from './types';

const DEFAULT_MESSAGE =
  'Success — you can continue signing in on your computer. Your browser will finish signing in. Tap continue when you are ready to close this screen.';

export function ContinueOnOtherDeviceStep({ isLoading, onContinue, prompts }: AuthStepProps) {
  const body = prompts?.message || DEFAULT_MESSAGE;

  return (
    <div className="space-y-6 text-center">
      <div className="space-y-2 mb-4">
        <h2 className="text-3xl text-foreground">Continue on your computer</h2>
        <p className="text-muted-foreground text-left whitespace-pre-wrap">{body}</p>
      </div>
      <Button
        type="button"
        className="w-full"
        disabled={isLoading}
        onClick={() => onContinue({ ack: 'true' })}
      >
        {isLoading ? 'Loading…' : 'Continue'}
      </Button>
    </div>
  );
}
