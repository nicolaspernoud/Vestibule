# Vestibule

Alternate URL (on the Grand Lyon software forge) : https://forge.grandlyon.com/NPERNOUD/vestibule .

Vestibule is an HTTP server, reverse proxy and webdav server, with OIDC/OAuth2 and local users file authentication, and a _Single Page Application_ GUI to configure everything.

[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=security_rating)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=alert_status)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=bugs)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=code_smells)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=sqale_index)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=ncloc)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=nicolaspernoud_Vestibule&metric=duplicated_lines_density)](https://sonarcloud.io/dashboard?id=nicolaspernoud_Vestibule)

## Features

- Reverse proxy any website (internal or external), with GUI configuration
- Authenticate users against OIDC/OAuth2 server and fetch user roles based on /userinfo endpoint
- Expose any file system directory as webdav server with web explorer
- Allow opening and saving of documents with onlyoffice integration
- Automatic let's encrypt https

### Screenshots

![Login screen](miscellaneous/images/login.png "Login screen")
![Application configuration](miscellaneous/images/app_config.png "Application configuration")
![Applications list](miscellaneous/images/apps_list.png "Applications lists")
![Opened application](miscellaneous/images/opened_app.png "Opened application")
![Dav configuration](miscellaneous/images/dav_config.png "Dav configuration")
![Davs list](miscellaneous/images/davs_list.png "Davs list")
![Opened dav](miscellaneous/images/opened_dav.png "Opened dav")
![File preview](miscellaneous/images/file_preview.png "File preview")
![Users management](miscellaneous/images/users_management.png "Users management")

### How does it works ?

Vestibules authenticates the users against a local user.json file or an OIDC/OAuth2 provider (with Open Id Connect userinfo endpoint). It issues an encrypted cookie for the global domain (say vestibule.io) containing the user roles gotten from the local user database or the "memberOf" claim of the OIDC/OAuth2 user gotten from the userinfo endpoint.

After, for every access to a proxied application or a webdav service (say myapp.vestibule.io), it checks the cookie to allow users based on their roles.

Applications and davs can be opened to everyone as well (no authentication).

Vestibule creates a subdomain for every services (apps and davs) and provide Let's encrypt certificates automatically.

## Installation

### Locally

Clone the repository.
Alter the `.env` file with your configuration.
Launch `start.sh`.

### With docker

Alter `.env` and `docker-compose.yml` according to your needs.
Launch with `COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 docker-compose up`.
A production deployment example is also provided in the production-deployment-example.sh file.
The mock ip geodatabase should be replaced with a real one from maxmind for real usefullness.

## Usage

### Configuration

Configuration is done through environment variables. The meaning of the different environment variables is detailed here :

| Environment variable     | Usage                                                                                                                                    | Default                             |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| HOSTNAME                 | Vestibule main hostname : needed to know when to respond with the main GUI instead of an application on a webdav service                 | vestibule.127.0.0.1.nip.io          |
| APPS_FILE                | Apps configuration file path                                                                                                             | "./configs/davs.json"               |
| DAVS_FILE                | Davs configuration file path                                                                                                             | "./configs/davs.json"               |
| LETS_CACHE_DIR           | Let's Encrypt cache directory                                                                                                            | "./letsencrypt_cache"               |
| LOG_FILE                 | Optional file to log to                                                                                                                  | defaults to no file logging         |
| HTTPS_PORT               | HTTPS port to serve on                                                                                                                   | 443                                 |
| HTTP_PORT                | HTTP port to serve on, only used for Let's Encrypt HTTP Challenge                                                                        | 80                                  |
| DEBUG_MODE               | Debug mode, disable Let's Encrypt, enable CORS and more logging                                                                          | false                               |
| ADMIN_ROLE               | Admin role                                                                                                                               | ADMINS                              |
| REDIRECT_URL             | Redirect url used by the idp to handle the callback                                                                                      |                                     |
| CLIENT_ID                | Client id to authenticate with the IdP for OAuth2 authentication                                                                         |                                     |
| CLIENT_SECRET            | Client id to authenticate with the IdP for OAuth2 authentication                                                                         |                                     |
| AUTH_URL                 | IdP's authentication URL                                                                                                                 |                                     |
| TOKEN_URL                | IdP's token URL                                                                                                                          |                                     |
| USERINFO_URL             | IdP's userinfo URL                                                                                                                       |                                     |
| ISSUER                   | IdP's issuer for autoconfiguration of AUTH_URL, TOKEN_URL, USERINFO_URL if they are not already set                                      |
| LOGOUT_URL               | IdP's logout URL                                                                                                                         |                                     |
| ONLYOFFICE_TITLE         | Title used on the OnlyOffice document editor window                                                                                      | VestibuleOffice                     |
| ONLYOFFICE_SERVER        | Url of the OnlyOffice document server used to edit documents                                                                             |                                     |
| INMEMORY_TOKEN_LIFE_DAYS | Lifetime of authentication tokens for local users                                                                                        | 1                                   |
| DISABLE_LETSENCRYPT      | Disable Let's Encrypt certificates (in normal mode) and use development certificates (./dev_certificates/localhost.crt and .key) instead | false (true if HOSTNAME is not set) |

### OIDC/OAuth2 configuration

The OIDC/OAuth2 provider is configured with environment variables. The user is recovered with the /userinfo endpoint (part of the Open Id Connect standard) with a standard OAuth2 dance.
Vestibule is compatible with most OpenIdConnect providers (including Keycloak), or OAuth2 providers with the /userinfo endpoint.

The users roles **must** be recovered in an "memberOf" claim array obtained when accessing the /userinfo endpoint. It can be configured to map any group/role configuration on most IdPs.

### Mounting webdav share on your OS

Vestibule allow using the login with the password **OR** the authentication token in an basic auth header to allow mounting webdavs.

### Override branding

Every branding asset is in `web/assets/brand` directory. They can be altered according to your needs.

## Development

### Update dependencies

```bash
go get -u -t ./...
go mod tidy
```

### Register both remotes

```bash
git remote add forge https://forge.grandlyon.com/NPERNOUD/vestibule.git
git remote set-url --add --push origin https://forge.grandlyon.com/NPERNOUD/vestibule.git
git remote set-url --add --push origin https://github.com/nicolaspernoud/Vestibule.git
```

### Get all branches

```bash
git fetch --all
```

## Update master from development and set development to follow master

```bash
git checkout master
git merge development --squash
# Alter commit message and commit
git checkout development
git reset --hard master
```

## Credits

Loosely based on Webfront (https://github.com/nf/webfront), by Andrew Gerrand, Google (Apache License, Version 2.0).

Uses :

- Bulma : https://bulma.io/, https://github.com/jgthms/bulma (MIT Licence)
- Animate.css : https://daneden.github.io/animate.css/, https://github.com/daneden/animate.css (MIT Licence)
- Font Awesome : https://fontawesome.com, https://github.com/FortAwesome/Font-Awesome (Font Awesome Free License)
- Secure IO : https://secure-io.org, https://github.com/secure-io/sio-go (MIT Licence), lots of thanks to @aead (Andreas Auernhammer) who has been a great help in understanding the library and for his cryptography insights
- MaxMind DB Reader for Go : https://github.com/oschwald/maxminddb-golang (ISC License)
- HTTP Cache by Victor Springer : https://github.com/victorspringer/http-cache (MIT Licence), parts are included in pkg/cache directory (to avoid getting unwanted redis dependencies)
- Go-Glob by Ryan Uber : https://github.com/ryanuber/go-glob (MIT Licence)
- go-disk-usage by ricochet2200 : https://github.com/ricochet2200/go-disk-usage (The Unlicense)

## Licence

The product is licenced under **_GNU AFFERO GENERAL PUBLIC LICENSE Version 3_**, it is made primarily by Nicolas Pernoud, a member of **Métropole de Lyon**, on professional time (some), and personal time (most). It is used on Métropole de Lyon "alpha" lab to allow quick prototyping and proof of concepts.

<img src="miscellaneous/images/logo_alpha_couleurs_RVB.png" alt="alpha logo" style="height:100px;">

## Beeing part of the project

Contributions of any kind welcome!
