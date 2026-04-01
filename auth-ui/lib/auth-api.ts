export type AuthStep = 'login' | 'register' | 'password' | 'askUserID' | 'askEmail' | 'askPassword' | 'askUsername' | 'askUsernamePassword' | 'askEmailPassword' | 'emailOTP' | 'passwordOrSocialLogin' | 'verifyYubikeyOtp' | 'registerPasskey' | 'otp' | 'terms' | 'success' | 'error' | 'not-implemented' | 'successResult' | 'failureResult' | 'listAvailableUsers' | (string & {});

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
  currentNodeType?: string;
  prompts?: Record<string, string>;
  result?: {
    success: boolean;
    redirect?: string;
    user_id?: string;
    [key: string]: any;
  };
  error?: {
    error: string;
    error_description: string;
  };
  errorMessage?: string;
  flow?: string;
  debug?: any;
}

export interface FlowRequest {
  executionId: string;
  sessionId: string;
  currentNode: string;
  responses: Record<string, string>;
}

/**
 * Service for interacting with the Authentication API (both internal mock and external backend)
 */
export class AuthAPI {
  /**
   * Helper to build a URL from a potentially relative backend URL
   */
  private static buildUrl(backendUrl: string, path: string): URL {
    // Ensure backendUrl doesn't end with / and path starts with /
    const sanitizedBackend = backendUrl.endsWith('/') ? backendUrl.slice(0, -1) : backendUrl;
    const sanitizedPath = path.startsWith('/') ? path : `/${path}`;
    const fullPath = `${sanitizedBackend}${sanitizedPath}`;
    
    if (typeof window !== 'undefined') {
      // On client-side, we can use window.location.origin as a fallback for relative URLs
      return new URL(fullPath, window.location.origin);
    }
    // On server-side, URL must be absolute or it will throw
    return new URL(fullPath);
  }

  /**
   * Fetches realm metadata from the backend
   */
  static async fetchMetadata(backendUrl: string, debug?: boolean): Promise<MetadataResponse> {
    const url = this.buildUrl(backendUrl, '/api/v1/');
    if (debug) url.searchParams.set('debug', 'true');

    const response = await fetch(url.toString(), {
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
  static async startFlow(backendUrl: string, flowRoute: string, debug?: boolean, isContinue?: boolean, isInit?: boolean): Promise<FlowResponse> {
    const url = this.buildUrl(backendUrl, `/api/v1/${flowRoute}`);
    if (debug) url.searchParams.set('debug', 'true');
    if (isContinue) url.searchParams.set('continue', 'true');
    if (isInit) url.searchParams.set('init', 'true');

    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: { 'Accept': 'application/json' },
    });

    if (!response.ok) {
      // Return the error response instead of throwing
      try {
        return await response.json();
      } catch (e) {
        return {
          error: {
            error: 'FETCH_ERROR',
            error_description: `Server returned ${response.status}`,
          }
        };
      }
    }

    return response.json();
  }

  /**
   * Continues the authentication flow with user input
   */
  static async continueFlow(backendUrl: string, flowRoute: string, request: FlowRequest, debug?: boolean): Promise<FlowResponse> {
    const url = this.buildUrl(backendUrl, `/api/v1/${flowRoute}`);
    if (debug) url.searchParams.set('debug', 'true');

    const response = await fetch(url.toString(), {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Accept': 'application/json' 
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      try {
        return await response.json();
      } catch (e) {
        return {
          error: {
            error: 'FETCH_ERROR',
            error_description: `Server returned ${response.status}`,
          }
        };
      }
    }

    return response.json();
  }

  /**
   * Resumes an existing authentication session using its ID
   */
  static async resumeSession(backendUrl: string, sessionId: string, debug?: boolean): Promise<FlowResponse> {
    const url = this.buildUrl(backendUrl, '/api/v1/');
    if (debug) url.searchParams.set('debug', 'true');

    const response = await fetch(url.toString(), {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Accept': 'application/json' 
      },
      body: JSON.stringify({ sessionId }),
    });

    if (!response.ok) {
      try {
        return await response.json();
      } catch (e) {
        return {
          error: {
            error: 'FETCH_ERROR',
            error_description: `Server returned ${response.status}`,
          }
        };
      }
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
