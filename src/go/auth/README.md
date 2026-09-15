# auth

This module contains:
- cryptography functions 
- signing functions (signing, verification, ..)

Only the cryptographic parts should remain in the "auth" module.

## --method m2m

Login flow intended for machine-to-machine auth. Requires a client_id and a client_secret.
Auth0 additionally requires --org_id to be set for this method.

### Testing with mlpa-iot-software-services keycloak

`docker compose up` in mlpa-iot-software-services repository root.
(Or if the latest build from sources is needed: `docker compose -f docker-compose.yml -f docker-compose-overrides-build.yml up --build`)

To create a client.

- Access http://localhost:46800 and login with credentials listed in the docker-compose.yml for the keycloak service.
- Manage Realms: select local-realm
- Clients -> Create Client
  - General Settings 
    - Client Type: OpenID Connect
    - Client ID: m2cp-cli-m2m-test
  - Capabilities
    - Client authentication: ON 
    - Service Accounts roles: ON
  - Save
- Select m2cp-cli-m2m-test client in the Clients list
- Tab Client scopes: m2cp-cli-m2m-test-dedicated
- Add mapper -> By configuration -> Hardcoded claim
  - Name: m2m-permissions
  - Token Claim Name: permissions
  - Claim value: *.read.all
  - Claim JSON Type: String
  - Add to access token: ON
  - Add to token introspection: ON
  - all others: OFF
  - Save
- Tab Credentials: copy the Client Secret to set as env for the command below

```sh
M2CP_CLIENT_SECRET=<copied-client-secret> m2cp user login --method m2m \
--client-id m2cp-cli-m2m-test \
--org-id local \
--token-endpoint http://localhost:46800/realms/local-realm/protocol/openid-connect/token \
--audience account \
--store http://localhost:46900/graphql
```

#### local keycloak and oidc discovery

For the backend to return keycloak as the authentication provider, further configuration is needed.
Add an additional `-f docker-compose-override-local-keycloak.yaml`. It should have the following contents:

```yaml
x-kc-local-env: &kc-local-env
  # Stop Auth0 being selected as the discovery / browser-login provider.
  IDENTITY_PROVIDERS:AUTH0:IMPLICIT_FLOW:ENABLED: "false"
  # Make Keycloak the discovery / browser-login provider.
  # CLIENT_ID + SCOPE are mandatory whenever implicit flow is enabled, but
  # for /auth/oidc-discovery they only need to be non-empty (no client
  # lookup happens at startup). Point CLIENT_ID at a real public client and
  # set a valid .../auth/sessions/callback REDIRECT_URL only if you also
  # want the interactive browser implicit-flow login to work via Keycloak.
  IDENTITY_PROVIDERS:KEYCLOAK:IMPLICIT_FLOW_ENABLED: "true"
  IDENTITY_PROVIDERS:KEYCLOAK:IMPLICIT_FLOW_CLIENT_ID: "local-app"
  IDENTITY_PROVIDERS:KEYCLOAK:IMPLICIT_FLOW_SCOPE: "openid"
  # docker internal ip: localhost is not reachable by other services and keycloak as hostname is not reachable by m2cp cli
  # otherwise use m2cp user login with --token-endpoint http://localhost:46800/realms/local-realm/protocol/openid-connect/token
  IDENTITY_PROVIDERS:KEYCLOAK:METADATA_ADDRESS: "http://172.31.255.20:8080/realms/local-realm/.well-known/openid-configuration"

services:
  gateway-service:
    environment:
      <<: *kc-local-env

  user-service:
    environment:
      <<: *kc-local-env

  asset-service:
    environment:
      <<: *kc-local-env

  appstore-service:
    environment:
      <<: *kc-local-env

  monitoring-service:
    environment:
      <<: *kc-local-env

  keycloak:
    networks:
      default:
        ipv4_address: 172.31.255.20 # ensure a static IP such that a fixed metadata address can be used above

networks:
  default:
    ipam:
      config:
        - subnet: 172.31.255.0/27 # 32 addresses, at the end of docker's typical allocation block to avoid collisions

```

### Testing with mlpa-iot-software-services auth0

Need to create a client with an associated secret
