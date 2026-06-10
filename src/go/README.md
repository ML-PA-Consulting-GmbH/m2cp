
# L-IoT command line tool

This repository contains the `liot` / `m2cp` command line tool.
For each command you can ask for help with the `--help` flag, e.g.:

```command
$ m2cp user --help
Manage the user session
...
```

## Credentials and where to get them

To be able to run `m2cp user login`, you need to ask a colleague from the M2CP or Backend team,
to register your email and SSH public key in the backend.

To be able to install a virtual device, you need an Access Token for the ML!PA Docker Container Registry.
Provided you have the proper access rights, open the [Mircosoft Azure Portal](https://portal.azure.com),
select the **rg-snapstore-dev** Resource Group, click on the **acrsnapstorest02dev** Container Registry.
In the **Repository permissions** section, click on the **Tokens** blade.
(TODO: Well, you can't see the password over there.)

