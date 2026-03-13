'use client';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { Eye, EyeOff } from 'lucide-react';
import { AuthStepProps } from './types';
import { useState } from 'react';

export function AskEmailPasswordStep({ isLoading, onContinue, accentColor, formData, settings, error }: AuthStepProps) {
  const [email, setEmail] = useState(formData.email || '');
  const [password, setPassword] = useState(formData.password || '');
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ email, password });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Sign In</h2>
        <p className="text-muted-foreground">Please enter your email and password to continue.</p>
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
          
          <Field>
            <FieldLabel>Password</FieldLabel>
            <div className="relative">
              <Input
                type={showPassword ? 'text' : 'password'}
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={isLoading}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
                disabled={isLoading}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </Field>
        </FieldGroup>

        <Button type="submit" disabled={isLoading || !email || !password} className="w-full">
          {isLoading ? 'Loading...' : 'Login'}
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
