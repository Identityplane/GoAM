'use client';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { SocialAuthButtons } from '../components/SocialAuthButtons';
import { PasskeyButton } from '../components/PasskeyButton';
import { AuthStepProps } from './types';
import { useState } from 'react';

export function RegisterStep({ isLoading, onContinue, accentColor, formData, settings, error }: AuthStepProps) {
  const [email, setEmail] = useState(formData.email || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ email });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Create Account</h2>
        <p className="text-muted-foreground">Create a new account to get started.</p>
      </div>

      {error && (
        <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md animate-in fade-in slide-in-from-top-1 text-center">
          {error}
        </div>
      )}

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

        <Button type="submit" disabled={isLoading} className="w-full">
          {isLoading ? 'Loading...' : 'Register'}
        </Button>
      </form>

      <div className="space-y-4">
        <PasskeyButton 
          settings={settings} 
          onAction={(action) => onContinue({}, action)}
          isLoading={isLoading}
        />

        <SocialAuthButtons 
          settings={settings} 
          onAction={(provider) => onContinue({}, provider)}
          isLoading={isLoading}
        />
      </div>

      <div className="text-center text-sm text-muted-foreground">
        Already Have An Account?{' '}
        <Button
          variant="link"
          className="p-0 h-auto text-sm hover:text-opacity-80 font-medium cursor-pointer"
          style={{ color: accentColor || '#3F3FF3' }}
          onClick={() => onContinue({}, 'login')}
        >
          Sign In.
        </Button>
      </div>
    </div>
  );
}

