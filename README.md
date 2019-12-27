# Vestibule

Vestibule is an HTTP server and reverse proxy, with OAuth2 / local users file authentication, and a nice GUI to configure everything.
Loosely based on Webfront (https://github.com/nf/webfront), by Andrew Gerrand, Google.

example :

```bash
. ./.env
vestibule -apps=./configs/apps.json -letsencrypt_cache=./letsencrypt_cache
```

and go to HOSTNAME to configure apps (do not forget to login before !)

## Development

# Update dependencies

```bash
go get -u
go mod tidy
```
