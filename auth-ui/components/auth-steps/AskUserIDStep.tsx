'use client';

import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { AuthStepProps } from './types';

export function AskUserIDStep({ isLoading, onContinue, accentColor, formData, settings }: AuthStepProps) {
  const [userId, setUserId] = useState(formData.user_id || formData.email || formData.username || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ user_id: userId });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Sign In</h2>
        <p className="text-muted-foreground">Please enter your user ID to continue.</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <FieldGroup>
          <Field>
            <FieldLabel>User ID</FieldLabel>
            <Input
              type="text"
              placeholder="Username or Email"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              required
              disabled={isLoading}
              autoFocus
              autoComplete="username"
            />
          </Field>
        </FieldGroup>

        <Button type="submit" disabled={isLoading || !userId} className="w-full">
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
