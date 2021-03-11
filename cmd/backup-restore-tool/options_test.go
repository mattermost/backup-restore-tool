package main

import (
	"testing"

	"github.com/mattermost/backup-restore-tool/pkg/backuprestore"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfigFromOptions(t *testing.T) {
	expectedConfig := backuprestore.Config{
		DatabaseConfig: backuprestore.DatabaseConfig{
			ConnectionString: "postgres://db",
		},
		StorageConfig: backuprestore.StorageConfig{
			Endpoint:   "bifrost:80",
			Bucket:     "my-bucket",
			Region:     "east",
			ObjectKey:  "backup-123",
			PathPrefix: "my-backups",
			AccessKey:  "abcd-key",
			SecretKey:  "secret",
			EnableTLS:  true,
			Bifrost:    true,
		},
	}

	viper.Set("database", "postgres://db")
	viper.Set("storage-endpoint", "bifrost:80")
	viper.Set("storage-bucket", "my-bucket")
	viper.Set("storage-region", "east")
	viper.Set("storage-path-prefix", "my-backups")
	viper.Set("storage-object-key", "backup-123")
	viper.Set("storage-access-key", "abcd-key")
	viper.Set("storage-secret-key", "secret")
	viper.Set("storage-tls", "true")
	viper.Set("storage-type", "bifrost")

	options := ConfigFromOptions()

	require.Equal(t, expectedConfig, options)
}
