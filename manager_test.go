package dbdog_test

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bool64/dbdog"
	"github.com/bool64/sqluct"
	"github.com/cucumber/godog"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func mustParseTime(value string) time.Time {
	var (
		t   time.Time
		err error
	)

	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02", "2006-01-02 15:04:05"} {
		t, err = time.Parse(layout, value)
		if err == nil {
			break
		}
	}

	if err != nil {
		panic(err)
	}

	return t
}

func TestManager_RegisterContext(t *testing.T) {
	type RowKey struct {
		Foo *string        `db:"foo"`
		Bar sql.NullString `db:"bar"`
	}

	type row struct {
		ID int `db:"id"`
		RowKey
		CreatedAt time.Time  `db:"created_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	dbm := dbdog.NewManager()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dbm.Instances = map[string]dbdog.Instance{
		"my_db": {
			Storage: sqluct.NewStorage(sqlx.NewDb(db, "sqlmock")),
			Tables: map[string]interface{}{
				"my_table":         new(row),
				"my_another_table": new(row),
			},
		},
	}

	//    Given there are no rows in table "my_table" of database "my_db"
	mock.ExpectExec(`DELETE FROM my_table`).
		WillReturnResult(driver.ResultNoRows)

	//    And rows from this file are stored in table "my_table" of database "my_db"
	//    """
	//    _testdata/rows.csv
	//    """
	mock.ExpectExec(`INSERT INTO my_table \(id,created_at,deleted_at,foo,bar\) VALUES .+`).
		WithArgs(
			1, mustParseTime("2021-01-01T00:00:00Z"), nil, "foo-1", "abc",
			2, mustParseTime("2021-01-02T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z"), "foo-1", "def",
			3, mustParseTime("2021-01-03T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z"), "foo-2", "hij",
		).
		WillReturnResult(driver.ResultNoRows)

	//    And these rows are stored in table "my_table" of database "my_db":
	//      | id | foo   | bar | created_at           | deleted_at           |
	//      | 1  | foo-1 | abc   | 2021-01-01T00:00:00Z | NULL                 |
	mock.ExpectExec(`INSERT INTO my_table \(id,created_at,deleted_at,foo,bar\) VALUES .+`).
		WithArgs(
			1, mustParseTime("2021-01-01T00:00:00Z"), nil, "foo-1", "abc",
		).
		WillReturnResult(driver.ResultNoRows)

	//    Then only these rows are available in table "my_table" of database "my_db":
	//      | id | foo   | bar | created_at           | deleted_at           |
	//      | 1  | foo-1 | abc   | 2021-01-01T00:00:00Z | NULL                 |
	//      | 2  | foo-1 | def      | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
	//      | 3  | foo-2 | hij      | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |
	mock.ExpectQuery(`SELECT COUNT\(1\) AS c FROM my_table`).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(3))

	mock.ExpectQuery(`SELECT .+ FROM my_table WHERE bar = \$1 AND created_at = \$2 AND deleted_at IS NULL`).
		WithArgs(
			"abc", mustParseTime("2021-01-01T00:00:00Z"),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "foo", "bar", "created_at", "deleted_at"}).
			AddRow(1, "foo-1", "abc", mustParseTime("2021-01-01T00:00:00Z"), nil))

	mock.ExpectQuery(`SELECT .+ FROM my_table WHERE foo = \$1 AND bar = \$2 AND created_at = \$3 AND deleted_at = \$4`).
		WithArgs(
			"foo-1", "def", mustParseTime("2021-01-02T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z"),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "foo", "bar", "created_at", "deleted_at"}).
			AddRow(2, "foo-1", "def", mustParseTime("2021-01-02T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z")))

	mock.ExpectQuery(`SELECT .+ FROM my_table WHERE foo = \$1 AND bar = \$2 AND created_at = \$3 AND deleted_at = \$4`).
		WithArgs(
			"foo-2", "hij", mustParseTime("2021-01-03T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z"),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "foo", "bar", "created_at", "deleted_at"}).
			AddRow(3, "foo-2", "hij", mustParseTime("2021-01-03T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z")))

	// Assertion with interpolated variables.
	//    Then only these rows are available in table "my_table" of database "my_db":
	//      | id    | foo   | bar | created_at           | deleted_at           |
	//      | <id1> | foo-1 | abc   | 2021-01-01T00:00:00Z | NULL                 |
	//      | <id2> | foo-1 | def      | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
	//      | <id3> | foo-2 | hij      | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |
	mock.ExpectQuery(`SELECT COUNT\(1\) AS c FROM my_table`).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(3))

	mock.ExpectQuery(`SELECT .+ FROM my_table WHERE id = \$1 AND foo = \$2 AND bar = \$3 AND created_at = \$4 AND deleted_at IS NULL`).
		WithArgs(
			1, "foo-1", "abc", mustParseTime("2021-01-01T00:00:00Z"),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "foo", "bar", "created_at", "deleted_at"}).
			AddRow(1, "foo-1", "abc", mustParseTime("2021-01-01T00:00:00Z"), nil))

	mock.ExpectQuery(`SELECT .+ FROM my_table WHERE id = \$1 AND foo = \$2 AND bar = \$3 AND created_at = \$4 AND deleted_at = \$5`).
		WithArgs(
			2, "foo-1", "def", mustParseTime("2021-01-02T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z"),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "foo", "bar", "created_at", "deleted_at"}).
			AddRow(2, "foo-1", "def", mustParseTime("2021-01-02T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z")))

	mock.ExpectQuery(`SELECT .+ FROM my_table WHERE id = \$1 AND foo = \$2 AND bar = \$3 AND created_at = \$4 AND deleted_at = \$5`).
		WithArgs(
			3, "foo-2", "hij", mustParseTime("2021-01-03T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z"),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "foo", "bar", "created_at", "deleted_at"}).
			AddRow(3, "foo-2", "hij", mustParseTime("2021-01-03T00:00:00Z"), mustParseTime("2021-01-03T00:00:00Z")))

	//    And no rows are available in table "my_another_table" of database "my_db"
	mock.ExpectQuery(`SELECT COUNT\(1\) AS c FROM my_another_table`).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(0))

	buf := bytes.NewBuffer(nil)

	suite := godog.TestSuite{
		Name:                 "DatabaseContext",
		TestSuiteInitializer: nil,
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			dbm.RegisterSteps(s)
		},
		Options: &godog.Options{
			Format:    "pretty",
			Output:    buf,
			Paths:     []string{"Database.feature"},
			Strict:    true,
			Randomize: time.Now().UTC().UnixNano(),
		},
	}
	status := suite.Run()

	if status != 0 {
		t.Fatal(buf.String())
	}
}

func TestManager_RegisterContext_fail(t *testing.T) {
	type RowKey struct {
		Foo string         `db:"foo"`
		Bar sql.NullString `db:"bar"`
	}

	type row struct {
		ID int `db:"id"`
		RowKey
		CreatedAt time.Time  `db:"created_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	dbm := dbdog.NewManager()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dbm.Instances = map[string]dbdog.Instance{
		"my_db": {
			Storage: sqluct.NewStorage(sqlx.NewDb(db, "sqlmock")),
			Tables: map[string]interface{}{
				"my_table": new(row),
			},
		},
	}

	mock.ExpectQuery(`SELECT COUNT\(1\) AS c FROM my_table`).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))

	createdAt := time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC)

	mock.ExpectQuery(`SELECT id, created_at, deleted_at, foo, bar FROM my_table LIMIT 50`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "deleted_at", "foo", "bar"}).
			AddRow(1, createdAt, nil, "my-foo", "bar-1").
			AddRow(2, createdAt, nil, "my-foo", "bar-122"))

	buf := bytes.NewBuffer(nil)

	suite := godog.TestSuite{
		Name:                 "DatabaseContext",
		TestSuiteInitializer: nil,
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			dbm.RegisterSteps(s)
		},
		Options: &godog.Options{
			Format: "pretty",
			Output: buf,
			Paths:  []string{"DatabaseFail.feature"},
			Strict: true,
		},
	}
	status := suite.Run()

	assert.Contains(t, buf.String(), `
| id | foo    | bar     | created_at                     | deleted_at |
| 1  | my-foo | bar-1   | 2020-01-01T01:01:01.000000001Z | NULL       |
| 2  | my-foo | bar-122 | 2020-01-01T01:01:01.000000001Z | NULL       |
`)

	if status == 0 {
		t.Fatal(buf.String())
	}
}
