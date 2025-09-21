# YubiKey OTP Authentication

YubiKey OTP (One-Time Password) is a simple yet strong authentication mechanism supported by all YubiKeys out of the box. GoAM provides three nodes for YubiKey integration: `createYubikeyOtp`, `verifyYubikeyOtp`, and `hasYubicoOtp`.

## Overview

YubiKey OTP uses the [Yubico API](https://developers.yubico.com/OTP/) to validate YubiKeys. YubiKeys must be registered with Yubico, and you need to obtain a client ID and API key from [Yubico's API key generator](https://upgrade.yubico.com/getapikey/) to use the validation service.

### Key Features

- **No client software needed**: The OTP is just a string that can be sent like a password
- **YubiKey ID embedded in OTP**: Allows for self-provisioning and authentication without a username
- **Easy to implement**: Using YubiCloud makes supporting YubiKey OTP straightforward
- **Multiple YubiKeys per user**: Users can have multiple YubiKeys associated with their account
- **Unique YubiKey enforcement**: Optional feature to ensure each YubiKey can only be used by one user

## YubiCloud vs Self-Hosted Validation

### YubiCloud (Default)
GoAM uses YubiCloud by default, which provides a single endpoint at `https://api.yubico.com/wsapi/2.0/verify`. This is the recommended approach for most deployments as it offers:
- High availability and reliability
- Automatic scaling
- No maintenance overhead
- Geolocated endpoints for optimal performance

### Self-Hosted Validation
For organizations requiring complete control over their authentication infrastructure, you can configure a custom validation server using the `yubikey-apiurl` option. This requires setting up your own YubiKey Validation Server (YK-VAL) and Key Storage Module (KSM).

## Authentication Flow

The YubiKey OTP validation follows the [Yubico OTP Validation Protocol Version 2.0](https://developers.yubico.com/OTP/Specifications/OTP_validation_protocol.html):

1. **OTP Generation**: When a user presses their YubiKey, it generates a 44-character ModHex-encoded OTP
2. **API Request**: GoAM sends an HTTP GET request to the validation server with the OTP, client ID, and a random nonce
3. **Response Validation**: The server responds with validation status and an HMAC signature for verification
4. **Signature Verification**: GoAM verifies the response signature to ensure authenticity
5. **Public ID Extraction**: The first 12 characters of the OTP represent the YubiKey's public ID

## Node Definitions

### 1. createYubikeyOtp

**Purpose**: Registers a new YubiKey with a user account.

**Node Type**: `QueryWithLogic`

**Required Context**: User must be present in the authentication session

**Configuration Options**:
- `yubikey-apiurl`: The URL of the YubiKey API (defaults to `https://api.yubico.com/wsapi/2.0/verify`)
- `yubikey-clientid`: The Client ID obtained from Yubico
- `yubikey-apikey`: The API Key obtained from Yubico
- `yubikey-checkunique`: If true, ensures the YubiKey can only be used by one user
- `skipsaveuser`: If true, the user won't be saved to the database after OTP creation

**Possible Results**:
- `success`: YubiKey successfully registered
- `failure`: Invalid OTP or verification failed
- `existing`: YubiKey already registered to another user (when `yubikey-checkunique` is enabled)

**Flow**:
1. Prompts user for YubiKey OTP if not provided
2. Validates OTP with Yubico API
3. Checks for uniqueness if required
4. Creates YubiKey attribute and associates it with the user
5. Saves user to database (unless `skipsaveuser` is true)

### 2. verifyYubikeyOtp

**Purpose**: Verifies a YubiKey OTP for authentication.

**Node Type**: `QueryWithLogic`

**Required Context**: User context (can be loaded via YubiKey if `yubikey-checkunique` is enabled)

**Configuration Options**:
- `yubikey-apiurl`: The URL of the YubiKey API
- `yubikey-clientid`: The Client ID for API authentication
- `yubikey-apikey`: The API Key for API authentication
- `yubikey-checkunique`: Enables user lookup by YubiKey
- `yubikey-createuserifnotfound`: Creates a new user if not found (requires `yubikey-checkunique`)

**Possible Results**:
- `success`: OTP verification successful
- `failure`: Invalid OTP or user doesn't have this YubiKey
- `locked`: YubiKey is locked due to security policy
- `notfound`: User not found (when using unique YubiKeys)
- `newuser`: New user created (when `yubikey-createuserifnotfound` is enabled)

**Flow**:
1. Prompts user for YubiKey OTP if not provided
2. Validates OTP with Yubico API
3. If `yubikey-checkunique` is enabled and no user in context:
   - Looks up user by YubiKey public ID
   - Creates new user if `yubikey-createuserifnotfound` is enabled
4. Verifies the user has the specific YubiKey registered
5. Checks if the YubiKey is locked

### 3. hasYubicoOtp

**Purpose**: Checks if a user has any YubiKeys registered.

**Node Type**: `Logic`

**Required Context**: User must be present in the authentication session

**Possible Results**:
- `yes`: User has one or more YubiKeys registered
- `no`: User has no YubiKeys registered

**Flow**:
1. Loads user from context
2. Checks for YubiKey attributes
3. Returns appropriate result based on presence of YubiKeys

## Configuration Examples

### Basic YubiKey Registration Flow

```yaml
nodes:
  - name: "register_yubikey"
    type: "createYubikeyOtp"
    config:
      yubikey-clientid: "12345"
      yubikey-apikey: "dGVzdC1hcGkta2V5"
      yubikey-checkunique: "false"
```

### YubiKey Authentication with User Lookup

```yaml
nodes:
  - name: "authenticate_yubikey"
    type: "verifyYubikeyOtp"
    config:
      yubikey-clientid: "12345"
      yubikey-apikey: "dGVzdC1hcGkta2V5"
      yubikey-checkunique: "true"
      yubikey-createuserifnotfound: "false"
```

### Self-Hosted Validation Server

```yaml
nodes:
  - name: "verify_yubikey_internal"
    type: "verifyYubikeyOtp"
    config:
      yubikey-apiurl: "https://internal-yubikey-server.company.com/wsapi/2.0/verify"
      yubikey-clientid: "12345"
      yubikey-apikey: "dGVzdC1hcGkta2V5"
      yubikey-checkunique: "true"
```

## Security Considerations

### YubiKey Uniqueness
When `yubikey-checkunique` is enabled:
- Each YubiKey can only be registered to one user
- This enables passwordless authentication using only the YubiKey
- The system can look up users by their YubiKey public ID
- Prevents YubiKey sharing between users

### Multiple YubiKeys
Users can register multiple YubiKeys:
- Each YubiKey is stored as a separate attribute
- All registered YubiKeys are valid for authentication
- Provides backup authentication methods
- Useful for users with multiple devices

### Locked YubiKeys
YubiKeys can be locked for security reasons:
- Prevents authentication even with valid OTP
- Can be used for account suspension
- Requires administrative action to unlock
- Failed attempts are not tracked (YubiKeys are considered non-brute-forceable)

## Implementation Details

### HMAC Signature Verification
GoAM implements HMAC-SHA1 signature verification as specified in the Yubico protocol:
- Sorts parameters alphabetically
- Constructs query string
- Generates HMAC-SHA1 signature using API key
- Compares with response signature

### Error Handling
- Network errors are retried automatically
- Invalid OTPs return `failure` (not internal errors)
- Signature verification failures are treated as validation failures
- User not found scenarios are handled gracefully

### Database Storage
YubiKey information is stored as user attributes with the following structure:

**Attribute Type**: `yubico`

**Attribute Value** (`YubicoAttributeValue`):
```json
{
  "public_id": "vvcijgklnrbf",
  "locked": false,
  "failed_attempts": 0
}
```

**Fields**:
- `public_id` (string): The 12-character public identifier extracted from the YubiKey OTP
- `locked` (boolean): Whether the YubiKey is locked for security reasons
- `failed_attempts` (integer): Number of failed authentication attempts (typically 0 for YubiKeys)

**Index**: When `yubikey-checkunique` is enabled, the attribute's `index` field is set to the public ID for fast user lookups.

**Multiple YubiKeys**: Users can have multiple YubiKey attributes, each representing a different physical YubiKey device.

## Flow Examples

### 1. YubiKey-Only Authentication

This flow demonstrates passwordless authentication using only YubiKey OTP. Users can either log in with an existing YubiKey or register a new one.

```yaml
description: 'Login with YubiKey OTP only - passwordless authentication'
start: init
nodes:
  init:
    name: init
    use: init
    custom_config: {}
    next:
      start: verifyYubikeyOtp 

  verifyYubikeyOtp:
    name: verifyYubikeyOtp
    use: verifyYubikeyOtp
    custom_config:
      title: Login or Register
      message: Please enter your Yubikey OTP code
      label: Yubikey OTP Code
      button_text: Validate
      yubikey-checkunique: "true"
      yubikey-createuserifnotfound: "true"
    next:
      success: successResult
      failure: verifyYubikeyOtp
      locked: failureResult
      new_user: successResult

  failureResult:
    name: failureResult
    use: failureResult
    custom_config:
      message: Failed to login.
    next: {}
    
  successResult:
    name: successResult
    use: successResult
    custom_config:
      message: Login successful!
    next: {}
```

**Key Features**:
- `yubikey-checkunique: "true"` - Ensures each YubiKey can only be used by one user
- `yubikey-createuserifnotfound: "true"` - Automatically creates new users when they use an unregistered YubiKey
- Passwordless authentication - Users only need their YubiKey
- Self-provisioning - New users are created automatically

### 2. Username/Password + YubiKey Authentication

This flow combines traditional username/password authentication with optional YubiKey registration or verification.

```yaml
description: 'Login with username and password, then register or verify YubiKey'
start: init
nodes:
  init:
    name: init
    use: init
    custom_config: {}
    next:
      start: askUsernamePassword 
      
  askUsernamePassword:
    name: askUsernamePassword
    use: askUsernamePassword
    custom_config:
      message: Please login to your account
    next:
      submitted: validatePassword

  validatePassword:
    name: validatePassword
    use: validatePassword
    custom_config: {}
    next:
      success: hasYubicoOtp
      fail: askUsernamePassword
      locked: failureResult
      noPassword: failureResult

  hasYubicoOtp:
    name: hasYubicoOtp
    use: hasYubicoOtp
    custom_config: {}
    next:
      yes: verifyYubikeyOtp
      no: createYubikeyOtp

  verifyYubikeyOtp:
    name: verifyYubikeyOtp
    use: verifyYubikeyOtp
    custom_config:
      yubikey-checkunique: false
    next:
      success: successResult
      fail: verifyYubikeyOtp
      locked: failureResult

  createYubikeyOtp:
    name: createYubikeyOtp
    use: createYubikeyOtp
    custom_config:
      yubikey-checkunique: false
    next:
      success: successResult
      fail: failureResult

  failureResult:
    name: failureResult
    use: failureResult
    custom_config:
      message: Failed to login.
    next: {}
    
  successResult:
    name: successResult
    use: successResult
    custom_config:
      message: Login successful!
    next: {}
```

**Key Features**:
- Two-factor authentication - Username/password + YubiKey
- Conditional YubiKey flow - Only prompts for YubiKey if user has one registered
- Optional YubiKey registration - New users can register a YubiKey during login
- `yubikey-checkunique: false` - Allows multiple users to register the same YubiKey (if desired)

### Flow Design Patterns

**Passwordless Authentication**:
- Use `yubikey-checkunique: "true"` to enable user lookup by YubiKey
- Use `yubikey-createuserifnotfound: "true"` for self-provisioning
- Start directly with `verifyYubikeyOtp` node

**Two-Factor Authentication**:
- Start with username/password validation
- Use `hasYubicoOtp` to check if user has YubiKeys
- Branch to `verifyYubikeyOtp` or `createYubikeyOtp` based on user state

**YubiKey Registration Only**:
- Use `createYubikeyOtp` after successful password authentication
- Set `yubikey-checkunique: false` to allow multiple users per YubiKey
- Ensure user is already in context before creating YubiKey

## References

- [Yubico OTP Overview](https://developers.yubico.com/OTP/)
- [OTP Validation Protocol Version 2.0](https://developers.yubico.com/OTP/Specifications/OTP_validation_protocol.html)
- [Self-Hosted OTP Validation Guide](https://developers.yubico.com/OTP/Guides/Self-hosted_OTP_validation.html)
- [Get API Key](https://upgrade.yubico.com/getapikey/)
