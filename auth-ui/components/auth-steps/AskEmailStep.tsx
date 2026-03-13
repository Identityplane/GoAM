'use client';

import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { AuthStepProps } from './types';

export function AskEmailStep({ isLoading, onContinue, accentColor, formData, settings }: AuthStepProps) {
  const [email, setEmail] = useState(formData.email || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ email });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Email</h2>
        <p className="text-muted-foreground">Please enter your email address to continue.</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <FieldGroup>
          <Field>
            <FieldLabel>Email Address</FieldLabel>
            <Input
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={isLoading}
              autoFocus
            />
          </Field>
        </FieldGroup>

        <Button type="submit" disabled={isLoading || !email} className="w-full">
          {isLoading ? 'Loading...' : 'Continue'}
        </Button>
      </form>

      {settings?.enable_register !== false && (
        <div className="text-center text-sm text-muted-foreground">
          Don't Have An Account?{' '}
          <Button
            variant="link"
            className="p-0 h-auto text-sm hover:text-opacity-80 font-medium cursor-pointer"
            style={{ color: accentColor || '#3F3FF3' }}
            onClick={() => onContinue({}, 'register')}
          >
            Register Now.
          </Button>
        </div>
      )}
    </div>
  );
}
