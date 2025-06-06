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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type rootFlagFields struct {
}

var rootFlags = rootFlagFields{}

var SilentErr = errors.New("SilentErr")

var baseRegex = regexp.MustCompile(`^(https?://[^/]+/[a-zA-Z0-9-]+/[a-zA-Z0-9-]+)/(.*)$`)
var pullRequestRegex = regexp.MustCompile(`^pull/([0-9]+)$`)
var commitRegex = regexp.MustCompile(`^commits?/([a-f0-9]+)$`)

var rootCmd = &cobra.Command{
	Use:           "gh deflake <pull request / commit URL>",
	Short:         "A GitHub CLI extension for rerunning flaky CI until it passes.",
	Version:       fmt.Sprintf("%s (%s)", version.Version(), version.Commit()),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("A URL must be specified.")
		}

		url := args[0]

		match := baseRegex.FindStringSubmatch(url)
		if match == nil {
			return errors.New("Invalid URL format.")
		}

		repository, err := repository.Parse(match[1])
		if err != nil {
			return errors.Wrap(err, "Failed to parse repository from URL")
		}

		ghClient, err := client.NewClient(repository.Host())
		if err != nil {
			return err
		}
		ghClient = paginated.WrapClient(ghClient)

		var headRefOrSha string

		pullRequestMatch := pullRequestRegex.FindStringSubmatch(match[2])
		commitMatch := commitRegex.FindStringSubmatch(match[2])
		if pullRequestMatch != nil {
			pullRequestNumber, err := strconv.Atoi(pullRequestMatch[1])
			if err != nil {
				panic(err)
			}

			pullRequest, err := pull_request.GetPullRequest(ghClient, repository, pullRequestNumber)
			if err != nil {
				return err
			}

			headRefOrSha = pullRequest.Head.Ref
		} else if commitMatch != nil {
			headRefOrSha = commitMatch[1]
		} else {
			return errors.New("Invalid URL format. Expected a pull request or commit URL.")
		}

		for {
			allSuitesGreen := true

			checkSuites, err := check_suites.GetCheckSuites(ghClient, repository, headRefOrSha)
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
					log.Warnf("Check suite is not rerunnable: %d", checkSuite.ID)
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
						log.Infof("Rerun triggered for check run %s.", checkRun.HTMLURL)
					} else {
						actionsRunID, err := actions.ExtractActionsRunIDFromURL(checkRun.HTMLURL)
						if err != nil {
							return err
						}
						err = actions.RerunActionsWorkflow(ghClient, repository, actionsRunID)
						if err != nil {
							return err
						}
						log.Infof("Rerun triggered for actions workflow run %s (%d).", checkRun.HTMLURL, actionsRunID)
					}
				}
			}

			if allSuitesGreen {
				log.Info("All check suites are now green.")
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
