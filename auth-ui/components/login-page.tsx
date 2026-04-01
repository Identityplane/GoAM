'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { ArrowLeft } from 'lucide-react';
import { StepRegistry } from '@/components/auth-steps';
import { AuthAPI } from '@/lib/auth-api';
import type { AuthStep, StepConfig, FlowInfo, MetadataResponse } from '@/lib/auth-api';
import { cn } from '@/lib/utils';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { DebugInspector } from './debug-inspector';
const DEBUG_STORAGE_KEY = 'goam_debug_execution_id';
const DEBUG_WIDTH_KEY = 'goam_debug_width';

// Helper to parse strings from settings into correct types (boolean, numbers etc)
const parseSettings = (settings: Record<string, string>) => {
  const parsed: Record<string, any> = {};
  for (const [key, value] of Object.entries(settings)) {
    if (value === 'true') parsed[key] = true;
    else if (value === 'false') parsed[key] = false;
    else if (!isNaN(Number(value)) && value.trim() !== '') parsed[key] = Number(value);
    else parsed[key] = value;
  }
  return parsed;
};

export default function LoginPage(): React.ReactElement | null {
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const tenantParam = params?.tenant as string;
  const realmParam = params?.realm as string;
  const flowParam = params?.flow as string;
  const isDebug = searchParams.has('debug');
  const isDynamicRoute = !!tenantParam && !!realmParam;

  const [currentStep, setCurrentStep] = useState<AuthStep>('login');
  const [executionId, setExecutionId] = useState<string | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [currentNode, setCurrentNode] = useState<string | null>(null);
  const [currentNodeType, setCurrentNodeType] = useState<string | null>(null);
  const [prompts, setPrompts] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<any>(null);
  const [debugWidth, setDebugWidth] = useState(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem(DEBUG_WIDTH_KEY);
      return saved ? parseInt(saved, 10) : 450;
    }
    return 450;
  });

  const handleWidthChange = useCallback((newWidth: number) => {
    setDebugWidth(newWidth);
    localStorage.setItem(DEBUG_WIDTH_KEY, newWidth.toString());
  }, []);

  const [settings, setSettings] = useState<{
    backgroundColor: string;
    accentColor: string;
    logoSvg?: string;
    logoName?: string;
    privacyPolicyUrl?: string;
    sidebarTitle?: string;
    sidebarText?: string;
    fontFamily?: string;
    primaryButtonColor?: string;
    primaryButtonHoverColor?: string;
    copyrightText?: string;
    show_sidebar?: boolean;
    pageBackgroundColor?: string;
    inputBackgroundColor?: string;
    backend_url?: string;
    enable_register?: boolean;
    [key: string]: any;
  } | null>(null);

  const [metadata, setMetadata] = useState<MetadataResponse | null>(null);
  const [selectedFlow, setSelectedFlow] = useState<string>('');
  const [debugData, setDebugData] = useState<any>(null);
  const [showDebugSheet, setShowDebugSheet] = useState(isDebug);

  // Debug panel state
  const [configName, setConfigName] = useState('acme');
  const [showDebug, setShowDebug] = useState(isDebug);

  const lastInitRef = useRef<string>('');
  const lastStartRef = useRef<string>('');
  const isTransitioningRef = useRef(false);

  const handleStartFlow = async (flowOverride?: string) => {
    const flowToStart = flowOverride || selectedFlow;
    if (!settings?.backend_url || !flowToStart || isTransitioningRef.current) return;
    isTransitioningRef.current = true;

    setError(null);
    setIsLoading(true);

    try {
      if (lastStartRef.current === flowToStart) return;
      lastStartRef.current = flowToStart;
      
      const flowResponse = await AuthAPI.startFlow(settings.backend_url, flowToStart, isDebug, false, true);

      if (flowResponse.error) {
        setError(flowResponse.error.error_description);
        setCurrentStep('error');
        setExecutionId(flowResponse.executionId || null);
        setDebugData(flowResponse.debug || null);
        if (isDebug) setShowDebugSheet(true);
        return;
      }

      if (flowResponse.currentNodeType || flowResponse.currentNode) {
        setExecutionId(flowResponse.executionId || null);
        setSessionId(flowResponse.sessionId || null);
        setCurrentNode(flowResponse.currentNode || null);
        setCurrentNodeType(flowResponse.currentNodeType || null);
        setPrompts(flowResponse.prompts || {});
        setCurrentStep((flowResponse.currentNodeType || flowResponse.currentNode) as AuthStep);
        setDebugData(flowResponse.debug || null);

        // Persist debug state if enabled
        if (isDebug && flowResponse.executionId) {
          localStorage.setItem(DEBUG_STORAGE_KEY, flowResponse.executionId);
        }

        // Update URL if flow picked from debug utility is different from current URL
        if (flowToStart !== flowParam) {
           const debugQuery = isDebug ? '?debug' : '';
           router.push(`/${tenantParam}/${realmParam}/authui/${flowToStart}${debugQuery}`);
        }
      }

    } catch (err: any) {
      setError(err.message || 'Failed to start flow');
    } finally {
      setIsLoading(false);
      isTransitioningRef.current = false;
    }
  };

  // Form data accumulation
  const [formData, setFormData] = useState<Record<string, any>>({
    email: '',
    password: '',
    otp: '',
  });

  const handleContinue = useCallback(
    async (stepData: Record<string, string | boolean>, action?: string) => {
      if (isTransitioningRef.current) return;
      isTransitioningRef.current = true;

      setError(null);
      setIsLoading(true);

      try {
        if (settings?.backend_url && selectedFlow && executionId && sessionId) {
          // Map stepData to responses (converting booleans to strings as expected by FlowRequest)
          const responses: Record<string, string> = {};
          for (const [key, value] of Object.entries(stepData)) {
            responses[key] = String(value);
          }

          if (action) {
            responses['option'] = action;
          }

          const flowResponse = await AuthAPI.continueFlow(settings.backend_url, selectedFlow, {
            executionId,
            sessionId,
            currentNode: currentNode as string,
            responses,
          }, isDebug);

          if (flowResponse.error) {
            setError(flowResponse.error.error_description);
            setCurrentStep('error');
            setExecutionId(flowResponse.executionId || null);
            setDebugData(flowResponse.debug || null);
            if (isDebug) setShowDebugSheet(true);
            return;
          }

          if (flowResponse.result?.success) {
            setResult(flowResponse.result);
            setExecutionId(flowResponse.executionId || null);
            setDebugData(flowResponse.debug || null);
            if (flowResponse.result.redirect) {
              window.location.href = flowResponse.result.redirect;
            } else {
              setCurrentStep('success');
            }
          } else if (flowResponse.currentNodeType || flowResponse.currentNode) {
            setExecutionId(flowResponse.executionId || null);
            setSessionId(flowResponse.sessionId || null);
            setCurrentNode(flowResponse.currentNode || null);
            setCurrentNodeType(flowResponse.currentNodeType || null);
            setPrompts(flowResponse.prompts || {});
            setCurrentStep((flowResponse.currentNodeType || flowResponse.currentNode) as AuthStep);
            setDebugData(flowResponse.debug || null);

            // Persist debug state if enabled
            if (isDebug && flowResponse.executionId) {
              localStorage.setItem(DEBUG_STORAGE_KEY, flowResponse.executionId);
            }

            if (flowResponse.errorMessage) {
              setError(flowResponse.errorMessage);
            }
          }
        } else {
          // Fallback to internal mock
          const nextStepConfig = await AuthAPI.processInternalStep(currentStep, stepData, action);
          setCurrentStep(nextStepConfig.step);
        }

        // Update form data with current values
        setFormData((prev) => ({
          ...prev,
          ...stepData,
        }));
      } catch (err: any) {
        setError(err.message || 'An error occurred');
        console.error('Auth error:', err);
      } finally {
        setIsLoading(false);
        isTransitioningRef.current = false;
      }
    },
    [currentStep, settings, selectedFlow, executionId, sessionId]
  );

  const handleRestart = () => {
    setFormData({
      email: '',
      password: '',
      otp: '',
    });
    setError(null);
    setResult(null);
    lastStartRef.current = '';

    if (settings?.backend_url && selectedFlow) {
      handleStartFlow(selectedFlow);
    } else {
      setCurrentStep('login');
    }
  };

  useEffect(() => {
    const initKey = `${configName}-${isDynamicRoute}-${tenantParam}-${realmParam}-${flowParam}`;
    if (lastInitRef.current === initKey) return;
    lastInitRef.current = initKey;

    setSettings(null); // Reset settings to show loading state on config change
    setMetadata(null);
    setCurrentStep('login'); // Reset to default step
    setError(null);

    // If we're on a dynamic route, construct backend URL directly.
    // Otherwise, fallback to the local dev API.
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || '';
    const configPromise = isDynamicRoute
      ? Promise.resolve({ backend_url: baseUrl ? `${baseUrl}/${tenantParam}/${realmParam}` : `/${tenantParam}/${realmParam}` })
      : fetch(`/api/settings?config=${configName}`).then((res) => res.json());

    configPromise
      .then(async (data) => {
        let finalSettings = { ...data };

        // Check for session ID in URL fragment
        const hash = window.location.hash;
        const sessionMatch = hash.match(/#session=([a-zA-Z0-9-]+)/);
        const resumeSessionId = sessionMatch ? sessionMatch[1] : null;

        // If mock is true, we skip all backend calls and just show the UI
        if (data.mock === true) {
          setSettings(finalSettings);
          return;
        }

        if (data.backend_url) {
          try {
            const meta = await AuthAPI.fetchMetadata(data.backend_url, isDebug);
            setMetadata(meta);

            // Merge realm settings if present
            if (meta.realm?.settings) {
              const parsedRealmSettings = parseSettings(meta.realm.settings);
              finalSettings = { ...finalSettings, ...parsedRealmSettings };
            }

            // If we have a session ID to resume, do that first
            if (resumeSessionId) {
              setIsLoading(true);
              try {
                // Check if we should enable debug mode based on stored execution ID
                let shouldEnableDebug = isDebug;
                const storedDebugId = localStorage.getItem(DEBUG_STORAGE_KEY);
                
                const flowResponse = await AuthAPI.resumeSession(data.backend_url, resumeSessionId, isDebug || !!storedDebugId);
                
                if (flowResponse.error) {
                  setError(flowResponse.error.error_description);
                  setCurrentStep('error');
                  setExecutionId(flowResponse.executionId || null);
                  setDebugData(flowResponse.debug || null);
                  if (isDebug || !!storedDebugId) setShowDebugSheet(true);
                } else if (flowResponse.currentNodeType || flowResponse.currentNode) {
                  setExecutionId(flowResponse.executionId || null);
                  setSessionId(flowResponse.sessionId || null);
                  setCurrentNode(flowResponse.currentNode || null);
                  setCurrentNodeType(flowResponse.currentNodeType || null);
                  setPrompts(flowResponse.prompts || {});
                  setCurrentStep((flowResponse.currentNodeType || flowResponse.currentNode) as AuthStep);
                  setResult(flowResponse.result || null);
                  setDebugData(flowResponse.debug || null);
                  
                  // If we resumed with a matching debug ID, ensure debug mode is active in URL
                  if (storedDebugId === flowResponse.executionId && !isDebug) {
                    const params = new URLSearchParams(window.location.search);
                    params.set('debug', 'true');
                    window.history.replaceState(null, '', `${window.location.pathname}?${params.toString()}${window.location.hash}`);
                    // Force a local state update since searchParams won't react to replaceState immediately for derived 'isDebug'
                    setShowDebug(true);
                    setShowDebugSheet(true);
                  } else if (isDebug && flowResponse.executionId) {
                    localStorage.setItem(DEBUG_STORAGE_KEY, flowResponse.executionId);
                  }
                  
                  if (flowResponse.flow) {
                    setSelectedFlow(flowResponse.flow);
                  }
                  
                  // Clear the hash to avoid multiple resumptions using a cleaner method
                  window.history.replaceState(null, '', window.location.pathname + window.location.search);
                }
              } catch (err: any) {
                console.error('Failed to resume session:', err);
                setError(err.message || 'Failed to resume session');
              } finally {
                setIsLoading(false);
              }
            } else if (flowParam || meta.flows?.length > 0) {
              const flowToStart = flowParam || meta.flows[0].route;
              setSelectedFlow(flowToStart);
 
              // Automatically start the flow
              setIsLoading(true);
              try {
                const flowResponse = await AuthAPI.startFlow(data.backend_url, flowToStart, isDebug, false, true);
                if (flowResponse.error) {
                  setError(flowResponse.error.error_description);
                  setCurrentStep('error');
                  setDebugData(flowResponse.debug || null);
                  if (isDebug) setShowDebugSheet(true);
                } else if (flowResponse.currentNodeType || flowResponse.currentNode) {
                  setExecutionId(flowResponse.executionId || null);
                  setSessionId(flowResponse.sessionId || null);
                  setCurrentNode(flowResponse.currentNode || null);
                  setCurrentNodeType(flowResponse.currentNodeType || null);
                  setPrompts(flowResponse.prompts || {});
                  setCurrentStep((flowResponse.currentNodeType || flowResponse.currentNode) as AuthStep);
                  setResult(flowResponse.result || null);
                  setDebugData(flowResponse.debug || null);
                }
              } catch (err: any) {
                console.error('Failed to start flow:', err);
                setError(err.message || 'Failed to start flow');
              } finally {
                setIsLoading(false);
              }
            }
          } catch (err) {
            console.error('Failed to load metadata:', err);
            setError('Failed to connect to authentication backend');
          }
        }

        setSettings(finalSettings);
      })
      .catch((err) => {
        console.error('Failed to load settings:', err);
        setError('Failed to load application settings');
      });
  }, [configName, isDynamicRoute, tenantParam, realmParam, flowParam]);

  if (!settings) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
          <p className="text-sm text-muted-foreground animate-pulse">Loading settings...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex font-sans" style={settings?.fontFamily ? { fontFamily: settings.fontFamily } : {}}>
      {settings && (
        <style dangerouslySetInnerHTML={{
          __html: `
          :root {
            ${settings.primaryButtonColor ? `--primary: ${settings.primaryButtonColor} !important;` : ''}
            --primary-hover: ${settings.primaryButtonHoverColor || settings.primaryButtonColor || 'var(--primary)'};
            ${settings.inputBackgroundColor ? `--input-bg: ${settings.inputBackgroundColor} !important;` : ''}
          }

          /* Target only standard input types, excluding hidden ones or those specific to OTP */
          [data-slot="input"], 
          [data-slot="input-otp-slot"],
          select, 
          textarea {
            ${settings.inputBackgroundColor ? `background-color: var(--input-bg) !important;` : ''}
          }

          /* Ensure hidden OTP input remains hidden and unstyled */
          [data-slot="input-otp"] input {
            background-color: transparent !important;
            border: none !important;
          }
        `}} />
      )}
      {/* Debug Controls */}
      {isDebug && (
        <div className="fixed bottom-4 right-4 z-[60] flex flex-col items-end gap-2">
          {/* Main Toggle Button */}
          <button
            onClick={() => setShowDebug(!showDebug)}
            className="bg-black/60 hover:bg-black/80 text-white rounded-full w-10 h-10 flex items-center justify-center cursor-pointer transition-all backdrop-blur-md border border-white/20 shadow-lg"
            title="Toggle Debug Panel"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m8 2 1.88 1.88" /><path d="M14.12 3.88 16 2" /><path d="M9 7.13v-1a3.003 3.003 0 1 1 6 0v1" /><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6" /><path d="M12 20v-9" /><path d="M6.53 9C4.6 8.8 3 7.1 3 5" /><path d="M6 13H2" /><path d="M3 21c0-2.1 1.7-3.9 3.8-4" /><path d="M20.97 5c0 2.1-1.6 3.8-3.5 4" /><path d="M22 13h-4" /><path d="M17.2 17c2.1.1 3.8 1.9 3.8 4" /></svg>
          </button>

          {/* Quick Inspector Button (Visible even when panel is closed) */}
          {!showDebug && (
            <button
              onClick={() => setShowDebugSheet(!showDebugSheet)}
              className={cn(
                "px-3 py-1.5 rounded-full text-[10px] font-bold uppercase tracking-wider transition-all shadow-lg border backdrop-blur-md",
                showDebugSheet 
                  ? "bg-primary text-primary-foreground border-primary" 
                  : "bg-white/80 hover:bg-white text-slate-700 border-slate-200"
              )}
            >
              {showDebugSheet ? "✕ Close Inspector" : "🔍 Inspect"}
            </button>
          )}

          {/* Debug Panel Details */}
          {showDebug && (
            <div className="w-64 bg-white/95 backdrop-blur-md rounded-xl shadow-2xl border border-slate-200 p-4 animate-in slide-in-from-top-4 duration-200" style={{ fontFamily: 'sans-serif' }}>
              <div className="text-sm font-bold text-slate-800 mb-3 border-b border-slate-100 pb-2 flex justify-between items-center">
                <span className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" /> Flow Debugger</span>
                <button onClick={() => setShowDebug(false)} className="p-1 hover:bg-slate-100 rounded-md transition-colors text-slate-400">✕</button>
              </div>

              <div className="space-y-4 text-sm">
                <div>
                  <label className="block text-slate-500 mb-1.5 text-[10px] uppercase font-bold tracking-tight">UI Theme</label>
                  <select
                    className="w-full border border-slate-200 rounded-lg p-2 bg-slate-50/50 focus:ring-2 focus:ring-primary/20 outline-none transition-all text-xs"
                    value={configName}
                    onChange={(e) => setConfigName(e.target.value)}
                  >
                    <option value="identityplane">IdentityPlane</option>
                    <option value="acme">Acme (External Backend)</option>
                    <option value="blue">Blue Theme</option>
                    <option value="default">Light Gray</option>
                    <option value="forest">Forest Green</option>
                    <option value="sunset">Sunset Orange</option>
                    <option value="minimal">Minimal (No Sidebar)</option>
                  </select>
                </div>

                {settings?.backend_url && metadata && (
                  <div className="pt-2 border-t border-slate-100">
                    <label className="block text-primary mb-1.5 text-[10px] uppercase font-bold tracking-tight">Active Flow</label>
                    <div className="flex gap-2">
                      <select
                        className="flex-1 border border-slate-200 rounded-lg p-2 bg-slate-50/50 focus:ring-2 focus:ring-primary/20 outline-none truncate transition-all text-xs"
                        value={selectedFlow}
                        onChange={(e) => {
                          const newFlow = e.target.value;
                          setSelectedFlow(newFlow);
                          handleStartFlow(newFlow);
                        }}
                      >
                        {metadata.flows.map(flow => (
                          <option key={flow.id} value={flow.route}>
                            {flow.id}
                          </option>
                        ))}
                      </select>
                      <button
                        onClick={() => handleStartFlow()}
                        disabled={isLoading}
                        className="bg-primary hover:opacity-90 text-primary-foreground w-9 flex-none rounded-lg flex items-center justify-center disabled:opacity-50 transition-all active:scale-95 shadow-sm"
                        title="Restart Flow"
                      >
                        🚀
                      </button>
                    </div>
                  </div>
                )}

                <div>
                  <label className="block text-slate-500 mb-1.5 text-[10px] uppercase font-bold tracking-tight">Fast-Forward</label>
                  <select
                    className="w-full border border-slate-200 rounded-lg p-2 bg-slate-50/50 focus:ring-2 focus:ring-primary/20 outline-none transition-all text-xs"
                    value={currentStep}
                    onChange={(e) => {
                      setError(null);
                      setCurrentStep(e.target.value as AuthStep);
                    }}
                  >
                    <option value="login">1. Login</option>
                    <option value="register">1. Register</option>
                    <option value="password">2. Password</option>
                    <option value="otp">3. OTP</option>
                    <option value="terms">4. Terms</option>
                    <option value="success">5. Success</option>
                  </select>
                </div>

                <div className="pt-2 border-t border-slate-100">
                  <button
                    onClick={() => setShowDebugSheet(!showDebugSheet)}
                    className={cn(
                      "w-full py-2 rounded-lg transition-all text-xs font-bold border flex items-center justify-center gap-2 shadow-sm",
                      showDebugSheet 
                        ? "bg-primary text-primary-foreground border-primary" 
                        : "bg-slate-900 text-white border-slate-900 hover:bg-slate-800"
                    )}
                  >
                    {showDebugSheet ? "✕ Close Inspector" : "🔍 View Data"}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      )}



      <div className={cn(
        "hidden relative overflow-hidden transition-all duration-300 shrink-0",
        (settings?.show_sidebar && !showDebugSheet) ? "lg:flex lg:w-1/2" : "w-0"
      )} style={{ backgroundColor: settings?.backgroundColor || '#374151' }}>
        <div className="relative z-10 flex flex-col justify-between w-full px-12 py-12">
          <div className="flex items-center">
            {settings?.logoSvg !== undefined ? (
              settings.logoSvg ? (
                <div
                  className="w-8 h-8 mr-3 flex items-center justify-center"
                  style={{ color: settings.accentColor }}
                  dangerouslySetInnerHTML={{ __html: settings.logoSvg }}
                />
              ) : null
            ) : (
              <div className="w-8 h-8 bg-white rounded-lg flex items-center justify-center mr-3">
                <div className="w-4 h-4 rounded-sm" style={{ backgroundColor: settings?.accentColor || '#3F3FF3' }}></div>
              </div>
            )}
            {settings?.logoName !== undefined ? (
              settings.logoName ? (
                <h1 className="text-xl font-semibold text-white">{settings.logoName}</h1>
              ) : null
            ) : (
              <h1 className="text-xl font-semibold text-white">GoAM</h1>
            )}
          </div>

          <div className="flex-1 flex flex-col justify-center">
            <h2 className="text-4xl text-white mb-6 leading-tight">
              {settings?.sidebarTitle || ''}
            </h2>
            <p className="text-white/90 text-lg leading-relaxed">
              {settings?.sidebarText || ''}
            </p>
          </div>

          <div className="flex justify-between items-center text-white/70 text-sm">
            <span>{settings?.copyrightText || ''}</span>
            {settings?.privacyPolicyUrl ? (
              <a href={settings.privacyPolicyUrl} className="hover:text-white/90 cursor-pointer">Privacy Policy</a>
            ) : (
              <span className="cursor-pointer hover:text-white/90">Privacy Policy</span>
            )}
          </div>
        </div>
      </div>

      <div className={cn(
        "flex items-center justify-center p-8 relative grow transition-[width] duration-300",
      )} style={{ 
        backgroundColor: settings?.pageBackgroundColor || '#ffffff',
        width: showDebugSheet ? `calc(100% - ${debugWidth}px)` : (settings?.show_sidebar ? "50%" : "100%")
      }}>
        {(currentStep === 'password' || currentStep === 'otp') && (
          <Button
            variant="ghost"
            onClick={() => setCurrentStep('login')}
            className="absolute left-8 top-8 p-2 hover:bg-gray-100 cursor-pointer"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
        )}

        <div className="w-full max-w-md space-y-8">
          <div className={cn("text-center mb-8", settings?.show_sidebar ? "lg:hidden" : "block")}>
            {settings?.logoSvg !== undefined ? (
              settings.logoSvg ? (
                <div
                  className="w-8 h-8 mx-auto mb-3 flex items-center justify-center"
                  style={{ color: settings.accentColor }}
                  dangerouslySetInnerHTML={{ __html: settings.logoSvg }}
                />
              ) : null
            ) : null}
            {settings?.logoName !== undefined ? (
              settings.logoName ? (
                <h1 className="text-xl font-semibold text-foreground">{settings.logoName}</h1>
              ) : null
            ) : (
              <h1 className="text-xl font-semibold text-foreground">GoAM</h1>
            )}
          </div>

          <div className="space-y-6">
            <div className="space-y-4">
              {(() => {
                const StepComponent = StepRegistry[currentStep] || StepRegistry['not-implemented'];
                return (
                  <StepComponent
                    isLoading={isLoading}
                    onContinue={handleContinue}
                    onRestart={handleRestart}
                    accentColor={settings?.accentColor}
                    formData={formData}
                    settings={settings}
                    error={error}
                    currentStep={currentStep}
                    currentNode={currentNode || undefined}
                    currentNodeType={currentNodeType || undefined}
                    prompts={prompts}
                    result={result}
                    executionId={executionId}
                  />
                );
              })()}
            </div>
          </div>
        </div>
      </div>

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

