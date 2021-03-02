package tests

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSchema = `CREATE TABLE Persons (
    id INT PRIMARY KEY,
    Name VARCHAR(255),
    Email TEXT,
    City VARCHAR(255)
);`

const rowsCount = 100000

type Log interface {
	Logf(format string, args ...interface{})
}

func PrepareTestDatabase(log Log) (*sql.DB, string, func(), error) {
	connectionStr, cancel, err := setupPostgres(log)
	if err != nil {
		return nil, "", nil, err
	}

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		cancel()
		return nil, "", nil, err
	}

	err = seedDatabase(db)
	if err != nil {
		cancel()
		return nil, "", nil, err
	}

	return db, connectionStr, cancel, nil
}

func setupPostgres(log Log) (string, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return "", nil, err
	}

	postgres, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return "", nil, err
	}

	cancel := func() {
		err := pool.Purge(postgres)
		if err != nil {
			log.Logf("Error while purging database container: %s", err.Error())
		}
	}

	connectionString := fmt.Sprintf("postgres://postgres:password@localhost:%s/postgres?sslmode=disable", postgres.GetPort("5432/tcp"))

	err = pool.Retry(func() error {
		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Logf("Error while opening db connection: %s", err.Error())
			return err
		}
		err = db.Ping()
		if err != nil {
			log.Logf("Error while pinging database: %s", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		cancel()
		return "", nil, err
	}

	return connectionString, cancel, nil
}

// TODO: fake db can be enhanced with more tables and other db objects (indexes, roles etc)
func seedDatabase(db *sql.DB) error {
	// Setup schema
	_, err := db.Exec(testSchema)
	if err != nil {
		return err
	}

	// Insert reference record
	builder := strings.Builder{}
	_, err = builder.WriteString("INSERT INTO Persons (id, name, email, city) Values (1, 'Test Person 1', 'test@person.com', 'Test city')")
	if err != nil {
		return err
	}

	// Insert more fake records
	for i := 2; i <= rowsCount; i++ {
		_, err = builder.WriteString(fmt.Sprintf(", (%d, '%s', '%s', '%s')",
			i, gofakeit.Name(), gofakeit.Email(), gofakeit.City(),
		))
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(fmt.Sprintf("%s;", builder.String()))
	if err != nil {
		return err
	}

	return nil
}

func AssertDBPopulated(t *testing.T, db *sql.DB) {
	exists, err := TestPersonExists(db)
	require.NoError(t, err)
	assert.True(t, exists)

	rCount, err := RowsCount(db)
	require.NoError(t, err)
	assert.Equal(t, rowsCount, rCount)
}

type Person struct {
	Id    int
	Name  string
	Email string
	City  string
}

func TestPersonExists(db *sql.DB) (bool, error) {
	rows, err := db.Query("SELECT * FROM Persons WHERE id=1;")
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	hasNext := rows.Next()
	if !hasNext {
		return false, nil
	}

	person := Person{}

	err = rows.Scan(&person.Id, &person.Name, &person.Email, &person.City)
	if err != nil {
		return false, err
	}

	if person.Id != 0 || person.Email != "test@person.com" {
		return true, nil
	}
	return false, nil
}

func RowsCount(db *sql.DB) (int, error) {
	rows, err := db.Query("SELECT * FROM Persons;")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}

func CleanTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM Persons;")
	if err != nil {
		return err
	}
	return nil
}
