# tfvars-tools

Layered `.tfvars` for Terraform root configs. Put shared values in parent directories and overrides in the environment directory. Terraform reads them in order, and the last value set wins.

- `findtfvars` lists `*.tfvars` from the current directory up through its parents.
- `tfvars-to-args` turns that list into `-var-file=` args, outermost first.

```bash
terraform plan $(findtfvars | tfvars-to-args)
```

## Example

```
/org
├── .git
├── common.tfvars
└── service1
    ├── service1.tfvars
    └── dev
        ├── main.tf
        └── dev.tfvars
```

```console
$ cd /org/service1/dev
$ findtfvars | tfvars-to-args
-var-file=/org/common.tfvars -var-file=/org/service1/service1.tfvars -var-file=/org/service1/dev/dev.tfvars
```

`dev.tfvars` overrides `service1.tfvars`, which overrides `common.tfvars`.

Stop at the repo root so stray files in `~` or `/` are ignored:

```bash
findtfvars -stop-at .git | tfvars-to-args
```

Climb at most two levels:

```bash
findtfvars -max-depth 2 | tfvars-to-args
```

Search from somewhere else, and show what was picked:

```bash
terraform -chdir=service1/dev plan $(findtfvars -d service1/dev | tfvars-to-args -verbose)
```

Other tools and file types:

```bash
some_tool $(findtfvars -f .toml | tfvars-to-args -format '--file=%s')
```

## Options

```
findtfvars
  -d DIR            start directory (default .)
  -f EXT            extension to match (default .tfvars)
  -max-depth N      parent levels to climb; -1 unlimited, 0 start dir only (default -1)
  -stop-at NAME     stop at the first directory containing NAME, e.g. .git

tfvars-to-args
  -format FMT       printf format per file (default -var-file=%s)
  -reverse          keep input order (innermost first)
  -verbose          print files in precedence order to stderr
```

| Env var | Equivalent |
|---|---|
| `FINDTFVARS_MAX_DEPTH` | `-max-depth` |
| `FINDTFVARS_STOP_AT` | `-stop-at` |
| `TFVARS_TO_ARGS_FORMAT` | `-format` |

Flags override env vars. Paths containing spaces break when the output is passed through `$(...)`.

## Install

```bash
curl -L https://github.com/iancaseydouglas/tfvars-tools/releases/latest/download/tfvars-tools_Linux_x86_64.tar.gz | tar xz
sudo mv findtfvars tfvars-to-args /usr/local/bin/
```

Replace `Linux` with `Darwin` on macOS and `x86_64` with `arm64` on ARM. Or build from source:

```bash
go install github.com/iancaseydouglas/tfvars-tools/cmd/...@latest
```
