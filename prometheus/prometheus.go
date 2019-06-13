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

type APIClient struct {
	Url    *url.URL
	ConfigPath string
	ConfigTemplate string
	HTTPClient *http.Client
	Id         int
	Key        string
	logger     log.Logger
}

type PrometheusConfig struct {
	Jobs string
}

// reload prometheus
func (c *APIClient) Reload() (error,int) {
	return c.doPost(c.Url.String())
}

// do post request
func (c *APIClient) doPost(url string) (error,int) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err, 0
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		for strings.Contains(err.Error(), "connection refused") {
			level.Error(c.logger).Log("err", err.Error())
			level.Info(c.logger).Log("msg", "Perhaps Prometheus is not ready. Waiting for 8 seconds and retry again...")
			time.Sleep(8 * time.Second)
			resp, err = c.HTTPClient.Do(req)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return err, 0
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code returned from Prometheus (got: %d, expected: 200, msg:%s)", resp.StatusCode, response), resp.StatusCode
	}
	return nil, 0
}

// return a new APIClient
func New(baseUrl *url.URL, configPath string, configTemplate string, id int, key string, logger log.Logger) *APIClient {
	return &APIClient{
		Url:    baseUrl,
		ConfigPath: configPath,
		ConfigTemplate : configTemplate,
		HTTPClient: http.DefaultClient,
		Id:         id,
		Key:        key,
		logger:     logger,
	}
}