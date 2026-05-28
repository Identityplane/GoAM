'use client';

import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { AuthStepProps } from './types';

export function AskUsernameStep({ isLoading, onContinue, accentColor, formData, settings }: AuthStepProps) {
  const [username, setUsername] = useState(formData.username || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ username });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Username</h2>
        <p className="text-muted-foreground">Please enter your username to continue.</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <FieldGroup>
          <Field>
            <FieldLabel>Username</FieldLabel>
            <Input
              type="text"
              placeholder="Enter your username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={isLoading}
              autoFocus
            />
          </Field>
        </FieldGroup>

        <Button type="submit" disabled={isLoading || !username} className="w-full">
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
