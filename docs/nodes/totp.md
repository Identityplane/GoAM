# TOTP (Time-based One-Time Password) Nodes

The TOTP nodes provide RFC 6238 compliant Time-based One-Time Password functionality for two-factor authentication (2FA). These nodes allow users to create and verify TOTP codes using authenticator apps like Google Authenticator, Authy, or Microsoft Authenticator.

## Overview

The TOTP implementation consists of two main nodes:
- **createTOTP**: Creates a new TOTP secret and generates a QR code for user enrollment
- **verifyTOTP**: Verifies TOTP codes during authentication

## TOTP Attribute

Both nodes work with a user attribute of type `"identityplane:totp"` that contains:

- **SecretKey** (`string`): The secret key used to generate TOTP codes
- **Locked** (`bool`): Whether the TOTP is locked due to too many failed attempts
- **FailedAttempts** (`int`): Counter for failed verification attempts

## createTOTP Node

### Purpose
Creates a new TOTP secret for a user and generates a QR code for enrollment in authenticator apps.

### Node Type
`NodeTypeQueryWithLogic`

### Required Context
- User must be present in the authentication session

### Custom Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `totpIssuer` | string | "GoAM" | The name of the issuer displayed in the TOTP QR code |
| `saveUser` | string | "false" | If "true", the user will be saved to the database after TOTP creation |

### Behavior

1. **Initial Call** (no input):
   - Generates a new TOTP secret
   - Creates a QR code image (base64 encoded PNG)
   - Stores the secret in the session context
   - Returns prompts for user to scan QR code and enter verification code

2. **Verification Call** (with `totpVerification` input):
   - Validates the provided TOTP code
   - If valid, creates the TOTP attribute and adds it to the user
   - Optionally saves the user to the database if `saveUser` is "true"
   - Returns success condition

### Output Context
- `totpSecret`: The generated TOTP secret
- `totpImageUrl`: Base64 encoded PNG QR code image
- `totpIssuer`: The issuer name
- `totpVerification`: Prompt for verification code

### Result States
- `success`: TOTP successfully created and verified

## verifyTOTP Node

### Purpose
Verifies TOTP codes during authentication and manages failed attempt counters.

### Node Type
`NodeTypeQueryWithLogic`

### Required Context
- User must be present in the authentication session
- User must have a TOTP attribute

### Custom Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_failed_attempts` | string | "10" | Maximum number of failed attempts before locking the TOTP |

### Behavior

1. **Initial Call** (no input):
   - Checks if user has TOTP attribute
   - Returns prompt for verification code

2. **Verification Call** (with `totpVerification` input):
   - Validates the provided TOTP code
   - If invalid:
     - Increments failed attempts counter
     - Locks TOTP if max attempts exceeded
     - Updates database immediately
     - Returns failure condition
   - If valid:
     - Resets failed attempts counter to 0
     - Updates database immediately
     - Returns success condition

### Result States
- `success`: TOTP code verified successfully
- `failure`: TOTP code verification failed
- `no_totp`: User has no TOTP attribute
- `locked`: TOTP is locked due to too many failed attempts

## Security Features

- **Rate Limiting**: Failed attempts are tracked and can lock the TOTP
- **Immediate Persistence**: Failed attempt counters are saved to database immediately
- **Configurable Thresholds**: Maximum failed attempts can be configured per node
- **Automatic Locking**: TOTP is automatically locked when threshold is exceeded

## Usage Examples

### Basic TOTP Enrollment Flow
```yaml
# Create TOTP node configuration
createTOTP:
  type: createTOTP
  customConfig:
    totpIssuer: "MyApp"
    saveUser: "true"
```

### TOTP Verification Flow
```yaml
# Verify TOTP node configuration
verifyTOTP:
  type: verifyTOTP
  customConfig:
    max_failed_attempts: "5"
```

## Integration Notes

- Both nodes require a user to be present in the authentication session
- The `createTOTP` node should be used during user registration or TOTP setup
- The `verifyTOTP` node should be used during login flows after password verification
- Failed attempt counters are persisted immediately to prevent bypassing security measures
- QR codes are generated as base64 encoded PNG images for easy display in web interfaces
