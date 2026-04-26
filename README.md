# go-tpl

Golang project template with the following integrations:

- **[cobra](https://github.com/spf13/cobra)** — CLI command structure
- **[viper](https://github.com/spf13/viper)** — configuration from `config.yaml` (with environment variable overrides)
- **[zerolog](https://github.com/rs/zerolog)** — structured, levelled logging

## Project structure

```
.
├── main.go                   # Entry point
├── config.yaml               # Sample configuration file
├── cmd/
│   ├── root.go               # Root cobra command (loads config + logger)
│   └── version.go            # Example sub-command
└── internal/
    ├── config/
    │   └── config.go         # Viper-based config loader
    └── logger/
        └── logger.go         # Zerolog logger factory
```

## Getting started

```bash
# Build
go build -o go-tpl .

# Run (reads ./config.yaml by default)
./go-tpl version

# Use a custom config file
./go-tpl --config /path/to/config.yaml version
```

## Configuration

Edit `config.yaml`:

```yaml
app:
  name: go-tpl
  version: 0.1.0

log:
  level: info   # debug | info | warn | error
```

Any config key can be overridden with an environment variable (e.g. `APP_NAME`, `LOG_LEVEL`).

## Running tests

```bash
go test ./...
```
