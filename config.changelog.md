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

### [v? TBD] - Next Version TBD
