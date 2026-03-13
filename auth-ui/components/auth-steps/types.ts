import { AuthStep } from '@/lib/auth-api';

export interface AuthStepProps {
  isLoading: boolean;
  onContinue: (data: Record<string, string | boolean>, action?: string) => void;
  accentColor?: string;
  onBack?: () => void;
  onRestart?: () => void;
  formData: Record<string, any>;
  settings?: any;
  error?: string | null;
  currentStep?: string;
  prompts?: Record<string, string>;
}


