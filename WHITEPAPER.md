# GoIAM Whitepaper

## Concept Overview

GoIAM is a cloud-native IAM built as a lightweight standalone service (Go binary), optimized for Kubernetes environments, but also capable of running standalone outside clusters.

It focuses on:

- **Fully customizable authentication flows**: Code-defined flows based on state-machine execution, offering endless possibilities from low-friction onboarding to complex KYC onboarding processes
- **Declarative authorization**: Rego/OPA integrated
- **GitOps-first configuration management**: Flows, policies, and realms as code
- **Performance and horizontal scaling**: First-class goals
- **Multi-tenancy**: Strong isolation baked in

## Flows Engine

Authentication is modeled as flexible flow graphs, where:

- **Nodes**: Small reusable steps (e.g., "Ask for Username", "Discover Passkey", "Check Group Membership")
- **Flows**: Linked nodes with conditional transitions
- **Configuration**: Flows are defined in YAML, version-controlled, and can be visualized and edited via a simple GUI

Custom Nodes can be written in Go, enabling extensions without waiting for platform updates.

This is designed to support:
- Custom login flows
- Onboarding processes
- Step-up authentication
- Impersonation flows
- Device linking
- And more — not just standard username/password

## Architecture and Deployment

GoIAM is implemented as a standalone Go executable. It is intended to run:

- As a containerized service in Kubernetes clusters (preferred for production)
- As a standalone binary on VM infrastructure (flexible for legacy or edge deployments)

### Configuration Models

1. **Immutable Configuration**
   - YAML-based configuration injected into containers at runtime
   - Recommended for enterprise GitOps flows
   - Enables rollbacks, A/B testing, version control, and performance optimization
   - Flows are memory-resident

2. **Dynamic Configuration**
   - Configuration stored in the database
   - Mutable at runtime via Admin API or Terraform provider
   - Provides operational flexibility

For on-premise deployments, the immutable configuration is the recommended approach if the organization has GitOps capabilities. The dynamic configuration option is suitable for SaaS models or development environments.

## Performance and Scalability

GoIAM is architected for horizontal scalability and high throughput:

- Stateless service design
- Minimal critical database reads per flow execution
- Any non-critical writes (e.g., login counters, audit events)
- Short-TTL memory caching for flows and realms with DB-backed config

### Database Design

- Loose coupling of data objects (no foreign keys)
- Shardable by tenant, realm, and (optionally) geolocation
- ACID semantics reserved only for critical security updates (e.g., password trial increments)

### Designed to Support

- 1,000+ realms
- 100,000+ logins per minute
- 1,000,000+ users

## Security Architecture

GoIAM is built with zero-trust internal principles:

- OAuth 2.1 per default
- Support for FAPI 2.0 Security Profile
- Tamper-proof audit logs stored in database
- Structured logging with correlation IDs (flow-execution-ID, request-ID, user-ID)
- Secrets integration with external managers (Vault, AWS Secrets Manager)
- Rate limiting and abuse prevention integrated at API level
- Security-critical events (e.g., password changes) are immediately logged and auditable

## Database and Storage

### Primary Backend Database
- PostgreSQL
- Optional: CockroachDB (for multi-region HA needs)

### Session Storage Options
- In-memory caching
- Redis cluster
- Database

All components are intentionally relational-friendly using simple data types (string, integer) to support portability across cloud providers and DB vendors.

## Multi-Tenancy Model

GoIAM implements a two-layer multi-tenancy system:

1. **Tenants**
   - Organizational level (e.g., different business units, regional divisions)

2. **Realms**
   - Security boundaries under each tenant (e.g., Development, Production realms)

Admin access and security boundaries are enforced separately at the tenant and realm levels. This model scales across SaaS, hybrid, and on-premises deployments with strong isolation guarantees.
