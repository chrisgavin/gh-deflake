package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/chrisgavin/gh-deflake/internal/actions"
	"github.com/chrisgavin/gh-deflake/internal/check_runs"
	"github.com/chrisgavin/gh-deflake/internal/check_suites"
	"github.com/chrisgavin/gh-deflake/internal/client"
	"github.com/chrisgavin/gh-deflake/internal/pull_request"
	"github.com/chrisgavin/gh-deflake/internal/version"
	"github.com/chrisgavin/paginated-go-gh/pkg/paginated"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type rootFlagFields struct {
}

var rootFlags = rootFlagFields{}

var SilentErr = errors.New("SilentErr")

var pullRequestRegex = regexp.MustCompile(`^(https?://[^/]+/[a-zA-Z0-9-]+/[a-zA-Z0-9-]+)/pull/([0-9]+)$`)

var rootCmd = &cobra.Command{
	Use:           "gh dispatch <workflow>",
	Short:         "A GitHub CLI extension that makes it easy to dispatch GitHub Actions workflows.",
	Version:       fmt.Sprintf("%s (%s)", version.Version(), version.Commit()),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("A PR must be specified.")
		}
		pullRequestURL := args[0]
		match := pullRequestRegex.FindStringSubmatch(pullRequestURL)

		repository, err := repository.Parse(match[1])
		if err != nil {
			return err
		}

		pullRequestNumber, err := strconv.Atoi(match[2])
		if err != nil {
			panic(err)
		}

		ghClient, err := client.NewClient(repository.Host())
		if err != nil {
			return err
		}
		ghClient = paginated.WrapClient(ghClient)

		pullRequest, err := pull_request.GetPullRequest(ghClient, repository, pullRequestNumber)
		if err != nil {
			return err
		}

		for {
			allSuitesGreen := true

			checkSuites, err := check_suites.GetCheckSuites(ghClient, repository, pullRequest.Head.Ref)
			if err != nil {
				return err
			}

			for _, checkSuite := range checkSuites.CheckSuites {
				// A check suite is created for every installed GitHub App, but some don't run any checks so these remain permanently queued.
				if checkSuite.LatestCheckRunsCount == 0 {
					continue
				}
				if !checkSuite.IsCompleted() || !checkSuite.IsSuccessful() {
					allSuitesGreen = false
				}
				if checkSuite.IsCompleted() && !checkSuite.IsSuccessful() && !checkSuite.RunsRerequestable && checkSuite.App.Slug != actions.ActionsAppSlug {
					fmt.Printf("Check suite is not rerunnable: %d\n", checkSuite.ID)
					continue
				}
				checkRuns, err := check_runs.GetCheckRuns(ghClient, repository, checkSuite.ID)
				if err != nil {
					return err
				}
				failedCheckRuns := check_runs.FilterFailedCheckRuns(checkRuns)
				if len(failedCheckRuns) > 0 {
					allSuitesGreen = false
				}
				for _, checkRun := range failedCheckRuns {
					if checkSuite.App.Slug != actions.ActionsAppSlug {
						err := check_runs.RerequestCheckRun(ghClient, repository, checkRun.ID)
						if err != nil {
							return err
						}
					} else {
						actionsRunID, err := actions.ExtractActionsRunIDFromURL(checkRun.HTMLURL)
						if err != nil {
							return err
						}
						err = actions.RerunActionsWorkflow(ghClient, repository, actionsRunID)
						if err != nil {
							return err
						}
					}
				}
			}

			if allSuitesGreen {
				break
			}

			time.Sleep(1 * time.Minute)
		}

		return nil
	},
}

func (f *rootFlagFields) Init(cmd *cobra.Command) error {
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.PrintErrln(err)
		cmd.PrintErrln()
		cmd.PrintErr(cmd.UsageString())
		return SilentErr
	})

	return nil
}

func Execute(ctx context.Context) error {
	err := rootFlags.Init(rootCmd)
	if err != nil {
		return err
	}

	return rootCmd.ExecuteContext(ctx)
}
