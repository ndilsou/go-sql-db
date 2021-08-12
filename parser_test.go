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
	tests := []struct {
		s    string
		stmt *sql.SelectStmt
		err  string
	}{
		// Single field statement
		{
			s: `SELECT name FROM tbl`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "name"}},
				TableName: sql.Ident{Name: "tbl"},
				// Fields:    []string{"name"},
				// TableName: "tbl",
			},
		},

		// Multi-field statement
		{
			s: `SELECT first_name, last_name, age FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "my_table"},
				// Fields:    []string{"first_name", "last_name", "age"},
				// TableName: "my_table",
			},
		},

		// Multi-field statement with whitespaces
		{
			s: `SELECT first_name
,last_name
, age  

FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "my_table"},
			},
		},
		// Offset Limit statement
		{
			s: `SELECT first_name, last_name, age FROM my_table OFFSET 100 LIMIT 4`,
			stmt: &sql.SelectStmt{
				Fields:       []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName:    sql.Ident{Name: "my_table"},
				LimitClause:  &sql.LimitClause{Value: 4},
				OffsetClause: &sql.OffsetClause{Value: 100},
			},
		},

		// Group By statement
		{
			s: `SELECT first_name, last_name, age FROM my_table GROUP BY first_name, last_name`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "my_table"},
				GroupByClause: &sql.GroupByClause{
					Fields: []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}},
				},
			},
		},

		// Where with multiple statement
		{
			s: `SELECT first_name, last_name, age 
					FROM my_table 
					WHERE first_name = 1 AND last_name <> 'TEST' OR age > 18;`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "my_table"},
				WhereClause: &sql.WhereClause{
					Predicate: &sql.BinaryExpr{
						LHS: &sql.BinaryExpr{
							LHS: &sql.Ident{Name: "first_name"},
							Op:  sql.EQ,
							RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
						},
						Op: sql.AND,
						RHS: &sql.BinaryExpr{
							LHS: &sql.BinaryExpr{
								LHS: &sql.Ident{Name: "last_name"},
								Op:  sql.NEQ,
								RHS: &sql.BasicLit{Kind: sql.STRING, Value: "'TEST'"},
							},
							Op: sql.OR,
							RHS: &sql.BinaryExpr{
								LHS: &sql.Ident{Name: "age"},
								Op:  sql.GT,
								RHS: &sql.BasicLit{Kind: sql.INT, Value: "18"},
							},
						},
					},
				},
			},
		},

		// Where statement
		{
			s: `SELECT first_name, last_name, age FROM my_table WHERE first_name = 1`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "my_table"},
				WhereClause: &sql.WhereClause{
					Predicate: &sql.BinaryExpr{
						LHS: &sql.Ident{Name: "first_name"},
						Op:  sql.EQ,
						RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
					},
				},
			},
		},

		// // Where statement
		// {
		// 	s: `SELECT first_name, last_name, age FROM my_table WHERE first_name = 1`,
		// 	stmt: &sql.SelectStmt{
		// 		Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
		// 		TableName: sql.Ident{Name: "my_table"},
		// 		WhereClause: &sql.WhereClause{
		// 			Predicate: sql.LogicalExpr{
		// 				Expr: &sql.BinaryExpr{
		// 					LHS: &sql.Ident{Name: "first_name"},
		// 					Op:  sql.EQ,
		// 					RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
		// 				},
		// 			},
		// 		},
		// 	},
		// },

		// Select all statement
		{
			s: `SELECT * FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "*"}},
				TableName: sql.Ident{Name: "my_table"},
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
