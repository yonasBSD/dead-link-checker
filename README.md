# Dead Link Checker (DeLiC)

Dead link checker written in Python

## 1. Config

```yaml
sites:
  - https://jensw.be
verbose: False # Optional
workers_per_site: 8 # Optional
```

## 2. Run DeLiC

By default, DeLiC tries to read the config file at `/config.yml`.

```bash
docker run -v /path/to/config.yml:/config.yml jenswbe/dead-link-checker
```
