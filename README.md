[![Build multi arch Docker](https://github.com/JenswBE/python-dead-link-checker/workflows/Build%20multi%20arch%20Docker/badge.svg)](https://github.com/JenswBE/python-dead-link-checker)
[![codecov](https://codecov.io/gh/JenswBE/python-dead-link-checker/branch/master/graph/badge.svg)](https://codecov.io/gh/JenswBE/python-dead-link-checker)
[![Docker Pulls](https://img.shields.io/docker/pulls/jenswbe/dead-link-checker)](https://hub.docker.com/r/jenswbe/dead-link-checker)

# Dead Link Checker (DeLiC)

Dead link checker written in Python

## 1. Config

```yaml
verbose: False # Optional
workers_per_site: 8 # Optional
internal_links_only: False # Optional

cron: "0 0 * * *" # Optional, run every night

sites:
  - https://jensw.be

notify: # Optional, example for Mailjet using SMTP
  provider: email
  data:
    host: in-v3.mailjet.com
    port: 587
    tls: True
    username: REPLACE_ME
    password: REPLACE_ME
    from: delic@example.com
    to: admin@example.com
    subject: Broken links found
```

See [Notifiers documentation](https://notifiers.readthedocs.io/en/latest/providers/index.html)
for the full list of supported notification providers and their configurations.

## 2. Run DeLiC

By default, DeLiC tries to read the config file at `/config.yml`.

```bash
docker run -v /path/to/config.yml:/config.yml jenswbe/dead-link-checker
```

## 3. CLI arguments
```
-c, --config    Location of the config file
-v, --verbose   Enable verbose output
```
