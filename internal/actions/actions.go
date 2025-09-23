package actions

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/pkg/errors"
)

const ActionsAppSlug = "github-actions"

var actionsRunURLRegex = regexp.MustCompile(`^.*/actions?/runs/(\d+)/jobs?/(\d+)$`)

type App struct {
	Slug string `json:"slug"`
}

func ExtractActionsRunIDFromURL(url string) (int64, error) {
	matches := actionsRunURLRegex.FindStringSubmatch(url)
	if matches == nil {
		return 0, errors.New("Unable to extract actions run ID from URL " + url + ".")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

func RerunActionsWorkflow(client api.RESTClient, repository repository.Repository, workflowRunID int64) error {
	if err := client.Post(fmt.Sprintf("repos/%s/%s/actions/runs/%d/rerun-failed-jobs", repository.Owner(), repository.Name(), workflowRunID), nil, nil); err != nil {
		if httpError, ok := err.(api.HTTPError); ok {
			if httpError.StatusCode == 403 && httpError.Message == "This workflow is already running" {
				return nil
			}
		}
		return errors.Wrap(err, "Unable to rerun actions workflow.")
	}
	return nil
}
