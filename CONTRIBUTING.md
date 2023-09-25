# Contributing

By participating in this project, you agree to abide our
[code of conduct](https://github.com/TMUniversal/papercrypt/blob/main/CODE_OF_CONDUCT.md).

## Set up your machine

`papercrypt` is written in [Go](https://golang.org/).

Prerequisites:

- [Task](https://taskfile.dev/installation)
- [Go 1.21+](https://go.dev/doc/install)

Other things you might need to run the tests:

- [cosign](https://github.com/sigstore/cosign)
- [Docker](https://www.docker.com/)
- [Syft](https://github.com/anchore/syft)
- [upx](https://upx.github.io/)
- [pdfcpu](https://github.com/pdfcpu/pdfcpu)

> On Windows, installed the packages below in [WSL](https://docs.microsoft.com/en-us/windows/wsl/install-win10).

Relevant System Packages:

- `poppler-utils`: must be installed for `pdftoppm` to be available
- `upx-ucl`: must be installed for `upx` to be available

Clone `papercrypt` anywhere:

```sh
git clone git@github.com:TMUniversal/papercrypt.git
```

`cd` into the directory and install the dependencies:

```sh
task setup
```

A good way of making sure everything is all right is running the test suite:

```sh
task test
```

## Test your change

You can create a branch for your changes and try to build from the source as you go:

```sh
task build
```

When you are satisfied with the changes, we suggest you run:

```sh
task ci
```

Before you commit the changes, we also suggest you run:

```sh
task fmt
```

## Create a commit

Commit messages should be well formatted, and to make that "standardized", we
are using Conventional Commits.

You can follow the documentation on
[their website](https://www.conventionalcommits.org).

## Submit a pull request

Push your branch to your `papercrypt` fork and open a pull request against the main branch.
