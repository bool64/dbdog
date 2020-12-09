Feature: Database Query

  Scenario: Failing Query
    Then only these rows are available in table "my_table" of database "my_db":
      | id | foo   | bar | created_at           | deleted_at |
      | 1  | foo-1 | abc | 2021-01-01T00:00:00Z | NULL       |
