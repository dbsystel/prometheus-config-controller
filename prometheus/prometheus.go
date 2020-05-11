package prometheus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// APIClient to reload Prometheus
type APIClient struct {
	URL            *url.URL
	ConfigPath     string
	ConfigTemplate string
	HTTPClient     *http.Client
	ID             string
	Key            string
	logger         log.Logger
}

// Config for Prometheus
type Config struct {
	Jobs string
}

// Reload prometheus
func (c *APIClient) Reload() (int, error) {
	return c.doPost(c.URL.String())
}

// do post request
func (c *APIClient) doPost(url string) (int, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		for strings.Contains(err.Error(), "connection refused") {
			//nolint:errcheck
			level.Error(c.logger).Log("err", err.Error())
			//nolint:errcheck
			level.Info(c.logger).Log("msg", "Perhaps Prometheus is not ready. Waiting for 8 seconds and retry again...")
			time.Sleep(8 * time.Second)
			resp, err = c.HTTPClient.Do(req)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		//nolint:lll
		return resp.StatusCode, fmt.Errorf("unexpected status code returned from Prometheus (got: %d, expected: 200, msg:%s)", resp.StatusCode, response)
	}
	return 0, nil
}

// New creates a new APIClient
func New(baseURL *url.URL, configPath string, configTemplate string, id string, key string, logger log.Logger) *APIClient {
	return &APIClient{
		URL:            baseURL,
		ConfigPath:     configPath,
		ConfigTemplate: configTemplate,
		HTTPClient:     http.DefaultClient,
		ID:             id,
		Key:            key,
		logger:         logger,
	}
}
