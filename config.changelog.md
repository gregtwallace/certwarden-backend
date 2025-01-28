# Config File Schema Changelog

This file tracks the changes made to the config.yaml file format over time. Some
changes do not require the schema to be incrmented, while others do. This will
also be noted in the log.

When you update, you should review changes that occured between your current
version and the new version. While migration will occur automatically for many
changes, it does not mean there aren't new features or options you may want to
take advantage of.


## Schema Updates

- Unknown -> any:
  + Manual intervention is required, review config.default.yaml,
    config.example.yaml, and this log.
- <0 -> any: 
  + Manual intervention is required, review config.default.yaml,
    config.example.yaml, and this log.
- 0 -> any:
  + Manual intervention is required, review config.default.yaml,
    config.example.yaml, and this log.
- 1 -> 2 -> 3 -> 4:
  + Automatic migration will occur.


## Log

### [v0.14.1] - 2023.10.17

- 2023.10.23
  + log begins
  + config_version is 1

### [v0.15.0] - 2023.10.23

- 2023.10.23
  https://github.com/gregtwallace/certwarden-backend/commit/523d2bf8dd8a8e5c4a43714ba7c728f7b4084c47
  + `cors_permitted_origins` RENAMED to `cors_permitted_crossorigins`
  + config_version incremented from 1 to 2

- 2023.10.23
  https://github.com/gregtwallace/certwarden-backend/commit/7b419a23ead8ccb7db9b48a379117f6df23c82a5
  + implement strict config_version schema enforcement (automatically update schema
    or if not possible, LeGo will not start)

- 2023.10.23
  https://github.com/gregtwallace/certwarden-backend/commit/db879ddf3e63e720d921f346b343ea5b1f2f7787
  + `disable_hsts` config option ADDED to allow disabling of HTTP Strict Transport
    Security (HSTS) header

- 2023.10.23
  https://github.com/gregtwallace/certwarden-backend/commit/978eecfb6b86f580a35ce7344c1aedc2a6bdb8eb
  + `frontend_show_debug_info` config option ADDED that controls if the frontend
    will show debug info (if it is being hosted by the backend)

### [v0.16.1] - 2023.12.03

- 2023.11.29
  https://github.com/gregtwallace/certwarden-backend/commit/172fea183414c51d531fc98016142774c24737d7
  + config_version incremented from 2 to 3
  + `pprof_port` RENAMED to `pprof_http_port`
  + added `pprof_https_port` for pprof https port


### [v0.17.0] - 2023.12.20
- 2023.12.20
  https://github.com/gregtwallace/certwarden-backend/commit/b80d47f7eff2df37cb9dd3b71299335e811c3515
  + config_version not incremented (no breaking changes)
  + add `backup` section with options to enable and disable automatic backup and 
    options to specify criteria for deletion of old backups.

### [v0.22.1] - 2024.09.07
- 2024.08.05
  https://github.com/gregtwallace/certwarden-backend/commit/e6acbec9b58ba6196740fe3a2394e18df39d34f2
  + auto ordering value `valid_remaining_days_threshold` removed and instead will be calculated 
    based on percentage of a certificate's validity remaining

### [v0.23.0] - 2024.12.07
- 2024.11.29
  https://github.com/gregtwallace/certwarden-backend/commit/9989ae9cfa08d30d0acf31cc18135d55e9a31316
  + add `domain_aliases` under `challenges`. This is not a breaking change.

### [v0.24.0]
- https://github.com/gregtwallace/certwarden-backend/commit/a5354ae23f4e46d33521066331c78e1a5a3c66e0
  + add `auth` config section to enable/disable local auth and OIDC

### [v0.24.3]
- https://github.com/gregtwallace/certwarden-backend/commit/fb9392ca1b61e3ded34836fc37a5263ac7f67e35
  + Remove `frontend_show_debug_info` in favor of a browser storage based
    value that can be toggled in the frontend.
