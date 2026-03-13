'use client';

import { useState, useCallback, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { ArrowLeft } from 'lucide-react';
import { StepRegistry } from '@/components/auth-steps';
import { AuthAPI } from '@/lib/auth-api';
import type { AuthStep, StepConfig, FlowInfo, MetadataResponse } from '@/lib/auth-api';
import { cn } from '@/lib/utils';

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

export default function LoginPage() {
  const [currentStep, setCurrentStep] = useState<AuthStep>('login');
  const [executionId, setExecutionId] = useState<string | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [prompts, setPrompts] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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

  // Debug panel state
  const [configName, setConfigName] = useState('acme');
  const [showDebug, setShowDebug] = useState(false);

  const handleStartFlow = async (flowOverride?: string) => {
    const flowToStart = flowOverride || selectedFlow;
    if (!settings?.backend_url || !flowToStart) return;

    setError(null);
    setIsLoading(true);

    try {
      const flowResponse = await AuthAPI.startFlow(settings.backend_url, flowToStart);
      
      if (flowResponse.error) {
        setError(flowResponse.error.error_description);
        setCurrentStep('error');
        return;
      }

      if (flowResponse.currentNode) {
        setExecutionId(flowResponse.executionId || null);
        setSessionId(flowResponse.sessionId || null);
        setPrompts(flowResponse.prompts || {});
        setCurrentStep(flowResponse.currentNode as AuthStep);
      }
      
    } catch (err: any) {
      setError(err.message || 'Failed to start flow');
    } finally {
      setIsLoading(false);
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
      setError(null);
      setIsLoading(true);

      try {
        if (settings?.backend_url && selectedFlow && executionId && sessionId) {
          // Map stepData to responses (converting booleans to strings as expected by FlowRequest)
          const responses: Record<string, string> = {};
          for (const [key, value] of Object.entries(stepData)) {
            responses[key] = String(value);
          }

          const flowResponse = await AuthAPI.continueFlow(settings.backend_url, selectedFlow, {
            executionId,
            sessionId,
            currentNode: currentStep as string,
            responses,
          });

          if (flowResponse.error) {
            setError(flowResponse.error.error_description);
            // Don't change step if it's just a validation error, but current implementation transitions to error step
            setCurrentStep('error');
            return;
          }

          if (flowResponse.result?.success) {
            setCurrentStep('success');
          } else if (flowResponse.currentNode) {
            setExecutionId(flowResponse.executionId || null);
            setSessionId(flowResponse.sessionId || null);
            setPrompts(flowResponse.prompts || {});
            setCurrentStep(flowResponse.currentNode as AuthStep);

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
    
    if (settings?.backend_url && selectedFlow) {
      handleStartFlow(selectedFlow);
    } else {
      setCurrentStep('login');
    }
  };

  useEffect(() => {
    setSettings(null); // Reset settings to show loading state on config change
    setMetadata(null);
    setCurrentStep('login'); // Reset to default step
    setError(null);

    fetch(`/api/settings?config=${configName}`)
      .then((res) => res.json())
      .then(async (data) => {
        let finalSettings = { ...data };
        
        if (data.backend_url) {
          try {
            const meta = await AuthAPI.fetchMetadata(data.backend_url);
            setMetadata(meta);
            
            // Merge realm settings if present
            if (meta.realm?.settings) {
              const parsedRealmSettings = parseSettings(meta.realm.settings);
              finalSettings = { ...finalSettings, ...parsedRealmSettings };
            }

            if (meta.flows?.length > 0) {
              const firstFlow = meta.flows[0].route;
              setSelectedFlow(firstFlow);
              
              // Automatically start the first flow
              setIsLoading(true);
              try {
                const flowResponse = await AuthAPI.startFlow(data.backend_url, firstFlow);
                if (flowResponse.error) {
                  setError(flowResponse.error.error_description);
                  setCurrentStep('error');
                } else if (flowResponse.currentNode) {
                  setExecutionId(flowResponse.executionId || null);
                  setSessionId(flowResponse.sessionId || null);
                  setCurrentStep(flowResponse.currentNode as AuthStep);
                }
              } catch (err: any) {
                console.error('Failed to start auto-flow:', err);
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
  }, [configName]);

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
        <style dangerouslySetInnerHTML={{ __html: `
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
      {/* Debug Panel Toggle */}
      <button 
        onClick={() => setShowDebug(!showDebug)}
        className="fixed bottom-4 right-4 z-50 bg-black/50 hover:bg-black/80 text-white rounded-full w-10 h-10 flex items-center justify-center cursor-pointer transition-colors backdrop-blur-sm"
        title="Toggle Debug Panel"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m8 2 1.88 1.88"/><path d="M14.12 3.88 16 2"/><path d="M9 7.13v-1a3.003 3.003 0 1 1 6 0v1"/><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6"/><path d="M12 20v-9"/><path d="M6.53 9C4.6 8.8 3 7.1 3 5"/><path d="M6 13H2"/><path d="M3 21c0-2.1 1.7-3.9 3.8-4"/><path d="M20.97 5c0 2.1-1.6 3.8-3.5 4"/><path d="M22 13h-4"/><path d="M17.2 17c2.1.1 3.8 1.9 3.8 4"/></svg>
      </button>

      {/* Debug Panel */}
      {showDebug && (
        <div className="fixed bottom-16 right-4 z-50 w-64 bg-white rounded-lg shadow-xl border border-gray-200 p-4 transition-all" style={{ fontFamily: 'sans-serif' }}>
          <div className="text-sm font-semibold mb-3 border-b pb-2 flex justify-between items-center">
            <span>Debug Panel</span>
            <button onClick={() => setShowDebug(false)} className="text-gray-400 hover:text-gray-700">✕</button>
          </div>
          
          <div className="space-y-4 text-sm">
            <div>
              <label className="block text-gray-500 mb-1 text-xs uppercase font-semibold">Settings Config</label>
              <select 
                className="w-full border rounded p-1.5 focus:ring-2 focus:ring-blue-500 outline-none"
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
              <div className="pt-2 border-t">
                <label className="block text-blue-600 mb-1 text-xs uppercase font-bold">Backend Flow</label>
                <div className="flex gap-2">
                  <select 
                    className="flex-1 border rounded p-1.5 focus:ring-2 focus:ring-blue-500 outline-none truncate"
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
                    className="bg-blue-600 hover:bg-blue-700 text-white px-3 rounded flex items-center justify-center disabled:opacity-50"
                    title="Start Flow"
                  >
                    🚀
                  </button>
                </div>
                {metadata.realm && (
                  <div className="mt-2 text-[10px] text-gray-400 italic">
                    Connected to: {metadata.realm.name}
                  </div>
                )}
              </div>
            )}

            <div>
              <label className="block text-gray-500 mb-1 text-xs uppercase font-semibold">Current Step</label>
              <select 
                className="w-full border rounded p-1.5 focus:ring-2 focus:ring-blue-500 outline-none"
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
          </div>
        </div>
      )}

      <div className={cn(
        "hidden relative overflow-hidden",
        settings?.show_sidebar ? "lg:flex lg:w-1/2" : "lg:hidden"
      )} style={{ backgroundColor: settings?.backgroundColor || '#374151' }}>
        <div className="relative z-10 flex flex-col justify-between w-full px-12 py-12">
          <div className="flex items-center">
            {settings?.logoSvg ? (
              <div
                className="w-8 h-8 mr-3 flex items-center justify-center"
                style={{ color: settings.accentColor }}
                dangerouslySetInnerHTML={{ __html: settings.logoSvg }}
              />
            ) : (
              <div className="w-8 h-8 bg-white rounded-lg flex items-center justify-center mr-3">
                <div className="w-4 h-4 rounded-sm" style={{ backgroundColor: settings?.accentColor || '#3F3FF3' }}></div>
              </div>
            )}
            <h1 className="text-xl font-semibold text-white">{settings?.logoName || 'Frello'}</h1>
          </div>

            <div className="flex-1 flex flex-col justify-center">
              <h2 className="text-4xl text-white mb-6 leading-tight">
                {settings?.sidebarTitle || 'Effortlessly manage your team and operations.'}
              </h2>
              <p className="text-white/90 text-lg leading-relaxed">
                {settings?.sidebarText || 'Log in to access your CRM dashboard and manage your team.'}
              </p>
            </div>

          <div className="flex justify-between items-center text-white/70 text-sm">
            <span>{settings?.copyrightText || 'Copyright © 2025 Frello Enterprises LTD.'}</span>
            {settings?.privacyPolicyUrl ? (
              <a href={settings.privacyPolicyUrl} className="hover:text-white/90 cursor-pointer">Privacy Policy</a>
            ) : (
              <span className="cursor-pointer hover:text-white/90">Privacy Policy</span>
            )}
          </div>
        </div>
      </div>

      <div className={cn(
        "flex items-center justify-center p-8 relative transition-colors duration-300",
        settings?.show_sidebar ? "w-full lg:w-1/2" : "w-full"
      )} style={{ backgroundColor: settings?.pageBackgroundColor || '#ffffff' }}>
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
            {settings?.logoSvg ? (
              <div
                className="w-8 h-8 mx-auto mb-3 flex items-center justify-center"
                style={{ color: settings.accentColor }}
                dangerouslySetInnerHTML={{ __html: settings.logoSvg }}
              />
            ) : (
              <div className="w-8 h-8 rounded-lg flex items-center justify-center mx-auto mb-3" style={{ backgroundColor: settings?.accentColor || '#3F3FF3' }}>
                <div className="w-4 h-4 bg-white rounded-sm"></div>
              </div>
            )}
            <h1 className="text-xl font-semibold text-foreground">{settings?.logoName || 'Frello'}</h1>
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
                    prompts={prompts}
                  />
                );
              })()}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

