'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Fingerprint, Loader2, ShieldCheck, AlertCircle } from 'lucide-react';
import { AuthStepProps } from './types';
import { cn } from '@/lib/utils';

// Helper functions adapted from legacy passkeys_helper.ts
function base64urlToBuffer(base64url: string): Uint8Array {
  const padding = '='.repeat((4 - base64url.length % 4) % 4);
  const base64 = (base64url + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(base64);
  return new Uint8Array([...raw].map((char) => char.charCodeAt(0)));
}

function bufferToBase64url(buffer: ArrayBuffer): string {
  const binary = String.fromCharCode(...new Uint8Array(buffer));
  const base64 = btoa(binary);
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function serializeCredential(cred: PublicKeyCredential) {
  if (!cred) return null;

  const response = cred.response as AuthenticatorAttestationResponse;
  const json: any = {
    id: cred.id,
    type: cred.type,
    rawId: bufferToBase64url(cred.rawId),
    authenticatorAttachment: cred.authenticatorAttachment || null,
    clientExtensionResults: cred.getClientExtensionResults?.() || {},
    response: {
      clientDataJSON: bufferToBase64url(response.clientDataJSON),
      attestationObject: bufferToBase64url(response.attestationObject),
      publicKey: response.getPublicKey ? bufferToBase64url(response.getPublicKey()!) : undefined,
      publicKeyAlgorithm: response.getPublicKeyAlgorithm ? response.getPublicKeyAlgorithm() : undefined,
      transports: response.getTransports ? response.getTransports() : undefined,
    },
  };

  return json;
}

export function RegisterPasskeyStep({ isLoading: apiLoading, onContinue, accentColor, prompts, error: apiError }: AuthStepProps) {
  const [status, setStatus] = useState<'idle' | 'prompting' | 'processing' | 'success' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    if (apiError) setStatus('error');
  }, [apiError]);

  const handleRegister = async () => {
    setErrorMessage(null);
    setStatus('prompting');

    try {
      const optionsJson = prompts?.passkeysOptions;
      if (!optionsJson) {
        throw new Error('No passkey options provided by backend');
      }

      const options = JSON.parse(optionsJson);

      // Transform options for navigator.credentials.create
      const publicKey = options.publicKey;
      publicKey.challenge = base64urlToBuffer(publicKey.challenge);
      publicKey.user.id = base64urlToBuffer(publicKey.user.id);

      if (publicKey.excludeCredentials) {
        publicKey.excludeCredentials = publicKey.excludeCredentials.map((cred: any) => ({
          ...cred,
          id: base64urlToBuffer(cred.id),
        }));
      }

      const credential = await navigator.credentials.create({ publicKey }) as PublicKeyCredential;
      
      if (!credential) {
        throw new Error('Failed to create credential');
      }

      setStatus('processing');
      const serialized = serializeCredential(credential);
      
      onContinue({
        passkeysFinishRegistrationJson: JSON.stringify(serialized)
      });
      
      setStatus('success');
    } catch (err: any) {
      console.error('Passkey registration error:', err);
      // Don't set error if user cancelled
      if (err.name === 'NotAllowedError' || err.name === 'AbortError') {
        setStatus('idle');
      } else {
        setStatus('error');
        setErrorMessage(err.message || 'Failed to register passkey');
      }
    }
  };

  return (
    <div className="space-y-8">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl font-semibold tracking-tight text-foreground">Secure your account</h2>
        <p className="text-muted-foreground">
          Register a passkey to enable faster, more secure sign-ins with biometric or hardware security.
        </p>
      </div>

      <div className="flex flex-col items-center justify-center py-4">
        <div className="relative">
          <div className={cn(
            "w-24 h-24 rounded-full flex items-center justify-center transition-all duration-500",
            status === 'idle' && "bg-primary/5",
            status === 'prompting' && "bg-primary/20 animate-pulse",
            status === 'processing' && "bg-primary/10",
            status === 'success' && "bg-green-100",
            status === 'error' && "bg-red-100"
          )} style={{ backgroundColor: status === 'idle' ? `${accentColor}10` : undefined }}>
            {status === 'success' ? (
              <ShieldCheck className="w-12 h-12 text-green-600 animate-in zoom-in-50 duration-300" />
            ) : status === 'error' ? (
              <AlertCircle className="w-12 h-12 text-red-600 animate-in zoom-in-50 duration-300" />
            ) : (
              <Fingerprint 
                className={cn(
                  "w-12 h-12 transition-colors duration-300",
                  status === 'idle' ? "text-primary" : "text-primary/70"
                )} 
                style={status === 'idle' ? { color: accentColor } : {}}
              />
            )}
          </div>
          
          {(status === 'prompting' || status === 'processing' || apiLoading) && (
            <div className="absolute -inset-2 border-2 border-primary/30 border-t-primary rounded-full animate-spin" style={{ borderTopColor: accentColor }} />
          )}
        </div>
      </div>

      <div className="space-y-4">
        <Button
          onClick={handleRegister}
          disabled={status === 'prompting' || status === 'processing' || status === 'success' || apiLoading}
          className="w-full h-12 text-lg font-medium shadow-sm transition-all active:scale-[0.98]"
          style={{ backgroundColor: accentColor }}
        >
          {status === 'idle' && 'Create Passkey'}
          {status === 'prompting' && 'Follow browser prompts...'}
          {status === 'processing' && (
            <span className="flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              Finalizing...
            </span>
          )}
          {status === 'success' && 'Registered!'}
          {status === 'error' && 'Try Again'}
        </Button>

        {status === 'error' && (errorMessage || apiError) && (
          <p className="text-sm text-center text-red-500 font-medium animate-in fade-in slide-in-from-top-1">
            {errorMessage || apiError}
          </p>
        )}

        <div className="flex flex-col gap-2">
          <Button
            variant="ghost"
            onClick={() => onContinue({}, 'skip')}
            disabled={status === 'prompting' || status === 'processing' || apiLoading}
            className="w-full text-muted-foreground hover:text-foreground"
          >
            Skip for now
          </Button>
        </div>
      </div>

      <div className="pt-4 border-t border-muted text-[10px] text-center text-muted-foreground uppercase tracking-widest font-semibold">
        Protected by WebAuthn
      </div>
    </div>
  );
}
