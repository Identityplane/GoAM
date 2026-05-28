'use client';

import { useState, useRef, useEffect } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { AuthStepProps } from './types';

export function VerifyYubikeyOTPStep({ 
  isLoading, 
  onContinue, 
  prompts, 
  accentColor,
  error 
}: AuthStepProps) {
  const [otp, setOtp] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const label = prompts?.label || 'Yubikey OTP Code';
  const buttonText = prompts?.button_text || 'Validate';

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (otp) {
      onContinue({ yubikeyOtpVerification: otp });
    }
  };

  // Auto-submit when OTP is pasted/entered by Yubikey
  useEffect(() => {
    if (otp.length >= 44) { // Standard Yubikey OTP length is 44 characters
      handleSubmit();
    }
  }, [otp]);

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  return (
    <div className="space-y-8">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl font-bold tracking-tight text-foreground">Yubikey Verification</h2>
        <p className="text-muted-foreground">Please tap your Yubikey to continue.</p>
      </div>

      <div className="flex flex-col items-center justify-center py-4">
        {/* Yubikey Graphic */}
        <div className="relative w-32 h-16 mb-4">
          {/* Key body */}
          <div className="absolute inset-0 bg-gray-400 rounded-lg shadow-md flex items-center justify-end pr-3">
             {/* The contact/button area */}
             <div className="relative w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center overflow-hidden">
                {/* Pulsing effect */}
                <div className="absolute inset-0 bg-white/30 animate-ping rounded-full" />
                <div className="w-5 h-5 rounded-full border border-gray-400 bg-gray-200 flex items-center justify-center">
                   <div className="w-1.5 h-1.5 rounded-full bg-yellow-500/80 shadow-[0_0_6px_rgba(234,179,8,0.6)]" />
                </div>
             </div>
             {/* Key length markings */}
             <div className="absolute left-4 top-1/2 -translate-y-1/2 flex gap-0.5">
                <div className="w-0.5 h-2.5 bg-gray-300 rounded-full" />
                <div className="w-0.5 h-2 bg-gray-300 rounded-full" />
                <div className="w-0.5 h-2.5 bg-gray-300 rounded-full" />
             </div>
          </div>
          {/* USB connector */}
          <div className="absolute -left-2 top-1/2 -translate-y-1/2 w-2 h-8 bg-gray-300 rounded-l-sm" />
          
          {/* Action indicator */}
          <div className="absolute -top-3 right-2 animate-bounce">
             <div className="bg-gray-600 text-white text-[8px] px-1.5 py-0.5 rounded shadow-sm border border-gray-500">
                TAP HERE
             </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="w-full space-y-4">
          <FieldGroup>
            <Field>
              <FieldLabel>{label}</FieldLabel>
              <Input
                ref={inputRef}
                type="text"
                placeholder="Insert Yubikey and tap..."
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                required
                disabled={isLoading}
                className="text-center font-mono tracking-widest bg-slate-50 border-slate-200 focus:ring-slate-500"
                autoComplete="off"
              />
              {error && <p className="mt-1 text-xs text-red-500 font-medium">{error}</p>}
            </Field>
          </FieldGroup>

          <Button 
            type="submit" 
            disabled={isLoading || !otp} 
            className="w-full font-semibold h-11"
            style={{ backgroundColor: accentColor }}
          >
            {isLoading ? 'Verifying...' : buttonText}
          </Button>
        </form>
      </div>

      <div className="text-center">
        <p className="text-sm text-muted-foreground italic">
          If your Yubikey is connected, simply touch the gold contact to verify.
        </p>
      </div>
    </div>
  );
}
