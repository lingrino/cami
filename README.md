# Cami

Cami is an API and CLI for removing unused AMIs from your AWS account.

## Usage

TODO - docs

## Limitations

Cami works by describing all of the AMIs in your account and all of your EC2 instances. It then creates a list of AMIs you own that have no associated EC2 instances and deletes those AMIs and the snapshots backing them. Do not use cami in the following situations:

- If you share AMIs with other accounts, cami will delete these anyway
- If you use non-EC2 services that depend on AMIs, cami will try to delete these as well
- If you have AMIs that are not running instances but will in the future, these will also be deleted.

## Contributing

I welcome issues and pull reuqests of all sizes! Especially those that resolve or mitigate the limitations listed above. Please open an issue if you're unsure that your change will be welcome.
