
# L-IoT command line tool

This repository contains the `liot` / `m2cp` command line tool.
For each command you can ask for help with the `--help` flag, e.g.:

```command
$ m2cp user --help
Manage the user session
...
```

## Credentials and where to get them

To be able to run `m2cp user login --method browser`, your user account needs to be setup with the relevant permissions at the identity provider.
SSH login is discouraged and will be removed in the future. Please use the browser login flow instead.

To be able to install a virtual device, you need an Access Token for the ML!PA Docker Container Registry.
Provided you have the proper access rights, open the [Mircosoft Azure Portal](https://portal.azure.com),
select the **rg-snapstore-dev** Resource Group, click on the **acrsnapstorest02dev** Container Registry.
In the **Repository permissions** section, click on the **Tokens** blade.
(TODO: Well, you can't see the password over there.)

## Machine-to-machine (M2M) login

CI pipelines and automation can authenticate **without a browser or a personal
account** using `--method m2m` (the OAuth 2.0 client-credentials grant).

### Prerequisites

- An M2M application registered with the identity provider and granted to your
  organization — you receive a **client id** and a **client secret**.
- Permissions have to be set up on the identity provider for the M2M application depending on the operations it needs to perform 
  (e.g. `*.read.all` for read access to everything; consult backend documentation for precise permissions).
- A backend that exposes the machine-login discovery route. Against an older
  backend the CLI fails with a clear *"backend does not support machine login;
  upgrade …"* message.  
  Discovery can be bypassed with `--token-endpoint` and `--audience`.

### Logging in

The client secret is supplied via the `M2CP_CLIENT_SECRET` environment variable (or `--client-secret-stdin`).
**Never** as a command-line flag, and it is **never** written to `~/.m2cp/state.json`.

```bash
export M2CP_CLIENT_SECRET='********'
m2cp user login --method m2m \
  --client-id <client-id> \
  --org-id <org-id> \
  --store https://example.com/graphql
```

Or pipe the secret via stdin:

```bash
cat path/to/file-with-secret | m2cp user login --method m2m \
  --client-id <client-id> \
  --org-id <org-id> \
  --store https://example.com/graphql --client-secret-stdin
```

The access token is stored in `~/.m2cp/state.json` and reused by subsequent `m2cp` commands until it expires. 
Re-run the login to obtain a fresh token.

### Pipeline usage

Log in once per job, then run as many commands as needed with the cached token
(the secret is only required for the login step):

```yaml
steps:
  - bash: m2cp user login --method m2m --client-id $(M2M_CLIENT_ID) --org-id $(M2M_ORG_ID) --store $(STORE_URL)
    env:
      M2CP_CLIENT_SECRET: $(M2M_CLIENT_SECRET)   # exposed to this step only
    displayName: Machine login
  - bash: m2cp snap push ./my-snap_1.2.3_arm64.snap
    displayName: Publish snap
```

Prefer reusing the cached token across a job over re-authenticating before every
command — each token request counts against the identity provider's M2M token
quota.

To check the remaining lifetime of the cached token, run:

```bash
m2cp user status --json | jq --raw-output '.output.session.jwtExpirationTime'
```

### Multiple organizations or stores

A session is scoped to one organization (`--org-id`). For work spanning several
organizations or store, log in again per org or store (each login refreshes the active session).
To cache multiple sessions, use `--state` to specify a different state file per org or store:

```sh
m2cp user login --method m2m --client-id <client-id> --org-id org1 --store https://example.com/graphql --state ~/.m2cp/state-org1.json
m2cp user login --method m2m --client-id <client-id> --org-id org2 --store https://example.com/graphql --state ~/.m2cp/state-org2.json
m2cp snap push ./my-snap_1.2.3_arm64.snap --state ~/.m2cp/state-org1.json
m2cp snap push ./my-snap_1.2.3_arm64.snap --state ~/.m2cp/state-org2.json
```
