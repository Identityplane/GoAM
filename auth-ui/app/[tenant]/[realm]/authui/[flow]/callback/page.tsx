'use client';

import { useEffect, useState, useRef } from 'react';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { AuthAPI } from '@/lib/auth-api';
import { Loader2 } from 'lucide-react';
import { DebugInspector } from '@/components/debug-inspector';

const DEBUG_STORAGE_KEY = 'goam_debug_execution_id';
const DEBUG_WIDTH_KEY = 'goam_debug_width';

export default function CallbackPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [debugData, setDebugData] = useState<any>(null);
  const [showDebug, setShowDebug] = useState(false);
  const [showDebugSheet, setShowDebugSheet] = useState(false);
  const [debugWidth, setDebugWidth] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem(DEBUG_WIDTH_KEY);
      return saved ? parseInt(saved, 10) : 450;
    }
    return 450;
  });

  const handleWidthChange = (newWidth: number) => {
    setDebugWidth(newWidth);
    localStorage.setItem(DEBUG_WIDTH_KEY, newWidth.toString());
  };
  const [isDebug, setIsDebug] = useState(false);
  const processedRef = useRef(false);

  useEffect(() => {
    if (processedRef.current) return;
    processedRef.current = true;
    
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
        const isDebugUrl = searchParams.has('debug') || window.location.href.includes('debug=true') || window.location.href.includes('?debug');
        const storedDebugId = localStorage.getItem(DEBUG_STORAGE_KEY);
        const activeDebug = isDebugUrl || !!storedDebugId;
        setIsDebug(activeDebug);
        setShowDebugSheet(activeDebug);
        setShowDebug(activeDebug);
        if (activeDebug) {
          setDebugData({ status: 'Processing callback...', tenant, realm, flow });
        }

        // 1. Load current state via GET request (now supports cookie resumption in backend)
        // We use isContinue: true to enforce resumption and error if no session exists.
        const sessionResponse = await AuthAPI.startFlow(backendUrl, flow, activeDebug, true);
        if (sessionResponse.error) {
          setDebugData(sessionResponse.debug || sessionResponse);
          throw new Error(sessionResponse.error.error_description);
        }
        
        // Persist debug if we have executionId
        if (activeDebug && sessionResponse.executionId) {
          localStorage.setItem(DEBUG_STORAGE_KEY, sessionResponse.executionId);
        }

        const resSessionId = sessionResponse.sessionId;
        const currentNode = sessionResponse.currentNode;
        
        if (!resSessionId || !currentNode) {
          throw new Error('Could not resume authentication session.');
        }

        // 2. Submit callback parameters as responses
        const responses: Record<string, string> = {};
        
        // From Query Params
        searchParams.forEach((value, key) => {
          if (key !== 'debug') { // Don't send debug as a response
            responses[key] = value;
          }
        });

        // From Fragment (Hash) - used by some OIDC providers (Implicit Flow)
        if (typeof window !== 'undefined' && window.location.hash) {
          const hash = window.location.hash.substring(1);
          const hashParams = new URLSearchParams(hash);
          hashParams.forEach((value, key) => {
            responses[key] = value;
          });
        }

        const continueResponse = await AuthAPI.continueFlow(backendUrl, flow, {
          sessionId: resSessionId,
          executionId: sessionResponse.executionId!,
          currentNode,
          responses,
        }, activeDebug);

        if (continueResponse.error) {
          setDebugData(continueResponse.debug || continueResponse);
          throw new Error(continueResponse.error.error_description);
        }
        
        setDebugData(continueResponse.debug || continueResponse);

        // 3. Redirect back to the main auth UI page with the session hash
        const debugQuery = activeDebug ? '?debug' : '';
        router.push(`/${tenant}/${realm}/authui/${flow}${debugQuery}#session=${resSessionId}`);
      } catch (err: any) {
        console.error('Callback processing error:', err);
        setError(err.message || 'An error occurred during callback processing.');
        setDebugData((prev: any) => prev || { error: err.message, source: 'callback catch block' });
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
        {showDebug && (
          <div className="fixed bottom-4 right-4 z-50 flex flex-col items-end gap-2">
            <button
              onClick={() => setShowDebugSheet(!showDebugSheet)}
              className="bg-black/50 hover:bg-black/80 text-white rounded-full w-10 h-10 flex items-center justify-center cursor-pointer transition-colors backdrop-blur-sm"
              title="Toggle Debug Panel"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m8 2 1.88 1.88" /><path d="M14.12 3.88 16 2" /><path d="M9 7.13v-1a3.003 3.003 0 1 1 6 0v1" /><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6" /><path d="M12 20v-9" /><path d="M6.53 9C4.6 8.8 3 7.1 3 5" /><path d="M6 13H2" /><path d="M3 21c0-2.1 1.7-3.9 3.8-4" /><path d="M20.97 5c0 2.1-1.6 3.8-3.5 4" /><path d="M22 13h-4" /><path d="M17.2 17c2.1.1 3.8 1.9 3.8 4" /></svg>
            </button>
          </div>
        )}
        {showDebugSheet && (
          <DebugInspector 
            debugData={debugData} 
            onClose={() => setShowDebugSheet(false)} 
            width={debugWidth}
            onWidthChange={handleWidthChange}
          />
        )}
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
      {showDebug && (
        <button
          onClick={() => setShowDebugSheet(!showDebugSheet)}
          className="fixed bottom-4 right-4 z-50 bg-black/50 hover:bg-black/80 text-white rounded-full w-10 h-10 flex items-center justify-center cursor-pointer transition-colors backdrop-blur-sm"
          title="Toggle Debug Panel"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m8 2 1.88 1.88" /><path d="M14.12 3.88 16 2" /><path d="M9 7.13v-1a3.003 3.003 0 1 1 6 0v1" /><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6" /><path d="M12 20v-9" /><path d="M6.53 9C4.6 8.8 3 7.1 3 5" /><path d="M6 13H2" /><path d="M3 21c0-2.1 1.7-3.9 3.8-4" /><path d="M20.97 5c0 2.1-1.6 3.8-3.5 4" /><path d="M22 13h-4" /><path d="M17.2 17c2.1.1 3.8 1.9 3.8 4" /></svg>
        </button>
      )}
      {showDebugSheet && (
        <DebugInspector 
          debugData={debugData} 
          onClose={() => setShowDebugSheet(false)} 
          width={debugWidth}
          onWidthChange={handleWidthChange}
        />
      )}
    </div>
  );
}
