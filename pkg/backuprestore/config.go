package backuprestore

import (
	"fmt"
	"strings"
)

type Config struct {
	DatabaseConfig
	StorageConfig
}

type DatabaseConfig struct {
	ConnectionString string
}

type StorageConfig struct {
	Endpoint  string
	Bucket    string
	Region    string
	ObjectKey string

	AccessKey string
	SecretKey string
	EnableTLS bool

	Bifrost bool
}

func (b Config) Validate() error {
	var errMsgs []string

	if isEmpty(b.ConnectionString) {
		errMsgs = append(errMsgs, "Database connection string not provided")
	}
	if isEmpty(b.Endpoint) {
		errMsgs = append(errMsgs, "Storage endpoint not provided")
	}
	if isEmpty(b.Bucket) {
		errMsgs = append(errMsgs, "Storage bucket not provided")
	}
	if isEmpty(b.Region) {
		errMsgs = append(errMsgs, "Storage region not provided")
	}
	if isEmpty(b.ObjectKey) {
		errMsgs = append(errMsgs, "Storage key not provided")
	}
	if isEmpty(b.AccessKey) {
		errMsgs = append(errMsgs, "Storage access key id not provided")
	}
	if isEmpty(b.SecretKey) {
		errMsgs = append(errMsgs, "Storage secret key not provided")
	}

	if len(errMsgs) == 0 {
		return nil
	}

	return fmt.Errorf("invalid config: %s", strings.Join(errMsgs, "; "))
}

func isEmpty(s string) bool {
	return s == ""
}
