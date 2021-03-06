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

var restoreCmd = &cobra.Command{
	Use:          "restore",
	Short:        "Run database restoration",
	SilenceUsage: true,
	RunE: func(command *cobra.Command, args []string) error {

		opts := ConfigFromOptions()

		restoreOpts := backuprestore.RestoreOptions{
			Config:              opts,
			PreserveRestoreFile: viper.GetBool("preserve"),
		}

		return runRestore(restoreOpts)
	},
}

func runRestore(opts backuprestore.RestoreOptions) error {
	err := opts.Validate()
	if err != nil {
		return errors.Wrap(err, "validation failed")
	}

	operator, err := backuprestore.NewDBOperator(opts.ConnectionString, viper.GetString("log-file"))
	if err != nil {
		return errors.Wrap(err, "failed to create DB operator")
	}

	downloader, err := storage.NewS3FileBackend(opts.StorageConfig)
	if err != nil {
		return errors.Wrap(err, "failed to prepare file downloader")
	}

	err = backuprestore.Restore(operator, downloader, opts, logrus.New())
	if err != nil {
		return errors.Wrap(err, "failed to perform restoration")
	}

	return nil
}
