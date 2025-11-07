package test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Integration test for health check endpoint

type HealthEndpointTestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
}

func (suite *HealthEndpointTestSuite) SetupSuite() {
	suite.baseURL = "http://localhost:30080" // NodePort for Qbert
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}

	suite.waitForServiceReady()
}

func (suite *HealthEndpointTestSuite) waitForServiceReady() {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := suite.client.Get(suite.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			return
		}
		time.Sleep(1 * time.Second)
	}
	suite.T().Fatal("Service did not become ready in time")
}

func (suite *HealthEndpointTestSuite) TestHealthEndpoint() {
	url := suite.baseURL + "/health"

	res, err := suite.client.Get(url)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "OK", string(body))
}

func TestHealthEndpointTestSuite(t *testing.T) {
	suite.Run(t, new(HealthEndpointTestSuite))
}
