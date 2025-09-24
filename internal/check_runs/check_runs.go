package check_runs

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/pkg/errors"
)

type CheckRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
}

type CheckRuns struct {
	CheckRuns []CheckRun `json:"check_runs"`
}

func GetCheckRuns(client *api.RESTClient, repository repository.Repository, checkSuiteID int64) (CheckRuns, error) {
	checkRuns := CheckRuns{}
	if err := client.Get(fmt.Sprintf("repos/%s/%s/check-suites/%d/check-runs", repository.Owner, repository.Name, checkSuiteID), &checkRuns); err != nil {
		return CheckRuns{}, errors.Wrap(err, "Unable to get check runs.")
	}
	return checkRuns, nil
}

func RerequestCheckRun(client *api.RESTClient, repository repository.Repository, checkRunID int64) error {
	if err := client.Post(fmt.Sprintf("repos/%s/%s/check-runs/%d/rerequest", repository.Owner, repository.Name, checkRunID), nil, nil); err != nil {
		return errors.Wrap(err, "Unable to rerequest check run.")
	}
	return nil
}

func FilterFailedCheckRuns(checkRuns CheckRuns) []CheckRun {
	var failedCheckRuns []CheckRun
	for _, checkRun := range checkRuns.CheckRuns {
		if checkRun.Status == "completed" && checkRun.Conclusion != "success" && checkRun.Conclusion != "skipped" && checkRun.Conclusion != "neutral" {
			failedCheckRuns = append(failedCheckRuns, checkRun)
		}
	}
	return failedCheckRuns
}
