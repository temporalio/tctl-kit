**:warning: Deprecation Notice :warning:**

The `tctl` CLI is now deprecated in favor of Temporal CLI. <br />
This repository is no longer maintained. <br />
Please use the new utility for all future development. <br />

* New [Temporal CLI repository](https://github.com/temporalio/cli).
* [Temporal CLI Documentation site](https://docs.temporal.io/cli).

# tctl-kit

tctl-kit contains a set of opinionated tooling for urfave/cli/v2 based CLIs

## Features:
* pagination of data based on `less`, `more` and other pagers. Pager can be switched with $PAGER env variable.
* limiting number of items in output (`--limit 10`)
* formatting output as Table/JSON/Card (`--output table/json/card`)
* datetime formatting (`--time-format relative`)
* color (`--color auto`)
* .yml based configuration of CLI. Supports configuring multiple environments.
* configuration of aliases for commands

### Usage
Usage examples can be found here https://github.com/temporalio/tctl.
