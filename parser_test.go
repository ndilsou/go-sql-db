package sql_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ndilsou/go-rdbms-playground"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *sql.SelectStmt
		err  string
	}{
		// Single field statement
		{
			s: `SELECT name FROM tbl`,
			stmt: &sql.SelectStmt{
				Fields:    []string{"name"},
				TableName: "tbl",
			},
		},

		// Multi-field statement
		{
			s: `SELECT first_name, last_name, age FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields:    []string{"first_name", "last_name", "age"},
				TableName: "my_table",
			},
		},

		// Multi-field statement with whitespaces
		{
			s: `SELECT first_name
,last_name
, age  

FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields:    []string{"first_name", "last_name", "age"},
				TableName: "my_table",
			},
		},
		// Offset Limit statement
		{
			s: `SELECT first_name, last_name, age FROM my_table OFFSET 100 LIMIT 4`,
			stmt: &sql.SelectStmt{
				Fields:       []string{"first_name", "last_name", "age"},
				TableName:    "my_table",
				LimitClause:  &sql.LimitClause{Value: 4},
				OffsetClause: &sql.OffsetClause{Value: 100},
			},
		},

		// Group By statement
		{
			s: `SELECT first_name, last_name, age FROM my_table GROUP BY first_name, last_name`,
			stmt: &sql.SelectStmt{
				Fields:    []string{"first_name", "last_name", "age"},
				TableName: "my_table",
				GroupByClause: &sql.GroupByClause{
					Fields: []string{"first_name", "last_name"},
				},
			},
		},

		// Where statement
		{
			s: `SELECT first_name, last_name, age FROM my_table WHERE first_name = 1`,
			stmt: &sql.SelectStmt{
				Fields:      []string{"first_name", "last_name", "age"},
				TableName:   "my_table",
				WhereClause: &sql.WhereClause{},
			},
		},

		// Select all statement
		{
			s: `SELECT * FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields:    []string{"*"},
				TableName: "my_table",
			},
		},

		// Errors
		{s: `foo`, err: `found "foo", expected SELECT`},
		{s: `SELECT !`, err: `found "!", expected field`},
		{s: `SELECT field xxx`, err: `found "xxx", expected FROM`},
		{s: `SELECT field FROM *`, err: `found "*", expected table name`},
		{s: `SELECT field FROM table OFFSET 1.5`, err: `found "1.5", expected INT`},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			stmt, err := sql.NewParser(strings.NewReader(tt.s)).Parse()
			if !reflect.DeepEqual(tt.err, errstring(err)) {
				t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
			} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
				t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
			}
		})
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
