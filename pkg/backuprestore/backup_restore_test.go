package backuprestore

import (
	"os"
	"path"
	"testing"

	"github.com/mattermost/backup-restore-tool/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStorage struct {
	t              *testing.T
	expectedObject string
	filePath       string
}

func (m *mockStorage) UploadFile(objectKey, path string) error {
	m.filePath = path
	assert.Equal(m.t, m.expectedObject, objectKey)
	return nil
}

func (m mockStorage) DownloadFile(objectKey, path string) error {
	assert.Equal(m.t, m.expectedObject, objectKey)
	assert.Equal(m.t, m.filePath, path)
	return nil
}

func TestBackupRestore(t *testing.T) {
	db, connectionStr, cleanup, err := tests.PrepareTestDatabase(t)
	require.NoError(t, err)
	defer cleanup()

	tests.AssertDBPopulated(t, db)

	logFile := path.Join(os.TempDir(), "test-logs-file.log")
	defer func() {
		err := os.Remove(logFile)
		assert.NoError(t, err)
	}()

	dbOperator, err := NewDBOperator(connectionStr, logFile)
	require.NoError(t, err)

	fileStorage := &mockStorage{t: t, expectedObject: "test-object"}

	config := Config{
		DatabaseConfig: DatabaseConfig{ConnectionString: connectionStr},
		StorageConfig:  StorageConfig{ObjectKey: "test-object"},
	}

	backupOpts := BackupOptions{Config: config, PreserveBackupFile: true}

	err = Backup(dbOperator, fileStorage, backupOpts, logrus.New())
	require.NoError(t, err)

	err = tests.CleanTable(db)
	require.NoError(t, err)

	exists, err := tests.TestPersonExists(db)
	require.NoError(t, err)
	assert.False(t, exists)

	rCount, err := tests.RowsCount(db)
	require.NoError(t, err)
	assert.Equal(t, 0, rCount)

	restoreOpts := RestoreOptions{Config: config}
	err = Restore(dbOperator, fileStorage, restoreOpts, logrus.New())
	require.NoError(t, err)

	tests.AssertDBPopulated(t, db)
}
