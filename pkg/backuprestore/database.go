package backuprestore

import (
	"strings"

	"github.com/mattermost/backup-restore-tool/pkg/database"
	"github.com/pkg/errors"
)

type DBOperator interface {
	Export(path string) error
	Restore(path string) error
}

func NewDBOperator(connectionStr, logFile string) (DBOperator, error) {
	prefEnd := strings.Index(connectionStr, "://")
	if prefEnd == -1 {
		return nil, errors.New("failed to determine database type from connection string")
	}

	driver := connectionStr[:prefEnd]
	switch driver {
	case "postgres":
		return database.NewPostgres(connectionStr, logFile)
	}

	return nil, errors.Errorf("driver '%s' is not supported", driver)
}
