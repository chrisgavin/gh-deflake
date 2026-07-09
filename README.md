# gh-deflake
A [GitHub CLI extension](https://cli.github.com/) for rerunning flaky CI until it passes.

deflake will iterate the check suites/runs on the pr, find failing ones, and re-run them.
E.g.

```
$ gh deflake https://github.com/myorg/myrepo/pull/42
INFO[0004] Rerun triggered for actions workflow run https://github.com/myorg/myrepo/actions/runs/1234/job/12345 (123456). 
INFO[0004] Rerun triggered for actions workflow run https://github.com/myorg/myrepo/actions/runs/5678/job/56789 (5678910). 
INFO[0010] All check suites are now green.
```

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
