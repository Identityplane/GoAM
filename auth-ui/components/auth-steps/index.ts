import { LoginStep } from './LoginStep';
import { RegisterStep } from './RegisterStep';
import { PasswordStep } from './PasswordStep';
import { OTPStep } from './OTPStep';
import { TermsStep } from './TermsStep';
import { SuccessStep } from './SuccessStep';
import { ErrorStep } from './ErrorStep';
import { NotImplementedStep } from './NotImplementedStep';
import { AskEmailPasswordStep } from './AskEmailPasswordStep';
import { VerifyYubikeyOTPStep } from './VerifyYubikeyOTPStep';
import { AskEmailStep } from './AskEmailStep';
import { AskPasswordStep } from './AskPasswordStep';
import { AskUsernameStep } from './AskUsernameStep';
import { AskUsernamePasswordStep } from './AskUsernamePasswordStep';
import { LoginPasskeyStep } from './LoginPasskeyStep';
import { RegisterPasskeyStep } from './RegisterPasskeyStep';
import { EmailOTPStep } from './EmailOTPStep';
import { PasswordOrSocialLoginStep } from './PasswordOrSocialLoginStep';
import { AskUserIDStep } from './AskUserIDStep';
import { AuthStep } from '@/lib/auth-api';
import { AuthStepProps } from './types';

export const StepRegistry: Record<AuthStep, React.ComponentType<AuthStepProps>> = {
  'login': LoginStep,
  'register': RegisterStep,
  'password': PasswordStep,
  'otp': OTPStep,
  'terms': TermsStep,
  'success': SuccessStep,
  'error': ErrorStep,
  'not-implemented': NotImplementedStep,
  'askEmailPassword': AskEmailPasswordStep,
  'askEmail': AskEmailStep,
  'askPassword': AskPasswordStep,
  'askUsername': AskUsernameStep,
  'askUsernamePassword': AskUsernamePasswordStep,
  'emailOTP': EmailOTPStep,
  'passwordOrSocialLogin': PasswordOrSocialLoginStep,
  'verifyYubikeyOtp': VerifyYubikeyOTPStep,
  'askUserID': AskUserIDStep,
  'registerPasskey': RegisterPasskeyStep,
  'successResult': SuccessStep,
  'failureResult': ErrorStep,
  // Add these if they are relevant AuthSteps (they might need to be added to the AuthStep union)
  // 'login-passkey': LoginPasskeyStep,
  // 'register-passkey': RegisterPasskeyStep,
};
