# doorman-keepassxc

A [Doorman](https://github.com/AmadlaOrg/doorman) plugin for retrieving secrets from [KeePassXC](https://keepassxc.org/) password databases.

## Requirements

- `keepassxc-cli` must be installed and available on PATH

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `KEEPASSXC_DB` | Yes | Path to `.kdbx` database file |
| `KEEPASSXC_KEYFILE` | No | Path to key file |
| `KEEPASSXC_PASSWORD` | No | Database password (if unset, `keepassxc-cli` reads from stdin) |

## Usage

### Plugin metadata

```bash
doorman-keepassxc info
```

### Retrieve a secret

```bash
export KEEPASSXC_DB=/path/to/passwords.kdbx
export KEEPASSXC_PASSWORD=masterpassword

doorman-keepassxc get "Root/myapp/db-password"
```

### Via Doorman

```bash
doorman get "Root/myapp/db-password" --from keepassxc
```

## Build

```bash
make build
make test
```
