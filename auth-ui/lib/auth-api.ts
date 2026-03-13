export type AuthStep = 'login' | 'register' | 'password' | 'otp' | 'terms' | 'success' | 'error' | 'not-implemented';

export interface StepConfig {
  step: AuthStep;
  message: string;
  fields: string[];
  isLast: boolean;
  error?: string;
}

export interface FlowInfo {
  id: string;
  route: string;
  description: string;
}

export interface MetadataResponse {
  flows: FlowInfo[];
  realm: {
    name: string;
    baseUrl: string;
    settings: Record<string, string>;
  };
}

export interface FlowResponse {
  executionId?: string;
  sessionId?: string;
  currentNode?: string;
  prompts?: Record<string, string>;
  result?: {
    success: boolean;
    [key: string]: any;
  };
  error?: {
    error: string;
    errorDescription: string;
  };
}

/**
 * Service for interacting with the Authentication API (both internal mock and external backend)
 */
export class AuthAPI {
  /**
   * Fetches realm metadata from the backend
   */
  static async fetchMetadata(backendUrl: string): Promise<MetadataResponse> {
    const response = await fetch(`${backendUrl}/api/v1/`, {
      headers: { 'Accept': 'application/json' },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch metadata');
    }

    return response.json();
  }

  /**
   * Starts a specific authentication flow
   */
  static async startFlow(backendUrl: string, flowRoute: string): Promise<FlowResponse> {
    const response = await fetch(`${backendUrl}/api/v1/${flowRoute}`, {
      method: 'GET',
      headers: { 'Accept': 'application/json' },
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error?.errorDescription || 'Failed to start flow');
    }

    return response.json();
  }

  /**
   * Continues the authentication flow with user input (Internal Mock API)
   */
  static async processInternalStep(
    currentStep: AuthStep,
    data: Record<string, any>,
    action?: string
  ): Promise<StepConfig> {
    const response = await fetch('/api/auth/step', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        currentStep,
        data,
        action,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'An error occurred');
    }

    return response.json();
  }
}
