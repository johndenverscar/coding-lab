package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Healthcheck endpoint suite
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
}

func (suite *HealthEndpointTestSuite) TestHealthEndpoint() {
	url := suite.baseURL + "/health"

	res, err := suite.client.Get(url)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "ok", string(body))
}

func TestHealthEndpointTestSuite(t *testing.T) {
	suite.Run(t, new(HealthEndpointTestSuite))
}

// /Replicas GET endpoint suite
type GetReplicasTestSuite struct {
	suite.Suite
	baseURL        string
	client         *http.Client
	testNamespace  string
	testDeployment string
}

func (suite *GetReplicasTestSuite) SetupSuite() {
	suite.baseURL = "http://localhost:30080" // NodePort for Qbert
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	suite.testNamespace = "integration-test"
	suite.testDeployment = "nginx"
}

func (suite *GetReplicasTestSuite) SetupTest() {
	err := createTestDeployment(suite.testNamespace, suite.testDeployment, 3)
	require.NoError(suite.T(), err)

	err = waitForDeployment(suite.testNamespace, suite.testDeployment, 60*time.Second)
	require.NoError(suite.T(), err)
}

func (suite *GetReplicasTestSuite) TearDownTest() {
	err := deleteTestDeployment(suite.testNamespace, suite.testDeployment)
	require.NoError(suite.T(), err)
}

func (suite *GetReplicasTestSuite) TearDownSuite() {
	// Clean up the namespace after all tests in this suite
	err := deleteNamespace(suite.testNamespace)
	if err != nil {
		suite.T().Logf("Warning: failed to delete namespace %s: %v", suite.testNamespace, err)
	}
}

func (suite *GetReplicasTestSuite) TestGetReplicas_Success() {
	// Is there a better way to format strings?
	url := fmt.Sprintf("%s/deployments/%s/%s/replicas", suite.baseURL, suite.testNamespace, suite.testDeployment)

	res, err := suite.client.Get(url)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), float64(3), response["replicas"])
}

func (suite *GetReplicasTestSuite) TestGetReplicas_DeploymentNotFound() {
	url := fmt.Sprintf("%s/deployments/%s/%s/replicas", suite.baseURL, suite.testNamespace, "nonexistent-deployment")

	res, err := suite.client.Get(url)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusNotFound, res.StatusCode)
}

func (suite *GetReplicasTestSuite) TestGetReplicas_NamespaceNotFound() {
	url := fmt.Sprintf("%s/deployments/%s/%s/replicas", suite.baseURL, "nonexistent-namespace", suite.testDeployment)

	res, err := suite.client.Get(url)

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusNotFound, res.StatusCode)
}

func TestGetReplicasTestSuite(t *testing.T) {
	suite.Run(t, new(GetReplicasTestSuite))
}

// Replicas PUT endpoint suite
type PutReplicasTestSuite struct {
	suite.Suite
	baseURL        string
	client         *http.Client
	testNamespace  string
	testDeployment string
}

func (suite *PutReplicasTestSuite) SetupSuite() {
	suite.baseURL = "http://localhost:30080" // NodePort for Qbert
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}
	suite.testNamespace = "integration-test"
	suite.testDeployment = "nginx"
}

func (suite *PutReplicasTestSuite) SetupTest() {
	err := createTestDeployment(suite.testNamespace, suite.testDeployment, 2)
	require.NoError(suite.T(), err)

	err = waitForDeployment(suite.testNamespace, suite.testDeployment, 60*time.Second)
	require.NoError(suite.T(), err)
}

func (suite *PutReplicasTestSuite) TearDownTest() {
	err := deleteTestDeployment(suite.testNamespace, suite.testDeployment)
	require.NoError(suite.T(), err)
}

func (suite *PutReplicasTestSuite) TearDownSuite() {
	// Clean up the namespace after all tests in this suite
	err := deleteNamespace(suite.testNamespace)
	if err != nil {
		suite.T().Logf("Warning: failed to delete namespace %s: %v", suite.testNamespace, err)
	}
}

func (suite *PutReplicasTestSuite) TestPutReplicas_ScaleUp() {
	url := fmt.Sprintf("%s/deployments/%s/%s/replicas", suite.baseURL, suite.testNamespace, suite.testDeployment)

	payload := map[string]int{"replicas": 5}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(suite.T(), err)

	req, err := http.NewRequest(http.MethodPut, url, io.NopCloser(io.Reader(bytes.NewReader(payloadBytes))))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	res, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), float64(5), response["replicas"])

	// Verify the deployment was actually scaled
	err = waitForDeploymentReplicas(suite.testNamespace, suite.testDeployment, 5, 60*time.Second)
	require.NoError(suite.T(), err)
}

func (suite *PutReplicasTestSuite) TestPutReplicas_ScaleDown() {
	url := fmt.Sprintf("%s/deployments/%s/%s/replicas", suite.baseURL, suite.testNamespace, suite.testDeployment)

	payload := map[string]int{"replicas": 1}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(suite.T(), err)

	req, err := http.NewRequest(http.MethodPut, url, io.NopCloser(io.Reader(bytes.NewReader(payloadBytes))))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	res, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), float64(1), response["replicas"])

	// Verify the deployment was actually scaled
	err = waitForDeploymentReplicas(suite.testNamespace, suite.testDeployment, 1, 60*time.Second)
	require.NoError(suite.T(), err)
}

func (suite *PutReplicasTestSuite) TestPutReplicas_ScaleZero() {
	url := fmt.Sprintf("%s/deployments/%s/%s/replicas", suite.baseURL, suite.testNamespace, suite.testDeployment)

	payload := map[string]int{"replicas": 0}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(suite.T(), err)

	req, err := http.NewRequest(http.MethodPut, url, io.NopCloser(io.Reader(bytes.NewReader(payloadBytes))))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	res, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), float64(0), response["replicas"])

	// Verify the deployment was actually scaled
	err = waitForDeploymentReplicas(suite.testNamespace, suite.testDeployment, 0, 60*time.Second)
	require.NoError(suite.T(), err)
}

func TestPutReplicasSuite(t *testing.T) {
	suite.Run(t, new(PutReplicasTestSuite))
}
