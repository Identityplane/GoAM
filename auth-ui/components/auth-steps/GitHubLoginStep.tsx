'use client';

import { useEffect } from 'react';
import { AuthStepProps } from './types';
import { Loader2 } from 'lucide-react';

export function GitHubLoginStep({ prompts, isLoading }: AuthStepProps) {
  const redirectUrl = prompts?.__redirect;

  useEffect(() => {
    if (redirectUrl) {
      // Small delay to show the "Redirecting" message
      const timer = setTimeout(() => {
        window.location.href = redirectUrl;
      }, 800);
      return () => clearTimeout(timer);
    }
  }, [redirectUrl]);

  if (!redirectUrl) {
    return (
      <div className="text-center space-y-4">
        <h2 className="text-xl font-semibold">GitHub Login</h2>
        <p className="text-muted-foreground text-sm">
          No redirect URL provided. Please check the flow configuration.
        </p>
      </div>
    );
  }

  return (
    <div className="text-center space-y-6 py-8">
      <div className="flex justify-center">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
      </div>
      <div className="space-y-2">
        <h2 className="text-xl font-semibold">Redirecting to GitHub</h2>
        <p className="text-muted-foreground text-sm">
          Please wait while we redirect you to authorize with GitHub...
        </p>
      </div>
    </div>
  );
}
