# gh-deflake
A [GitHub CLI extension](https://cli.github.com/) for rerunning flaky CI until it passes.

deflake will iterate the check suites/runs on the pr, find failing ones, and re-run them.

## Installation

```sh
gh extension install https://github.com/chrisgavin/gh-deflake/
```

## Upgrading

```sh
gh extension upgrade gh-deflake
```

## Usage

```sh
gh deflake <pull request> [flags]
```
