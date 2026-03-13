import { NextRequest, NextResponse } from 'next/server';

export type AuthStep = 'login' | 'register' | 'password' | 'otp' | 'terms' | 'success' | 'error';

export interface StepConfig {
  step: AuthStep;
  message: string;
  fields: string[];
  isLast: boolean;
}

interface AuthRequest {
  currentStep: AuthStep;
  data: Record<string, string>;
  action?: string;
}

// State machine definition
const stateTransitions: Record<AuthStep, AuthStep> = {
  login: 'password',
  register: 'password',
  password: 'otp',
  otp: 'terms',
  terms: 'success',
  success: 'login',
  error: 'login',
};

const stepConfigs: Record<AuthStep, Omit<StepConfig, 'step'>> = {
  login: {
    message: 'Enter your email to continue',
    fields: ['email'],
    isLast: false,
  },
  register: {
    message: 'Create a new account',
    fields: ['email'],
    isLast: false,
  },
  password: {
    message: 'Enter your password',
    fields: ['password'],
    isLast: false,
  },
  otp: {
    message: 'Enter the 6-digit code sent to your email',
    fields: ['otp'],
    isLast: false,
  },
  terms: {
    message: 'Accept our terms and conditions',
    fields: ['agreedTc', 'agreedPrivacy', 'agreedMarketing'],
    isLast: false,
  },
  success: {
    message: 'Authentication successful!',
    fields: [],
    isLast: true,
  },
  error: {
    message: 'An error occurred',
    fields: [],
    isLast: false,
  }
};


export async function POST(request: NextRequest) {
  try {
    const body: AuthRequest = await request.json();
    const { currentStep, data, action } = body;

    // Handle explicit actions to switch starting step or handle socials
    if (action === 'register') {
      const response: StepConfig = {
        step: 'register',
        ...stepConfigs['register'],
      };
      return NextResponse.json(response);
    }

    if (action === 'login') {
      const response: StepConfig = {
        step: 'login',
        ...stepConfigs['login'],
      };
      return NextResponse.json(response);
    }

    // Handle social and passkey actions
    if (action && ['google', 'apple', 'telegram', 'github', 'passkeys'].includes(action)) {
      const response: StepConfig = {
        step: 'success',
        ...stepConfigs['success'],
      };
      return NextResponse.json(response);
    }


     // Get next step based on current step
    const nextStep = stateTransitions[currentStep];

    if (!nextStep) {
      return NextResponse.json({
        step: 'error',
        ...stepConfigs['error'],
        message: 'Invalid or unknown step: ' + currentStep
      });
    }

    // Get step configuration
    const config = stepConfigs[nextStep];

    if (!config) {
      return NextResponse.json(
        { error: 'Invalid step' },
        { status: 400 }
      );
    }

    // Mock validation
    if ((currentStep === 'login' || currentStep === 'register') && !data.email) {
      return NextResponse.json(
        { error: 'Email is required' },
        { status: 400 }
      );
    }

    if (currentStep === 'password' && !data.password) {
      return NextResponse.json(
        { error: 'Password is required' },
        { status: 400 }
      );
    }

    if (currentStep === 'otp' && !data.otp) {
      return NextResponse.json(
        { error: 'OTP is required' },
        { status: 400 }
      );
    }

    // Return next step configuration
    const response: StepConfig = {
      step: nextStep,
      ...config,
    };

    return NextResponse.json(response);
  } catch (error) {
    console.error('Auth step error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
