# User Attributes Specification

## Overview

User attributes represent a flexible and extensible way to store additional information about users in the system. Each attribute has a one-to-one mapping with a user and can store various types of data including credentials, personal information, application-specific settings, and business logic data.

## Attribute Structure

### Core Fields

- **UserID**: Unique identifier linking the attribute to a specific user
- **Tenant**: Organization context for multi-tenancy
- **Realm**: Authentication realm context
- **Index**: Unique lookup key within a realm (e.g., email address for email lookup)
- **Type**: Category of the attribute (e.g., "email", "social", "password")
- **Value**: The actual attribute data (stored as interface{} for flexibility)
- **CreatedAt**: Timestamp when the attribute was created
- **UpdatedAt**: Timestamp when the attribute was last modified

### Index Field

The index field serves as a lookup mechanism to find users by specific attribute values. For example:
- Email attributes use the email address as the index for user lookup
- Social login attributes use a combination of provider and ID as the index
- Phone attributes use the phone number as the index

This enables efficient user discovery without requiring full user object retrieval.

## Attribute Types

The system supports multiple attribute types, each with their own value structure:

- **Social**: OAuth/SSO provider information
- **Password**: Authentication credentials and security status
- **Email**: Contact information with verification status
- **Phone**: Contact information with verification status
- **TOTP**: Multi-factor authentication setup
- **Passkey**: WebAuthn credential information
- **UserProfile**: Basic profile information
- **UserPicture**: Profile image references
- **Device**: Trusted device information

## Multiple Attributes of Same Type

Users can have multiple attributes of the same type. For example:
- Multiple email addresses (primary, work, personal)
- Multiple social login accounts (Google, GitHub, Microsoft)
- Multiple phone numbers (mobile, work, home)

Each attribute instance has a unique identifier within the user's attribute collection.

## API Endpoints

### User Attributes Management

```
GET /admin/{tenant}/{realm}/users/{id}/attributes
POST /admin/{tenant}/{realm}/users/{id}/attributes
```

- **GET**: Returns all attributes for the user (without detailed values)
- **POST**: Creates a new attribute for the user

### Specific Attribute Management

```
GET /admin/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id}
PATCH /admin/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id}
DELETE /admin/{tenant}/{realm}/users/{id}/attributes/{attribute-type}/{attribute-id}
```

- **GET**: Retrieves a specific attribute instance with full details
- **PATCH**: Updates an existing attribute instance
- **DELETE**: Removes a specific attribute instance

## Use Cases

### User Lookup
Attributes with indexes enable efficient user discovery:
- Find user by email address
- Find user by social login ID
- Find user by phone number

### Flexible Data Storage
The interface{} value field allows for:
- Structured data storage (JSON objects)
- Simple value storage (strings, numbers)
- Complex nested data structures

### Multi-Value Support
Users can maintain multiple instances of the same attribute type:
- Multiple contact methods
- Multiple authentication factors
- Multiple social accounts
