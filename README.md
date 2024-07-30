# seaway

### sync

Upload the 

```sh
seactl sync
```

manifest.yaml

```yaml
include:
  - go.*
  - pkg
environments:
  dev:
    command: <command>
    args:
      - <args>
    config:
      path: <path/to/config>
      context: <context_name>
    target:
      bucket: <bucket_name>
      path: <base/path>
    vars:
      - <path/to/envvars>
  test:
    ...
  staging:
    ...
  production:
    ...
```