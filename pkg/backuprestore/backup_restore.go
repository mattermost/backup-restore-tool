package backuprestore

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type Storage interface {
	UploadFile(objectKey, path string) error
	DownloadFile(objectKey, path string) error
}

type BackupOptions struct {
	Config
	PreserveBackupFile bool
}

type RestoreOptions struct {
	Config
	PreserveRestoreFile bool
}

func Backup(dbOperator DBOperator, fileStorage Storage, options BackupOptions, log logrus.FieldLogger) error {
	tempFile := path.Join(os.TempDir(), options.ObjectKey)
	if !options.PreserveBackupFile {
		defer cleanupFile(tempFile)
	}

	log.Infof("Starting database export to file: %s", tempFile)
	err := dbOperator.Export(tempFile)
	if err != nil {
		return err
	}
	log.Info("Database export finished.")

	log.Info("Starting file upload")
	err = fileStorage.UploadFile(options.ObjectKey, tempFile)
	if err != nil {
		return err
	}
	log.Info("File upload finished")

	return nil
}

func Restore(dbOperator DBOperator, fileStorage Storage, options RestoreOptions, log logrus.FieldLogger) error {
	tempFile := path.Join(os.TempDir(), options.ObjectKey)
	if !options.PreserveRestoreFile {
		defer cleanupFile(tempFile)
	}

	log.Infof("Staring backup file download to: %s", tempFile)
	err := fileStorage.DownloadFile(options.ObjectKey, tempFile)
	if err != nil {
		return err
	}
	logrus.Infof("Backup file download finished.")

	log.Infof("Staring database restoration")
	err = dbOperator.Restore(tempFile)
	if err != nil {
		return err
	}
	log.Info("Database restoration finished.")

	return nil
}

func cleanupFile(path string) {
	err := os.Remove(path)
	if err != nil {
		logrus.Errorf("Failed to remove file %q: %s", path, err.Error())
	}
}
