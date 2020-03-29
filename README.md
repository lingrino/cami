# Cami

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/lingrino/cami/cami)
[![goreportcard](https://goreportcard.com/badge/github.com/lingrino/cami)](https://goreportcard.com/report/github.com/lingrino/cami)
[![Maintainability](https://api.codeclimate.com/v1/badges/9dfa18d69da6065c9e5c/maintainability)](https://codeclimate.com/github/lingrino/cami/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/9dfa18d69da6065c9e5c/test_coverage)](https://codeclimate.com/github/lingrino/cami/test_coverage)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1e6ded484d4c4df0936f6607c562b6cb)](https://www.codacy.com/manual/lingrino/cami?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=lingrino/cami&amp;utm_campaign=Badge_Grade)

Cami is an API and CLI for removing unused AMIs from your AWS account.

## Usage

For API docs see the [godoc] reference

Cami requires that you are already authenticated with AWS and has no mechanism for passing credentials or other configuration directly. Cami uses the AWS Go SDK default credential chain. The easiest way is to have your environment variables or aws profile set up correctly. Make sure you have `AWS_REGION` set or you will get an error.

```shell
$ cami --help
cami is an API and CLI for removing unused AMIs from your AWS account.

Usage:
  cami [flags]

Flags:
  -d, --dryrun   Set dryrun to true to run through the deletion without deleting any AMIs.
  -h, --help     help for cami
```

```shell
$ cami
Successfully deleted:
  ami-002d2dbacdfc0420b
  snap-0f3c81d418d295671
```

## Limitations

Cami works by describing all of the AMIs in your account and all of your EC2 instances. It then creates a list of AMIs you own that have no associated EC2 instances and deletes those AMIs and the snapshots backing them. Do not use cami in the following situations:

- If you share AMIs with other accounts, cami will delete these anyway
- If you use non-EC2 services that depend on AMIs, cami will try to delete these as well
- If you have AMIs that are not running instances but will in the future, these will also be deleted.

## Contributing

I welcome issues and pull reuqests of all sizes! Especially those that resolve or mitigate the limitations listed above. Please open an issue if you're unsure that your change will be welcome.

[godoc]: https://pkg.go.dev/github.com/lingrino/cami/cami
