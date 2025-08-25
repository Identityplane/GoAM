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

**Index Uniqueness Constraint**: Within a realm, the combination of `(tenant, realm, type, index_value)` must be unique. This constraint exists because the index is used as a lookup mechanism to find users by attribute values. For example:
- Only one user can have an email address "john@example.com" as an indexed attribute within a realm
- This ensures that when someone tries to log in with "john@example.com", the system can uniquely identify which user to authenticate
- Multiple users cannot share the same email address, phone number, or social login ID as an indexed attribute within the same realm

**Backup/Secondary Attributes**: Users can have multiple attributes of the same type, but only one should have an index for lookup purposes. Additional attributes (like backup emails, secondary phone numbers) should have their index set to `null`:
- **Primary email**: `index="john@example.com"` (used for login/lookup)
- **Backup email**: `index=null` (not used for lookup, just stored for reference)

**Nullable Index Support**: The index field is nullable, allowing attributes that don't need to be searchable (such as passwords, user preferences, backup contact methods, or internal flags) to exist without an index value.

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

## Database Constraints and Validation

### Unique Constraint
The database enforces a unique constraint on the combination of `(tenant, realm, type, index_value)`:
```sql
UNIQUE (tenant, realm, type, index_value)
```

This constraint ensures:
- **User Identity Uniqueness**: No two users can have the same email address, phone number, or social login ID within the same realm
- **Realm Isolation**: The same index value can exist across different realms
- **Type Separation**: Different attribute types can share the same index value (e.g., an email and username could both be "john.doe")

### Index Value Handling
- **String Indexes**: Must be unique within the realm for the given type (e.g., only one user can have "john@example.com" as an email)
- **Null Indexes**: Multiple attributes with `null` indexes are allowed (e.g., multiple password attributes, user preferences)
- **Empty String Indexes**: Treated as distinct from `null` and must follow uniqueness rules

### Examples of Valid and Invalid Scenarios

**Valid (Allowed):**
- User A: email="john@example.com", phone="+1234567890"
- User B: email="jane@example.com", phone="+0987654321"
- Both users can have password attributes with `null` indexes
- User A can have multiple email attributes:
  - Primary: `index="john@example.com"` (used for login)
  - Backup: `index=null` (stored for reference, not for lookup)

**Invalid (Constraint Violation):**
- User A: email="john@example.com"
- User B: email="john@example.com" (same realm) ❌
- User A: email="john@example.com", username="john@example.com" ✅ (different types)
- User A cannot have two email attributes with `index="john@example.com"` ❌

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
