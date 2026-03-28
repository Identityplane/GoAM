'use client';

import { useEffect, useState } from 'react';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { AuthAPI } from '@/lib/auth-api';
import { Loader2 } from 'lucide-react';

export default function CallbackPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const tenant = params?.tenant as string;
    const realm = params?.realm as string;
    const flow = params?.flow as string;

    if (!tenant || !realm || !flow) {
      setError('Missing required parameters for callback processing.');
      return;
    }

    const processCallback = async () => {
      try {
        const baseUrl = process.env.NEXT_PUBLIC_API_URL || '';
        const backendUrl = `${baseUrl}/${tenant}/${realm}`;
        const isDebug = searchParams.has('debug');

        // 1. Load current state via GET request (now supports cookie resumption in backend)
        // We use isContinue: true to enforce resumption and error if no session exists.
        const sessionResponse = await AuthAPI.startFlow(backendUrl, flow, isDebug, true);
        if (sessionResponse.error) {
          throw new Error(sessionResponse.error.error_description);
        }

        const resSessionId = sessionResponse.sessionId;
        const currentNode = sessionResponse.currentNode;
        
        if (!resSessionId || !currentNode) {
          throw new Error('Could not resume authentication session.');
        }

        // 2. Submit callback parameters as responses
        const responses: Record<string, string> = {};
        searchParams.forEach((value, key) => {
          if (key !== 'debug') { // Don't send debug as a response
            responses[key] = value;
          }
        });

        const continueResponse = await AuthAPI.continueFlow(backendUrl, flow, {
          sessionId: resSessionId,
          executionId: sessionResponse.executionId!,
          currentNode,
          responses,
        }, isDebug);

        if (continueResponse.error) {
          throw new Error(continueResponse.error.error_description);
        }

        // 3. Redirect back to the main auth UI page with the session hash
        router.push(`/${tenant}/${realm}/authui/${flow}#session=${resSessionId}${isDebug ? '?debug' : ''}`);
      } catch (err: any) {
        console.error('Callback processing error:', err);
        setError(err.message || 'An error occurred during callback processing.');
      }
    };

    processCallback();
  }, [params, searchParams, router]);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8 text-center space-y-4">
          <div className="text-red-500 text-5xl mb-4">⚠️</div>
          <h1 className="text-2xl font-bold text-gray-900">Authentication Error</h1>
          <p className="text-gray-600">{error}</p>
          <button
            onClick={() => window.location.href = `/${params?.tenant}/${params?.realm}/authui/${params?.flow}`}
            className="mt-6 w-full py-2 px-4 bg-primary text-primary-foreground rounded-md hover:opacity-90 transition-opacity"
          >
            Return to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="flex flex-col items-center gap-4">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground animate-pulse text-lg">
          Finalizing authentication...
        </p>
      </div>
    </div>
  );
}
