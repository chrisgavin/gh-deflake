package client

import (
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/pkg/errors"
)

func NewClient(host string) (api.RESTClient, error) {
	client, err := gh.RESTClient(&api.ClientOptions{Host: host})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create GitHub client.")
	}
	return client, nil
}
