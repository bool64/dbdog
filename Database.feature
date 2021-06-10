Feature: Database Query

  Scenario: Successful Query
    Given there are no rows in table "my_table" of database "my_db"

    And rows from this file are stored in table "my_table" of database "my_db"
    """
    _testdata/rows.csv
    """

    Then only these rows are available in table "my_table" of database "my_db"
      | id   | foo   | bar | created_at           | deleted_at           |
      | $id1 | $foo1 | abc | 2021-01-01T00:00:00Z | NULL                 |
      | $id2 | $foo1 | def | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
      | $id3 | foo-2 | hij | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |

    Then only these rows are available in table "my_table" of database "my_db"
      | id   | foo   | bar | created_at           | deleted_at           |
      | $id1 | $foo1 | abc | 2021-01-01T00:00:00Z | NULL                 |
      | $id2 | $foo1 | def | 2021-01-02T00:00:00Z | 2021-01-03T00:00:00Z |
      | $id3 | foo-2 | hij | 2021-01-03T00:00:00Z | 2021-01-03T00:00:00Z |

    And no rows are available in table "my_another_table" of database "my_db"
