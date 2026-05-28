'use client';

import { useEffect } from 'react';
import { AuthStepProps } from './types';
import { Loader2 } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export function ExternalRedirectStep({ prompts }: AuthStepProps) {
  useEffect(() => {
    const redirectUrl = prompts?.__redirect;
    if (redirectUrl) {
      window.location.href = redirectUrl;
    }
  }, [prompts]);

  return (
    <Card className="border-none shadow-none bg-transparent">
      <CardHeader className="text-center">
        <CardTitle className="text-2xl font-bold">Redirecting...</CardTitle>
        <CardDescription>
          We are redirecting you to the external authentication provider to complete your login.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center justify-center p-6 space-y-4">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground animate-pulse">
          Please wait while we transfer you securely.
        </p>
      </CardContent>
    </Card>
  );
}
