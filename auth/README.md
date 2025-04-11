# Auth - Authentication and Authorization service

Auth service provides authentication features as an API for managing authentication keys as well as administering groups of entities - `clients` and `users`.

## Authentication

User service is using Auth service gRPC API to obtain login token or password reset token. Authentication key consists of the following fields:

- ID - key ID
- Type - one of the three types described below
- IssuerID - an ID of the Mitras User who issued the key
- Subject - user ID for which the key is issued
- IssuedAt - the timestamp when the key is issued
- ExpiresAt - the timestamp after which the key is invalid

There are four types of authentication keys:

- Access key - keys issued to the user upon login request
- Refresh key - keys used to generate new access keys
- Recovery key - password recovery key
- API key - keys issued upon the user request
- Invitation key - keys used to invite new users

Authentication keys are represented and distributed by the corresponding [JWT](jwt.io).

User keys are issued when user logs in. Each user request (other than `registration` and `login`) contains user key that is used to authenticate the user.

API keys are similar to the User keys. The main difference is that API keys have configurable expiration time. If no time is set, the key will never expire. For that reason, API keys are _the only key type that can be revoked_. This also means that, despite being used as a JWT, it requires a query to the database to validate the API key. The user with API key can perform all the same actions as the user with login key (can act on behalf of the user for Client, Channel, or user profile management), _except issuing new API keys_.

Recovery key is the password recovery key. It's short-lived token used for password recovery process.

The following actions are supported:

- create (all key types)
- verify (all key types)
- obtain (API keys only)
- revoke (API keys only)

## Domains

Domains are used to group users and clients. Each domain has a unique alias that is used to identify the domain. Domains are used to group users and their entities.

Domain consists of the following fields:

- ID - UUID uniquely representing domain
- Name - name of the domain
- Tags - array of tags
- Metadata - Arbitrary, object-encoded domain's data
- Alias - unique alias of the domain
- CreatedAt - timestamp at which the domain is created
- UpdatedAt - timestamp at which the domain is updated
- UpdatedBy - user that updated the domain
- CreatedBy - user that created the domain
- Status - domain status
