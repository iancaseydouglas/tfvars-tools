# tfvars-tools

Tools for managing root configuration repos for multi-environment or multi-service cloud infrastructure.

> _**Root configurations** or **root modules** are the working directories from which you run the Terraform CLI._ ℹ️
>
## TLDR

`findtfvars`: Recursively finds .tfvars files in _*parent directories*_

`tfvars-to-args`: Converts a list of .tfvars files to Terraform command-line arguments.

### Quick usage

Find all `.tfvars` files in **_parent directories_** and convert to Terraform arguments:
```bash
findtfvars | tfvars-to-args
```

Use in a Terraform command:

```bash
TF_ARGS=$(findtfvars | tfvars-to-args)
terraform plan ${TF_ARGS}
```
## Overview


Whether your root config repo is structured as a
- `Monorepo`
- `repo-per-environment`
- `repo-per-team`
- `repo-per-service`

your configurations will fall into two categories:

1. **Environment Independent**
   - Variables may be set at any parent directory common to all environments
   - Organizational hierarchies can set variables to standardize cloud estate

2. **Environment Specific**
   - Variables should be set in the environment directory
   - Values take precedence over those established generally

> #### Simplify ~~Simplify Simplify~~

`tfvars-tools` simplifies the management of Terraform variables across different organizational structures and environments, ensuring consistency and flexibility in your cloud infrastructure configurations, while
[Keeping your Infra Repo DRY](https://terragrunt.gruntwork.io/docs/features/keep-your-terragrunt-architecture-dry/) .

## USAGE

### Total Classic

Consider this multi-service / multi-environment monorepo:

```
/org
├── common.tfvars
├── service1
│   ├── service1-common.tfvars
│   ├── dev
│   │   ├── main.tf
│   │   └── dev.tfvars
│   └── prod
│       ├── main.tf
│       └── prod.tfvars
└── service2
    ├── service2-common.tfvars
    ├── dev
    │   ├── main.tf
    │   └── dev.tfvars
    └── prod
        ├── main.tf
        └── prod.tfvars
```

To apply Terraform for the dev environment of service1 (and keep your sanity):

```bash
cd /org/service1/dev
terraform apply $(findtfvars | tfvars-to-args)
```

This will apply variables in this order, with conflicts being resolved by favoring the most recent assignment:
```
/org/common.tfvars
/org/service1/service1-common.tfvars
/org/service1/dev/dev.tfvars
```
### Examples

#### Example 1

Default use-case, Terraform


```bash
TF_ARGS=$(findtfvars | tfvars-to-args)
terraform plan ${TF_ARGS}
```
#### Example 2
With Custom format

```
TF_ARGS=$(findtfvars -f ".toml" | tfvars-to-args -format="--file=%s")
some_other_tool ${TF_ARGS}
```

#### Example 3
With verbose output

```bash
TF_ARGS=$(findtfvars | tfvars-to-args -verbose)
echo "Arguments: ${TF_ARGS}"
terraform plan ${TF_ARGS}
```

### Usage Signature

```
findtfvars [OPTIONS]
  OPTIONS:
    -d, --dir DIR   Start searching from DIR (default: current directory)
    -h, --help      Show this help message

tfvars-to-args [OPTIONS]
  OPTIONS:
    -r, --reverse   Output args in the order they were found (default: reverse order)
    -h, --help      Show this help message
```

Environment variables:
- `FINDTFVARS_MAX_DEPTH`: Maximum directory depth to search (default: unlimited)
- `TFVARS_TO_ARGS_PREFIX`: Prefix for generated arguments (default: "-var-file=")

Config file: `~/.tfvars-toolsrc`
```
max_depth=10
arg_prefix="--var-file="
```

## INSTALL

For those who prefer their tools like their coffee - instant and strong:

```bash
# For Linux
curl -Lo tfvars-tools.tar.gz https://github.com/yourusername/tfvars-tools/releases/latest/download/tfvars-tools_Linux_x86_64.tar.gz

# For macOS
curl -Lo tfvars-tools.tar.gz https://github.com/yourusername/tfvars-tools/releases/latest/download/tfvars-tools_Darwin_x86_64.tar.gz

# Extract, make executable, and move to a directory in your PATH
tar xzf tfvars-tools.tar.gz
chmod +x findtfvars tfvars-to-args
sudo mv findtfvars tfvars-to-args /usr/local/bin/
```

Now go forth and DRY up your tfvars.

## DEVELOP

So you like recursively defined functions, you say...

1. Clone the repo
2. Read the Makefile
3. `make` your way to glory
4. Open PR

## SEE ALSO

Many of the functions in use by Terragrunt or similar tools can be achieved with a wrapper around the `tfvars-tools`.

[Keep your Infra Repo DRY](https://terragrunt.gruntwork.io/docs/features/keep-your-terragrunt-architecture-dry/)

[Google Internal best practices for Root Config Repos](https://cloud.google.com/docs/terraform/best-practices/root-modules)

