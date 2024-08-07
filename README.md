# derisk-sql
## :rocket: Remove unexpected risks from your SQL migrations :rocket:
derisk-sql is a extensibility-first SQL linting tool to prevent mistakes from sneaking into your SQL migration files.

This includes SQL linting rules (aka `analyzer`s) like:
- requiring keywords like `CONCURRENTLY` for `INDEX` operations to improve performance
- re-organizing table definition statements to optimize table storage usage
- requiring specific reviewers on pull requests for high throughput / sufficiently large tables
- enforcing naming conventions
- etc.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Installation](#installation)
- [Usage](#usage)
  - [Picking analyzers](#picking-analyzers)
  - [Config files](#config-files)
- [Extensibility](#extensibility)
  - [Examples](#examples)
  - [Demo: extending a custom analyzer](#demo-extending-a-custom-analyzer)
    - [Sample input/output](#sample-inputoutput)
    - [Analyzer: warn.sh](#analyzer-warnsh)
    - [Analyzer: forbid-drop-table.sh](#analyzer-forbid-drop-tablesh)
  - [Ta-da!](#ta-da)
- [Limitations](#limitations)
- [Github Workflow](#github-workflow)
- [Feature requests](#feature-requests)
- [Collaboration](#collaboration)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Installation
```
$ go install github.com/aprimetechnology/derisk-sql/...
```

# Usage
```
# --migrations-dir can be set explicitly, and defaults to ‘./migrations’
$ derisk-sql check run
```

## Picking analyzers
By default, all analyzers (defined in [./analyzers](./analyzers)) are run.

To specify a subset, or your own, or a mix of both, provide the paths to all those analyzers like so:
```
$ derisk-sql check run --analyzers ./my-binary /home/user/some-other-binary ...
```
## Config files
Alternatively, a config file can be specified in the current directory for all CLI options.

The config file must be named `settings`, with any file extension (`.json`, `.yaml`, `.toml`, etc) supported by [viper](https://github.com/spf13/viper/blob/v1.19.0/viper.go#L422).

# Extensibility
Want to extend the tool with your own custom functionality?

This tool was designed with end-user extensibility as a first-class concept.

## Examples
Next, this README will step through some examples in [./examples/extensibility](./examples/extensibility)

## Demo: extending a custom analyzer
Every SQL linting rule (aka **analyzer**) is implemented as:
- a subprocess that the tool spawns
- that receives a JSON blob to its process stdin
- that produces a JSON blob to its process stdout

That means you can extend this tool with **any language, library, binary, etc**!!

### Sample input/output
Here's what some sample input JSON and sample output JSON look like:

![](./examples/gifs/input-output.gif)

### Analyzer: warn.sh
Here follows an example of a dummy bash script analyzer that always outputs a warning.

![](./examples/gifs/warning-sh.gif)

### Analyzer: forbid-drop-table.sh
Let's see another bash script example, but that does something more meaningful.
Ie, a script that just greps for the string `DROP TABLE`

![](./examples/gifs/forbid-drop-table-sh.gif)

## Ta-da!
That's it!

You can extend functionality with a shell script, with Python, with Golang, with Java, whatever you'd like.

It only has to take in JSON of the expected schema, and produce JSON of the expected schema.

# Limitations
Currently, derisk-sql only supports:
- the following migration management tools:
    - dbmate
- the following database systems:
    - postgres
- the following Version Control Systems (VCS)
    - github

# Github Workflow
Want to add this tool to your pull requests?

Add our [example workflow](./examples/workflows/derisk-sql-ci.yml) to your repo's `.github/workflows/` directory:
```
name: derisk-sql-CI
on:
  pull_request:
    branches:
    - main
jobs:
  derisk-sql:
    runs-on: ubuntu-latest
    # Sets the permissions granted to the `GITHUB_TOKEN` for the actions in this job.
    permissions:
        # permission to actions/checkout the contents of this PR branch
        contents: write
        # permission to pull the derisk-sql docker image from the GitHub Container Registry
        packages: read
        # permission to post comments on the PR
        pull-requests: write
    container:
      image: ghcr.io/aprimetechnology/derisk-sql
    steps:
      - name: Checkout the contents of this repo
        uses: actions/checkout@v4
      - name: produce derisk-sql reports
        run: derisk-sql check run
      - name: process derisk-sql reports
        if: always()
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_PULL_REQUEST_NUMBER: ${{ github.event.pull_request.number }}
          # GITHUB_REPOSITORY is <owner>/<repo>, this will be just <repo>
          GITHUB_REPOSITORY_NAME: ${{ github.event.repository.name }}
          # GITHUB_REPOSITORY_OWNER is set here automatically
        run: derisk-sql check ci
```

# Feature requests
We are very happy to take any and all feature requests!

In fact, this tool's very existence came out of a request from our end users.

We do value your input, and want to make this tool as streamlined and useful as possible.

# Collaboration
If you find yourself wanting a feature request with private support, we can help!

[APrime](https://www.aprime.com/) operates with companies of all sizes and provides flexible engagement models: ranging from flex capacity and fractional leadership to fully embedding our team at your company.

We are passionate about innovating, love solving tough problems, shipping products and code, and being able to see the tremendous impact on both our client companies and their end users.
No matter where you are in your journey, [schedule a call](https://www.aprime.com/contact/#contact-form) with our founders today to explore how we can help you achieve your goals.

[<img src="https://www.aprime.io/wp-content/uploads/2023/08/Aprime_logo@0.5x-1.png" width=225/>](https://www.aprime.com/)
