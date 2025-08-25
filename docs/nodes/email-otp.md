# Email OTP Node

The Email OTP node provides one-time password (OTP) verification via email for multi-factor authentication. This node can be used both for existing user authentication and for new user onboarding scenarios.

## Overview

The Email OTP node is a versatile authentication component that:
- Generates cryptographically secure 6-digit OTPs
- Sends OTPs via email (when SMTP is configured)
- Tracks failed attempts and implements account lockout
- Works with or without existing user accounts
- Integrates with the user attributes system for security tracking

## Node Type
`NodeTypeQueryWithLogic`

## Required Context
- `email`: The email address to send the OTP to

## Output Context
- `email_otp`: The generated OTP stored in session context
- `error`: Error message if OTP validation fails

## Custom Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `smtp_server` | string | Yes* | SMTP server address for sending emails |
| `smtp_port` | string | Yes* | SMTP server port number |
| `smtp_username` | string | Yes* | Username for SMTP authentication |
| `smtp_password` | string | Yes* | Password for SMTP authentication |
| `smtp_sender_email` | string | Yes* | Email address that appears as sender |
| `mfa_max_attempts` | string | No | Maximum failed OTP attempts before lockout (default: 10) |

*Required for actual email delivery. If not provided, OTPs are logged but not sent.

## Behavior

### 1. Initial OTP Generation (No OTP Input)
When the node is called without an OTP input:

1. **User Lookup**: Attempts to load user from context or by email attribute
2. **Email Validation**: Ensures email is provided in context or user attributes
3. **Account Lock Check**: If user exists, checks if email is locked due to too many failed attempts
4. **OTP Generation**: Generates a cryptographically secure 6-digit OTP
5. **Email Sending**: Sends OTP via email (if SMTP configured and account not locked)
6. **Context Storage**: Stores OTP in session context for verification
7. **Prompt Return**: Returns OTP input prompt to user

### 2. OTP Verification (With OTP Input)
When the node is called with an OTP input:

1. **OTP Validation**: Compares input OTP with stored OTP from context
2. **Failed Attempt Handling**: If OTP is incorrect and user exists:
   - Increments failed attempt counter
   - Locks account if max attempts exceeded
   - Updates email attribute in database
3. **Success Handling**: If OTP is correct and user exists:
   - Resets failed attempt counter to 0
   - Unlocks account
   - Sets email as verified
   - Updates verification timestamp
   - Updates email attribute in database
4. **Result Return**: Returns success or prompts for retry

## Email Attribute Integration

The node works with the `EmailAttributeValue` which includes:

- **Email** (`string`): The email address
- **Verified** (`bool`): Whether the email has been verified
- **VerifiedAt** (`*time.Time`): Timestamp when email was verified
- **OtpFailedAttempts** (`int`): Number of failed OTP attempts
- **OtpLocked** (`bool`): Whether the email is locked due to too many failed attempts

## Security Features

### Account Lockout
- **Configurable Threshold**: Default 10 failed attempts, configurable via `mfa_max_attempts`
- **Automatic Locking**: Account is locked when threshold is exceeded
- **Silent Failure**: When locked, OTP generation continues but emails are not sent
- **Automatic Unlocking**: Account unlocks on successful OTP verification

### OTP Security
- **Cryptographic Generation**: Uses `crypto/rand` for secure random number generation
- **6-Digit Format**: Consistent 6-digit format (000000-999999)
- **Session Storage**: OTP stored in session context, not persisted to database
- **One-Time Use**: Each OTP is single-use and expires with session

## Usage Scenarios

### 1. Existing User Authentication
```yaml
# Flow configuration for existing user MFA
flow:
  - askEmail
  - emailOTP
  - success
```

**Behavior**:
- Loads existing user by email
- Tracks failed attempts and implements lockout
- Verifies email on successful OTP
- Updates user attributes with verification status

### 2. New User Onboarding
```yaml
# Flow configuration for new user registration
flow:
  - askEmail
  - emailOTP
  - createUser
  - success
```

**Behavior**:
- Works without existing user account
- No attempt limiting (unlimited OTP generation)
- Suitable for registration and onboarding flows
- Should be protected by CAPTCHA or similar rate limiting

### 3. Account Recovery
```yaml
# Flow configuration for password reset
flow:
  - askEmail
  - emailOTP
  - resetPassword
  - success
```

**Behavior**:
- Verifies email ownership before allowing password reset
- Implements security measures to prevent abuse
- Tracks attempts for security monitoring

## Configuration Examples

### Basic Configuration
```yaml
emailOTP:
  type: emailOTP
  customConfig:
    mfa_max_attempts: "5"
```

### Full SMTP Configuration
```yaml
emailOTP:
  type: emailOTP
  customConfig:
    smtp_server: "smtp.gmail.com"
    smtp_port: "587"
    smtp_username: "your-app@gmail.com"
    smtp_password: "your-app-password"
    smtp_sender_email: "noreply@yourapp.com"
    mfa_max_attempts: "3"
```

## Result States

| State | Description | When Returned |
|-------|-------------|---------------|
| `success` | OTP verification successful | Valid OTP provided |
| `failure` | OTP verification failed | Invalid OTP provided |
| `locked` | Account locked due to too many attempts | Max failed attempts exceeded |

## Prompts

| Prompt | Type | Description |
|--------|------|-------------|
| `otp` | `number` | 6-digit OTP input field |

## Error Handling

### Common Error Scenarios
1. **Missing Email**: Returns error if no email provided in context
2. **SMTP Configuration**: Logs warning if SMTP not configured (OTP still generated)
3. **User Lookup Failure**: Handles gracefully for onboarding scenarios
4. **Attribute Loading Failure**: Returns error if email attribute cannot be loaded

### Silent Failures
- **Locked Account**: Fails silently when account is locked (security best practice)
- **SMTP Issues**: Continues OTP generation even if email delivery fails

## Integration Notes

### User Repository Requirements
- Must implement `UpdateUserAttribute` method
- Should support email attribute lookups via `GetByAttributeIndex`

### Email Sender Service
- Must implement `SendEmail` method
- Should handle SMTP authentication and delivery
- Graceful fallback when SMTP is not configured

### Security Considerations
1. **Rate Limiting**: Implement CAPTCHA or rate limiting for onboarding flows
2. **Session Management**: Ensure OTP context is properly scoped to user session
3. **Audit Logging**: Log OTP generation and verification attempts
4. **Account Recovery**: Provide alternative recovery methods for locked accounts

## Best Practices

1. **SMTP Configuration**: Always configure SMTP for production use
2. **Attempt Limiting**: Use reasonable attempt limits (3-5 for security, 10 for usability)
3. **User Experience**: Provide clear feedback on failed attempts and lockout status
4. **Monitoring**: Track OTP success/failure rates and lockout events
5. **Fallback**: Implement alternative MFA methods for critical accounts

## Example Email Template

The node generates emails with this default template:
```
Subject: Verify your identity with OTP

Please use the verification code below to confirm your identity.

Verification code:

123456
```