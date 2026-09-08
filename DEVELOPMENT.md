# Provider Development

## Recommended: Dev Container

The repository includes a [Development Container](https://containers.dev/) configuration with Go, Terraform, the Azure CLI, and the recommended VS Code extensions. Using it is the easiest way to get a consistent development environment.

Install [Docker](https://docs.docker.com/get-docker/), [Visual Studio Code](https://code.visualstudio.com/), and the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers). Then open the repository in VS Code and run **Dev Containers: Reopen in Container** from the Command Palette.

After the container is created, continue with [Set Up the Repository](#set-up-the-repository).

## Manual Prerequisites

- [Go](https://go.dev/doc/install) 1.25 or later
- [Terraform](https://developer.hashicorp.com/terraform/install) 1.8 or later
- An Azure account and the [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) for tests and data updates that query Azure

## Set Up the Repository

Download the Go dependencies and build the provider from the repository root:

```shell
go mod download
go build ./...
```

## Run Tests

Run the Go test suite with:

```shell
go test ./...
```

The acceptance tests invoke Terraform and are enabled by `TF_ACC`. Run the full suite with:

```shell
make testacc
```

To run a specific test, pass Go test arguments through `TESTARGS`:

```shell
make testacc TESTARGS='-run TestCustomNameFunction'
```

Some tests fetch the current Azure CAF resource definitions and therefore require network access. Tests that query Azure locations also require an authenticated Azure CLI session:

```shell
az login
```

Tests that require Azure authentication are skipped when credentials are unavailable.

## Generate Code and Documentation

Generation is configured in `main.go`. From the repository root, run:

```shell
go generate
```

This command:

- formats the Terraform examples;
- regenerates the Azure models under `internal/cloud/azure` from the checked-in data files; and
- regenerates the provider documentation under `docs` using `templates` and `examples`.

Edit templates, examples, or schema descriptions rather than editing generated documentation when one of those files owns the content. Commit generated changes with their source changes.

## Update Azure Source Data

The Azure resource and location definitions are checked into `tools/azure/data`. Refresh them only when updating the bundled data:

```shell
./download_resources.sh
./download_locations.sh
go generate
```

`download_resources.sh` downloads the latest Azure CAF resource definitions and requires network access. `download_locations.sh` queries Azure through the CLI and requires `az login`.

Review generated model changes carefully before committing them.

## Before Opening a Pull Request

Run the checks relevant to your change:

```shell
go fmt ./...
go generate
go test ./...
go build ./...
```

Run `make testacc` when changing provider behavior, schemas, data sources, or functions. Include generated documentation and model updates when their sources change.