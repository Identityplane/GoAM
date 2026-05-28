'use client';

import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { FieldGroup, Field, FieldLabel } from '@/components/ui/field';
import { Eye, EyeOff, Key } from 'lucide-react';
import { AuthStepProps } from './types';
import { base64urlToBuffer, serializeCredential } from '@/lib/passkeys';

/**
 * Migration of passwordOrSocialLogin.html
 */
export function PasswordOrSocialLoginStep({ 
  isLoading, 
  onContinue, 
  formData, 
  prompts, 
  settings, 
  accentColor,
  error 
}: AuthStepProps) {
  const [email, setEmail] = useState(formData.email || '');
  const [username, setUsername] = useState(formData.username || '');
  const [password, setPassword] = useState(formData.password || '');
  const [showPassword, setShowPassword] = useState(false);

  // Configuration from prompts (which flattens CustomConfig in the backend)
  // Configuration from prompts
  const useEmail = !!prompts?.email;
  const useUsername = !!prompts?.username;
  const usePassword = !!prompts?.password;
  const usePasskeys = !!prompts?.passkeysLoginOptions;
  const showForgotPassword = !!prompts?.forgotPassword;
  const showRegistrationLink = !!prompts?.register;
  
  // CustomConfig options might be passed as prompts too if the backend is configured to do so
  const disableSubmit = prompts?.disableSubmit === 'true';
  const disableDivider = prompts?.disableDivider === 'true';
  const submitBtnText = prompts?.['submit-btn-text'] || 'Sign In';
  const passkeyBtnText = prompts?.['passkey-btn-text'] || 'Sign in with passkey';
  const registerBtnText = prompts?.['register-btn-text'] || 'Register';
  const orContinueWithText = prompts?.['or-continue-with-text'] || 'or continue with';

  const hasInputs = useEmail || useUsername || usePassword;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onContinue({ 
      option: 'password',
      email, 
      username, 
      password 
    });
  };

  const handlePasskeyLogin = async () => {
    if (!prompts?.passkeysLoginOptions) return;

    try {
      const options = JSON.parse(prompts.passkeysLoginOptions);
      const pk = options.publicKey || options; // Handle both wrapped and unwrapped for safety

      // Convert base64url to buffers
      if (pk.challenge) pk.challenge = base64urlToBuffer(pk.challenge);
      if (pk.user?.id) pk.user.id = base64urlToBuffer(pk.user.id);
      if (pk.allowCredentials) {
        pk.allowCredentials = pk.allowCredentials.map((c: any) => ({
          ...c,
          id: base64urlToBuffer(c.id)
        }));
      }

      const credential = await navigator.credentials.get({ publicKey: pk });
      const serialized = serializeCredential(credential);
      
      onContinue({ 
        option: 'passkey',
        passkeysFinishLoginJson: JSON.stringify(serialized)
      });
    } catch (err) {
      console.error('Passkey error:', err);
      // We don't want to show passkey errors as main errors unless they are critical
    }
  };

  const handleSocialLogin = (option: string) => {
    onContinue({ option });
  };

  const handleForgotPassword = (e: React.MouseEvent) => {
    e.preventDefault();
    onContinue({ option: 'forgot-password' });
  };

  const handleRegister = () => {
    onContinue({ option: 'register' });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-3xl font-bold tracking-tight text-foreground">Sign In</h2>
        <p className="text-muted-foreground">
          {hasInputs ? 'Welcome back! Please enter your details.' : 'Sign in with your account to continue.'}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        {hasInputs && (
          <FieldGroup>
            {useEmail && (
              <Field>
                <FieldLabel>Email Address</FieldLabel>
                <Input
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={isLoading}
                  autoComplete="username webauthn"
                />
              </Field>
            )}

            {useUsername && (
              <Field>
                <FieldLabel>Username</FieldLabel>
                <Input
                  type="text"
                  placeholder="Username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                  disabled={isLoading}
                  autoComplete="username webauthn"
                />
              </Field>
            )}

            {usePassword && (
              <Field>
                <div className="flex items-center justify-between">
                  <FieldLabel>Password</FieldLabel>
                  {showForgotPassword && (
                    <button
                      onClick={handleForgotPassword}
                      className="text-xs font-medium hover:underline cursor-pointer"
                      style={{ color: accentColor || 'var(--primary)' }}
                    >
                      Forgot password?
                    </button>
                  )}
                </div>
                <div className="relative">
                  <Input
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    disabled={isLoading}
                    autoComplete="current-password webauthn"
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    disabled={isLoading}
                  >
                    {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
                {error && <p className="mt-1 text-xs text-red-500 font-medium">{error}</p>}
              </Field>
            )}
          </FieldGroup>
        )}

        {hasInputs && !disableSubmit && (
          <Button type="submit" disabled={isLoading} className="w-full font-semibold">
            {isLoading ? 'Signing in...' : submitBtnText}
          </Button>
        )}
      </form>

      {usePasskeys && (
        <Button
          variant="outline"
          onClick={handlePasskeyLogin}
          disabled={isLoading}
          className="w-full flex items-center justify-center gap-2 border-slate-200 hover:bg-slate-50"
        >
          <Key size={18} />
          {passkeyBtnText}
        </Button>
      )}

      {showRegistrationLink && (
        <div className="text-center text-sm text-muted-foreground">
          Don't have an account?{' '}
          <button
            onClick={handleRegister}
            className="font-semibold hover:underline cursor-pointer"
            style={{ color: accentColor || 'var(--primary)' }}
          >
            {registerBtnText}
          </button>
        </div>
      )}

      {(prompts?.social1 || prompts?.social2 || prompts?.social3) && (
        <div className="space-y-4">
          {hasInputs && !disableDivider && (
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-slate-200" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  {orContinueWithText}
                </span>
              </div>
            </div>
          )}

          <div className="grid grid-cols-1 gap-3">
             {prompts?.social1 && (
               <SocialButton 
                  provider={prompts.social1} 
                  onClick={() => handleSocialLogin('social1')} 
                  disabled={isLoading} 
                />
             )}
             {prompts?.social2 && (
               <SocialButton 
                  provider={prompts.social2} 
                  onClick={() => handleSocialLogin('social2')} 
                  disabled={isLoading} 
                />
             )}
             {prompts?.social3 && (
               <SocialButton 
                  provider={prompts.social3} 
                  onClick={() => handleSocialLogin('social3')} 
                  disabled={isLoading} 
                />
             )}
          </div>
        </div>
      )}
    </div>
  );
}

function SocialButton({ provider, onClick, disabled }: { provider: string, onClick: () => void, disabled?: boolean }) {
  // Map provider to icon/text
  const isGithub = provider.toLowerCase().includes('github');
  const isGoogle = provider.toLowerCase().includes('google');
  
  let label = `Sign in with ${provider}`;
  if (isGithub) label = "Sign in with GitHub";
  if (isGoogle) label = "Sign in with Google";

  return (
    <Button 
      variant="outline" 
      onClick={onClick} 
      disabled={disabled}
      className="w-full flex items-center justify-center gap-2 border-slate-200 hover:bg-slate-50"
    >
      {isGithub && (
        <svg height="18" width="18" viewBox="0 0 16 16" fill="currentColor">
          <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
        </svg>
      )}
      {isGoogle && (
        <svg height="18" width="18" viewBox="0 0 24 24">
          <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
          <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
          <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
          <path d="M12 5.38c1.62 0 3.06.56 4.21 1.66l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
        </svg>
      )}
      {label}
    </Button>
  );
}
