'use client';

import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { Eye, EyeOff } from 'lucide-react';
import { AuthStepProps } from './types';

export function AskUsernamePasswordStep({ isLoading, onContinue, formData, error }: AuthStepProps) {
  const [username, setUsername] = useState(formData.username || '');
  const [password, setPassword] = useState(formData.password || '');
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ username, password });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl text-foreground">Sign In</h2>
        <p className="text-muted-foreground">Please enter your username and password.</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <FieldGroup>
          <Field>
            <FieldLabel>Username</FieldLabel>
            <Input
              type="text"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={isLoading}
              autoFocus={!username}
            />
          </Field>
          <Field>
            <FieldLabel>Password</FieldLabel>
            <div className="relative">
              <Input
                type={showPassword ? 'text' : 'password'}
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={isLoading}
                className="pr-10"
                autoFocus={!!username}
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
            {error && <p className="text-red-500 text-xs mt-1">{error}</p>}
          </Field>
        </FieldGroup>

        <Button type="submit" disabled={isLoading || !username || !password} className="w-full">
          {isLoading ? 'Loading...' : 'Login'}
        </Button>
      </form>
    </div>
  );
}
