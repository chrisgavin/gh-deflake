package pull_request

import (
	"fmt"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/pkg/errors"
)

type PullRequestEnd struct {
	Ref string `json:"ref"`
}

type PullRequest struct {
	Head *PullRequestEnd `json:"head"`
}

func GetPullRequest(client api.RESTClient, repository repository.Repository, pullRequestNumber int) (*PullRequest, error) {
	pullRequest := PullRequest{}
	if err := client.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", repository.Owner(), repository.Name(), pullRequestNumber), &pullRequest); err != nil {
		return nil, errors.Wrap(err, "Unable to get pull request.")
	}
	return &pullRequest, nil
}
