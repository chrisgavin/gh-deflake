package client

import (
	"net/http"

	"github.com/chrisgavin/paginated-go-gh/v2/pkg/paginated"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/pkg/errors"
)

func NewClient(host string) (*api.RESTClient, error) {
	client, err := api.NewRESTClient(api.ClientOptions{Host: host, Transport: paginated.NewRoundTripper(http.DefaultTransport)})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create GitHub client.")
	}
	return client, nil
}
