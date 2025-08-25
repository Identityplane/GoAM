# Telegram Login Nodes

The Telegram nodes provide Telegram OAuth authentication functionality for user login. These nodes allow users to authenticate using their Telegram accounts through the Telegram Login Widget.

## Overview

The Telegram authentication implementation consists of one main node:
- **telegramLogin**: Handles the complete Telegram authentication flow, including redirect to Telegram and processing the authentication result

## Telegram Attribute

The node works with a user attribute of type `"identityplane:telegram"` that contains:

- **TelegramUserID** (`string`): The unique Telegram user ID
- **TelegramUsername** (`string`): The Telegram username/login
- **TelegramFirstName** (`string`): The user's first name from Telegram
- **TelegramPhotoURL** (`string`): URL to the user's Telegram profile picture
- **TelegramAuthDate** (`int64`): Unix timestamp when the authentication occurred

## telegramLogin Node

### Purpose
Handles the complete Telegram authentication flow, including redirect to Telegram, authentication processing, and user management.

### Node Type
`NodeTypeQueryWithLogic`

### Required Context
- No specific context required (handles both initial redirect and callback)

### Custom Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `botToken` | string | Yes | The bot token obtained from [@BotFather](https://telegram.me/BotFather) |
| `requestWriteAccess` | string | No | If "true", requests write access to the user's Telegram account |
| `createUser` | string | No | If "true", creates a new user if they don't exist |

### Behavior

1. **Initial Call** (no `tgAuthResult` input):
   - Generates a Telegram authentication URL
   - Redirects the user to Telegram for authentication
   - Returns a `__redirect` prompt with the Telegram authentication URL

2. **Callback Call** (with `tgAuthResult` input):
   - Parses and verifies the Telegram authentication result
   - Creates a `TelegramAttributeValue` and stores it in the session context
   - Checks if a user with this Telegram ID already exists using the new attribute system
   - Optionally creates a new user if `createUser` is "true"
   - Returns appropriate result state based on user existence

### Output Context
- `telegram`: JSON-serialized Telegram attribute value containing all user data
- `user`: The created or existing user object

### Result States
- `existing-user`: User with this Telegram ID already exists in the system
- `new-user`: This is a new Telegram user (no existing account found)
- `failure`: Authentication flow failed (invalid hash, outdated data, etc.)

## Telegram Authentication Flow

The Telegram authentication follows this flow:

1. **Initial Request**: `telegramLogin` node redirects user to Telegram
2. **Telegram Authentication**: User authenticates with Telegram and grants permissions
3. **Callback Processing**: `telegramLogin` processes the authentication result and verifies the hash
4. **User Management**: Creates or retrieves user account with Telegram attributes
5. **Session Management**: User is now authenticated and can access the system

## Security Features

- **Hash Verification**: Uses HMAC-SHA256 to verify that authentication data comes from Telegram
- **Timestamp Validation**: Ensures authentication data is not outdated (15-minute threshold)
- **Bot Token Security**: Bot token is used as the secret key for hash verification
- **Data Integrity**: All Telegram user data is cryptographically verified

## BotFather Setup

To use Telegram authentication, you need to create a Telegram bot using [@BotFather](https://telegram.me/BotFather). Follow these steps:

### 1. Start BotFather
- Open Telegram and search for [@BotFather](https://telegram.me/BotFather)
- Send `/start` to begin the bot creation process

### 2. Create New Bot
- Send `/newbot` command
- Provide a display name for your bot
- Provide a username for your bot (must end with "bot")

### 3. Configure Domain
- Send `/setdomain` command
- Set your domain where the Telegram Login Widget will be used
- This ensures the widget only works on your authorized domain

### 4. Get Bot Token
- BotFather will provide you with a bot token
- Use this token in your `telegramLogin` node configuration

## Usage Examples

### Basic Telegram Login Flow
```yaml
# Telegram login node configuration
telegramLogin:
  type: telegramLogin
  customConfig:
    botToken: "1234567890:YourBotTokenHere"
    createUser: "true"
    requestWriteAccess: "false"
```

### Complete Authentication Flow
```yaml
# Flow configuration example
flow:
  - telegramLogin
  - success
```

## Integration Notes

- **Bot Setup**: Requires a Telegram bot created through [@BotFather](https://telegram.me/BotFather)
- **Domain Configuration**: Must configure authorized domain using `/setdomain` command
- **User Management**: New users are automatically created with "active" status if `createUser` is enabled
- **Attribute Indexing**: Telegram user ID is used as the attribute index for efficient lookups
- **Database Schema**: Requires user attributes table with JSON value support
- **Error Handling**: Comprehensive error handling for authentication failures and verification errors

## References

- [@BotFather](https://telegram.me/BotFather) - Official bot creation and management tool
- [Telegram Login Widget Documentation](https://core.telegram.org/widgets/login) - Official widget documentation
- [Telegram Bot API](https://core.telegram.org/bots/api) - Bot API reference
