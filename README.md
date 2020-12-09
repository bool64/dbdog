# Cucumber database steps for Go

[![Build Status](https://github.com/bool64/dbdog/workflows/test/badge.svg)](https://github.com/bool64/dbdog/actions?query=branch%3Amaster+workflow%3Atest)
[![Coverage Status](https://codecov.io/gh/bool64/dbdog/branch/master/graph/badge.svg)](https://codecov.io/gh/bool64/dbdog)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/bool64/dbdog)
[![Time Tracker](https://wakatime.com/badge/github/bool64/dbdog.svg)](https://wakatime.com/badge/github/bool64/dbdog)
![Code lines](https://sloc.xyz/github/bool64/dbdog/?category=code)
![Comments](https://sloc.xyz/github/bool64/dbdog/?category=comments)

This module implements database-related step definitions
for [`github.com/cucumber/godog`](https://github.com/cucumber/godog).

## Database Configuration

Databases instances should be configured with `Manager.Instances`.

```go
dbm := dbdog.Manager{}

dbm.Instances = map[string]dbdog.Instance{
    "my_db": {
        Storage: storage,
        Tables: map[string]interface{}{
            "my_table":           new(repository.MyRow),
            "my_another_table":   new(repository.MyAnotherRow),
        },
    },
}
```

## Table Mapper Configuration

Table mapper allows customizing decoding string values from godog table cells into Go row structures and back.

```go
tableMapper := dbdog.NewTableMapper()

// Apply JSON decoding to a particular type.
tableMapper.Decoder.RegisterFunc(func(s string) (interface{}, error) {
    m := repository.Meta{}
    err := json.Unmarshal([]byte(s), &m)
    if err != nil {
        return nil, err
    }
    return m, err
}, repository.Meta{})

// Apply string splitting to github.com/lib/pq.StringArray.
tableMapper.Decoder.RegisterFunc(func(s string) (interface{}, error) {
    return pq.StringArray(strings.Split(s, ",")), nil
}, pq.StringArray{})

// Create database manager with custom mapper.
dbm := dbdog.Manager{
    TableMapper: tableMapper,
}
```

## Step Definitions

Delete all rows from table.

```gherkin
Given there are no rows in table "my_table" of database "my_db"
```

Populate rows in a database.

```gherkin
And these rows are stored in table "my_table" of database "my_db"
| id | foo   | bar | created_at           | deleted_at           |
| 1  | foo-1 | abc | 2021-01-01T00:00:00Z | NULL                 |
| 2  | foo-1 | def | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
| 3  | foo-2 | hij | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |
```

Assert rows existence in a database.

For each row in gherkin table database is queried to find a row with `WHERE` condition that includes provided column
values.

If a column has `NULL` value, it is excluded from `WHERE` condition.

Column can contain variable (any unique string starting with `$` or other prefix configured with `Manager.VarPrefix`).
If variable has not yet been populated, it is excluded from `WHERE` condition and populated with value received from
database. When this variable is used in next steps, it replaces the value of column with value of variable.

Variables can help to assert consistency of dynamic data, for example variable can be populated as ID of one entity and
then checked as foreign key value of another entity. This can be especially helpful in cases of UUIDs.

If column value represents JSON array or object it is excluded from `WHERE` condition, value assertion is done by
comparing Go value mapped from database row field with Go value mapped from gherkin table cell.

```gherkin
Then these rows are available in table "my_table" of database "my_db"
| id   | foo   | bar | created_at           | deleted_at           |
| $id1 | foo-1 | abc | 2021-01-01T00:00:00Z | NULL                 |
| $id2 | foo-1 | def | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
| $id3 | foo-2 | hij | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |
```

It is possible to check table contents exhaustively by adding "only" to step statement. Such assertion will also make
sure that total number of rows in database table matches number of rows in gherkin table.

```gherkin
Then only these rows are available in table "my_table" of database "my_db"
| id   | foo   | bar | created_at           | deleted_at           |
| $id1 | foo-1 | abc | 2021-01-01T00:00:00Z | NULL                 |
| $id2 | foo-1 | def | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
| $id3 | foo-2 | hij | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |
```

Assert no rows exist in a database.

```gherkin
And no rows are available in table "my_another_table" of database "my_db"
```
