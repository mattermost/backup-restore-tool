package main

import (
	"github.com/mattermost/backup-restore-tool/pkg/backuprestore"
	"github.com/spf13/viper"
)

func ConfigFromOptions() backuprestore.Config {
	return backuprestore.Config{
		DatabaseConfig: backuprestore.DatabaseConfig{
			ConnectionString: viper.GetString("database"),
		},
		StorageConfig: backuprestore.StorageConfig{
			Endpoint:  viper.GetString("storage-endpoint"),
			Bucket:    viper.GetString("storage-bucket"),
			Region:    viper.GetString("storage-region"),
			ObjectKey: viper.GetString("storage-object-key"),

			AccessKey: viper.GetString("storage-access-key"),
			SecretKey: viper.GetString("storage-secret-key"),

			EnableTLS: viper.GetBool("storage-tls"),
			Bifrost:   viper.GetString("storage-type") == "bifrost",
		},
	}
}
