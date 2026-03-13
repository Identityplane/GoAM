'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp';
import { AuthStepProps } from './types';

export function EmailOTPStep({ isLoading, onContinue, formData, prompts, error }: AuthStepProps) {
  const [otp, setOtp] = useState(formData.otp || '');
  const [resendTimer, setResendTimer] = useState<number>(0);
  
  // Initialize timer from prompts
  useEffect(() => {
    const resendIn = parseInt(prompts?.resend_in_seconds || '0', 10);
    if (resendIn > 0) {
      setResendTimer(resendIn);
    }
  }, [prompts?.resend_in_seconds]);

  // Countdown logic
  useEffect(() => {
    if (resendTimer > 0) {
      const interval = setInterval(() => {
        setResendTimer((prev) => prev - 1);
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [resendTimer]);

  const handleSubmit = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (otp.length === 6) {
      onContinue({ otp });
    }
  };

  const handleResend = () => {
    onContinue({ option: 'resend' });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Verify Email</h2>
        <p className="text-muted-foreground">
          Enter the code sent to <b>{prompts?.email}</b>
        </p>
      </div>

      <div className="flex flex-col items-center space-y-6">
        <form onSubmit={handleSubmit} className="w-full flex flex-col items-center space-y-6">
          <InputOTP
            maxLength={6}
            value={otp}
            onChange={(value) => setOtp(value)}
            disabled={isLoading}
            onComplete={() => onContinue({ otp: otp })} // Auto submit on complete would be nice but otp is updated after this if we are not careful
          >
            <InputOTPGroup>
              <InputOTPSlot index={0} />
              <InputOTPSlot index={1} />
              <InputOTPSlot index={2} />
              <InputOTPSlot index={3} />
              <InputOTPSlot index={4} />
              <InputOTPSlot index={5} />
            </InputOTPGroup>
          </InputOTP>

          {error && <p className="text-red-500 text-xs">{error}</p>}

          <Button 
            type="submit" 
            disabled={isLoading || otp.length !== 6} 
            className="w-full"
          >
            {isLoading ? 'Verifying...' : 'Login'}
          </Button>
        </form>

        <div className="text-center w-full">
          <Button
            variant="ghost"
            onClick={handleResend}
            disabled={isLoading || resendTimer > 0}
            className="text-muted-foreground hover:text-foreground text-sm"
          >
            {resendTimer > 0 
              ? `Resend in ${resendTimer}s` 
              : "Didn't receive a code? Resend"}
          </Button>
        </div>
      </div>
    </div>
  );
}
