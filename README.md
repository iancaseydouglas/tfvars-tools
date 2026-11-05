# tfv

Layered `.tfvars` for Terraform. Put shared values in parent directories and overrides next to `main.tf`, then run Terraform as usual.

```bash
eval "$(tfv init bash)"    # once, in ~/.bashrc or ~/.zshrc (use zsh there)

cd org/service1/dev
terraform plan
```

## Example

```
org/
├── .git
├── common.tfvars
└── service1/
    ├── service1.tfvars
    └── dev/
        ├── main.tf
        └── dev.tfvars
```

```console
$ cd org/service1/dev
$ tfv status
state: active
root:  /src/org
env:   service1/dev
files, last wins:
  /src/org/common.tfvars
  /src/org/service1/service1.tfvars
  /src/org/service1/dev/dev.tfvars
```

`dev.tfvars` beats `service1.tfvars`, which beats `common.tfvars`. Terraform parses the files itself, so lists, maps and heredocs all work.

Files are resolved each time `terraform` runs, so nothing goes stale when you change directory.

## Usage

```bash
tfv status                    # what terraform would get
tfv -deactivate               # pause the wrapper in this shell
tfv -activate                 # resume it

terraform plan -out=tfplan
terraform apply tfplan        # uses the values saved in the plan
terraform -chdir=svc/dev plan # searches from svc/dev
```

Scripts and CI don't inherit the shell function:

```bash
tfv exec plan                 # same as the wrapper
terraform plan $(tfv -args)   # or explicit args; breaks on paths with spaces
```

Other tools and file types:

```bash
some_tool $(tfv -args -f .toml -format '--file=%s')
```

## Search

`tfv` climbs from the current directory to the one containing `.git`. If there is no `.git` above, it uses the current directory only and warns, so stray files in `~` or `/` are never picked up.

| Flag | Env | Default | |
|---|---|---|---|
| `-stop-at NAME` | `TFV_STOP_AT` | `.git` | stop at the directory containing `NAME`; empty climbs to `/` |
| `-max-depth N` | `TFV_MAX_DEPTH` | `-1` | parent levels to climb; `-1` is unlimited |
| `-d DIR` | | `.` | start directory |
| `-f EXT` | | `.tfvars` | extension to match |
| `-format FMT` | | `-var-file=%s` | argument format for `-args` |

Flags override env vars, and both apply to `status` and `-args`. `tfv exec` reads only the env vars.

Var files are added to `plan`, `apply`, `destroy`, `refresh`, `import` and `console`, ahead of any existing `TF_CLI_ARGS_<command>`.

## Install

```bash
curl -L https://github.com/iancaseydouglas/tfvars-tools/releases/latest/download/tfv_Linux_x86_64.tar.gz | tar xz tfv
sudo mv tfv /usr/local/bin/
```

Replace `Linux` with `Darwin` on macOS and `x86_64` with `arm64` on ARM. Or:

```bash
go install github.com/iancaseydouglas/tfvars-tools/cmd/tfv@latest
```

## Develop

```bash
make        # vet, test, gofmt check, build to bin/tfv
make cross  # compile for other platforms
```

CI runs the same targets. Pushing a `v*` tag publishes a release.
