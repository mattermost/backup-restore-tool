package main

import (
	"github.com/mattermost/backup-restore-tool/pkg/backuprestore"
	"github.com/mattermost/backup-restore-tool/pkg/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {}

var backupCmd = &cobra.Command{
	Use:          "backup",
	Short:        "Run database backup",
	SilenceUsage: true,
	RunE: func(command *cobra.Command, args []string) error {

		opts := ConfigFromOptions()

		backupOpts := backuprestore.BackupOptions{
			Config:             opts,
			PreserveBackupFile: viper.GetBool("preserve"),
		}

		return runBackup(backupOpts)
	},
}

func runBackup(opts backuprestore.BackupOptions) error {
	err := opts.Validate()
	if err != nil {
		return errors.Wrap(err, "validation failed")
	}

	operator, err := backuprestore.NewDBOperator(opts.ConnectionString, viper.GetString("log-file"))
	if err != nil {
		return errors.Wrap(err, "failed to create DB operator")
	}

	uploader, err := storage.NewS3FileBackend(opts.StorageConfig)
	if err != nil {
		return errors.Wrap(err, "failed to prepare file uploader")
	}

	err = backuprestore.Backup(operator, uploader, opts, logrus.New())
	if err != nil {
		return errors.Wrap(err, "failed to perform backup")
	}

	return nil
}
