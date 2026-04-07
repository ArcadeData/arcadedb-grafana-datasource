package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testArcadeDB holds the shared ArcadeDB container for all tests in this package.
var testArcadeDB *TestArcadeDB

const (
	arcadeDBImage    = "arcadedata/arcadedb:latest"
	arcadeDBPort     = "2480/tcp"
	arcadeDBUser     = "root"
	arcadeDBPassword = "playwithdata"
)

// TestArcadeDB wraps a testcontainers ArcadeDB instance.
type TestArcadeDB struct {
	container testcontainers.Container
	baseURL   string
}

// TestMain starts a single ArcadeDB container for all tests in the package.
func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        arcadeDBImage,
		ExposedPorts: []string{arcadeDBPort},
		Env: map[string]string{
			"JAVA_OPTS":                        "-Darcadedb.server.rootPassword=" + arcadeDBPassword,
			"arcadedb.server.defaultDatabases": "",
		},
		WaitingFor: wait.ForHTTP("/api/v1/ready").
			WithPort("2480/tcp").
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK || status == http.StatusNoContent
			}).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start ArcadeDB container: %v\n", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get container host: %v\n", err)
		os.Exit(1)
	}

	mappedPort, err := container.MappedPort(ctx, "2480")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get mapped port: %v\n", err)
		os.Exit(1)
	}

	testArcadeDB = &TestArcadeDB{
		container: container,
		baseURL:   fmt.Sprintf("http://%s:%s", host, mappedPort.Port()),
	}

	code := m.Run()

	_ = container.Terminate(ctx)
	os.Exit(code)
}

// BaseURL returns the base URL for the running ArcadeDB instance.
func (t *TestArcadeDB) BaseURL() string {
	return t.baseURL
}

// CreateDatabase creates a new database in the test ArcadeDB instance.
func (t *TestArcadeDB) CreateDatabase(name string) error {
	url := fmt.Sprintf("%s/api/v1/server", t.baseURL)
	body := fmt.Sprintf(`{"command":"create database %s"}`, name)
	return t.serverCommand(url, body)
}

// DropDatabase drops a database from the test ArcadeDB instance.
func (t *TestArcadeDB) DropDatabase(name string) error {
	url := fmt.Sprintf("%s/api/v1/server", t.baseURL)
	body := fmt.Sprintf(`{"command":"drop database %s"}`, name)
	return t.serverCommand(url, body)
}

// ExecuteCommand executes a command against a database and returns the raw response body.
func (t *TestArcadeDB) ExecuteCommand(database, language, command string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/command/%s", t.baseURL, database)
	payload := map[string]interface{}{
		"language": language,
		"command":  command,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(arcadeDBUser, arcadeDBPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("command failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// NewTestClient creates an ArcadeDBClient pointed at the test instance with the given database.
func (t *TestArcadeDB) NewTestClient(database string) *ArcadeDBClient {
	return NewArcadeDBClient(t.baseURL, database, arcadeDBUser, arcadeDBPassword)
}

func (t *TestArcadeDB) serverCommand(url, body string) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))
	if err != nil {
		return err
	}
	req.SetBasicAuth(arcadeDBUser, arcadeDBPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server command failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// uniqueDBName generates a unique database name for test isolation.
func uniqueDBName(testName string) string {
	return fmt.Sprintf("test_%s_%d", testName, time.Now().UnixNano())
}
