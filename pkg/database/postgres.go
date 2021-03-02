package database

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

type Postgres struct {
	connectionString string
	logFile          string
}

func NewPostgres(connectionStr, logFile string) (Postgres, error) {
	readyCmd := pgIsReadyCmd(connectionStr, 30)
	errorBuff := captureError(readyCmd)

	err := readyCmd.Run()
	if err != nil {
		return Postgres{}, errors.Wrapf(err, "failed to access database, error output: %s", errorBuff.String())
	}

	return Postgres{connectionString: connectionStr, logFile: logFile}, nil
}

func (p Postgres) Export(path string) error {
	dumpCmd := pgDumpCustomCmd(p.connectionString, path)
	err := p.setOutStream(dumpCmd)
	if err != nil {
		return errors.Wrap(err, "failed to set output stream")
	}

	err = dumpCmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to dump Postgres database")
	}

	return nil
}

func (p Postgres) Restore(path string) error {
	restoreCmd := pgRestoreCmd(p.connectionString, path)
	err := p.setOutStream(restoreCmd)
	if err != nil {
		return errors.Wrap(err, "failed to set output stream")
	}

	err = restoreCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to restore Postgres database")
	}

	return nil
}

func (p Postgres) setOutStream(cmd *exec.Cmd) error {
	if p.logFile == "" {
		cmd.Stderr = os.Stderr
		return nil
	}

	file, err := os.Create(p.logFile)
	if err != nil {
		return errors.Wrap(err, "failed to create log file")
	}

	cmd.Stdout = file
	cmd.Stderr = file

	return nil
}

func captureError(cmd *exec.Cmd) *bytes.Buffer {
	buff := &bytes.Buffer{}
	cmd.Stderr = buff
	return buff
}

func pgIsReadyCmd(connStr string, timeoutSec int) *exec.Cmd {
	return pgIsReady(
		"-d", connStr,
		"-t", strconv.Itoa(timeoutSec),
	)
}

func pgDumpCustomCmd(connStr, dumpPath string) *exec.Cmd {
	return pgDump(
		"-d", connStr,
		"-F", "c",
		"-f", dumpPath,
	)
}

func pgRestoreCmd(connStr string, dumpPath string) *exec.Cmd {
	return pgRestore(
		"-d", connStr,
		"-F", "c",
		"--clean",
		dumpPath,
	)
}

func pgIsReady(options ...string) *exec.Cmd {
	return exec.Command("pg_isready", options...)
}

func pgDump(options ...string) *exec.Cmd {
	return exec.Command("pg_dump", options...)
}

func pgRestore(options ...string) *exec.Cmd {
	return exec.Command("pg_restore", options...)
}
