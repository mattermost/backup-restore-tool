//+build e2e

package tests

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test runs with real S3 bucket
// To run it, make sure to login into your AWS account
// and provide appropriate environment variables.
// You can run the test with `go test -tags=e2e ./...`

func TestBackupRestore_Postgres(t *testing.T) {
	requireEnvs(t, "BRT_STORAGE_REGION", "BRT_STORAGE_BUCKET", "BRT_STORAGE_ACCESS_KEY", "BRT_STORAGE_SECRET_KEY")

	// When using different file storage then S3 - make it insecure (for testing with Bifrost)
	endpoint := os.Getenv("BRT_STORAGE_ENDPOINT")
	if endpoint != "s3.amazon.com" {
		err := os.Setenv("BRT_STORAGE_TLS", "false")
		require.NoError(t, err)
	}

	db, connectionStr, cleanup, err := PrepareTestDatabase(t)
	require.NoError(t, err)
	defer cleanup()

	AssertDBPopulated(t, db)

	err = os.Setenv("BRT_DATABASE", connectionStr)
	require.NoError(t, err)

	err = runBackupRestoreTool(fmt.Sprintf("backup --storage-object-key %s", "backup-restore-e2e-test-key"))
	require.NoError(t, err)

	err = CleanTable(db)
	require.NoError(t, err)

	exists, err := TestPersonExists(db)
	require.NoError(t, err)
	assert.False(t, exists)

	rCount, err := RowsCount(db)
	require.NoError(t, err)
	assert.Equal(t, 0, rCount)

	err = runBackupRestoreTool(fmt.Sprintf("restore --storage-object-key %s", "backup-restore-e2e-test-key"))
	require.NoError(t, err)

	AssertDBPopulated(t, db)
}

func runBackupRestoreTool(command string) error {
	cmd := exec.Command("backup-restore-tool", strings.Split(command, " ")...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func requireEnvs(t *testing.T, env ...string) {
	errors := []error{}

	for _, e := range env {
		v, set := os.LookupEnv(e)
		if !set {
			errors = append(errors, fmt.Errorf("env %q not set", e))
			continue
		}
		if v == "" {
			errors = append(errors, fmt.Errorf("env %q is empty", e))
			continue
		}
	}

	for _, e := range errors {
		t.Errorf("missing required env: %s", e.Error())
	}
	require.Empty(t, errors)
}
