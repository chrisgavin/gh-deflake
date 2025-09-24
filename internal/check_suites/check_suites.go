package check_suites

import (
	"fmt"

	"github.com/chrisgavin/gh-deflake/internal/actions"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/pkg/errors"
)

type CheckSuite struct {
	ID                   int64       `json:"id"`
	Status               string      `json:"status"`
	Conclusion           string      `json:"conclusion"`
	RunsRerequestable    bool        `json:"runs_rerequestable"`
	App                  actions.App `json:"app"`
	LatestCheckRunsCount int         `json:"latest_check_runs_count"`
}

type CheckSuites struct {
	CheckSuites []CheckSuite `json:"check_suites"`
}

func (checkSuite *CheckSuite) IsCompleted() bool {
	return checkSuite.Status == "completed"
}

func (checkSuite *CheckSuite) IsSuccessful() bool {
	return checkSuite.Conclusion == "success" || checkSuite.Conclusion == "skipped" || checkSuite.Conclusion == "neutral"
}

func GetCheckSuites(client *api.RESTClient, repository repository.Repository, ref string) (*CheckSuites, error) {
	checkSuites := CheckSuites{}
	if err := client.Get(fmt.Sprintf("repos/%s/%s/commits/%s/check-suites", repository.Owner, repository.Name, ref), &checkSuites); err != nil {
		return nil, errors.Wrap(err, "Unable to get check suites.")
	}
	return &checkSuites, nil
}
