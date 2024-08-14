# tfvars-tools

Go tools for finding and processing tfvars files

## Tools

1. findtfvars - Recursively finds .tfvars files in parent directories
2. tfvars-to-args - Converts a list of .tfvars files to Terraform command-line arguments

## Usage

### findtfvars

```bash
findtfvars
```

### tfvars-to-args

```bash
findtfvars | tfvars-to-args
```

## Development

1. Clone the repository
2. Run `make dev-start` to start the development container
3. Use `make dev-exec CMD='command'` to run commands in the container

## Building

Run `make build` to build the tools.

## Testing

Run `make docker-test` to run tests.

## Releasing

1. Update version in Makefile
2. Run `make release` to create a new GitHub release
