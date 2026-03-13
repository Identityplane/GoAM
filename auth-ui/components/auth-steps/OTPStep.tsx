'use client';

import { InputOTP, InputOTPGroup, InputOTPSlot } from '@/components/ui/input-otp';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { AuthStepProps } from './types';
import { useState } from 'react';

export function OTPStep({ isLoading, onContinue, formData, error }: AuthStepProps) {
  const [otp, setOtp] = useState(formData.otp || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (otp.length === 6) {
      onContinue({ otp });
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Verify Your Email</h2>
        <p className="text-muted-foreground">Enter the 6-digit code sent to your email.</p>
      </div>

      {error && (
        <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md animate-in fade-in slide-in-from-top-1 text-center">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <FieldGroup>
          <Field>
            <FieldLabel>Verification Code</FieldLabel>
            <p className="text-sm text-gray-600 mb-4">
              Enter the 6-digit code sent to your email
            </p>
            <InputOTP
              maxLength={6}
              value={otp}
              onChange={(val) => setOtp(val)}
              disabled={isLoading}
            >
              <InputOTPGroup className="flex justify-center gap-2 w-full">
                {[...Array(6)].map((_, index) => (
                  <InputOTPSlot key={index} index={index} className="w-12 h-12 text-center" />
                ))}
              </InputOTPGroup>
            </InputOTP>
          </Field>
        </FieldGroup>

        <Button 
          type="submit" 
          disabled={isLoading || otp.length !== 6} 
          className="w-full"
        >
          {isLoading ? 'Verifying...' : 'Verify'}
        </Button>
      </form>
    </div>
  );
}

