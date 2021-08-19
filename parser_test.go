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
				Fields: []sql.Ident{{Name: "name"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "tbl"},
				},
				// Fields:    []string{"name"},
				// From: "tbl",
			},
		},

		// Multi-field statement
		{
			s: `SELECT first_name, last_name, age FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
				// Fields:    []string{"first_name", "last_name", "age"},
				// From: "my_table",
			},
		},

		// Multi-field statement with whitespaces
		{
			s: `SELECT first_name
,last_name
, age  

FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
			},
		},
		// Offset Value statement
		{
			s: `SELECT first_name, last_name, age FROM my_table ORDER BY age OFFSET 100 LIMIT 4`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
				OrderBy: &sql.OrderByClause{Fields: []*sql.Ident{{Name: "age"}}},
				Limit:   &sql.LimitClause{Value: 4},
				Offset:  &sql.OffsetClause{Value: 100},
			},
		},

		// Group By statement
		{
			s: `SELECT first_name, last_name, age FROM my_table GROUP BY first_name, last_name`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
				GroupBy: &sql.GroupByClause{
					Fields: []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}},
				},
			},
		},

		// Order By statement
		{
			s: `SELECT first_name, last_name, age FROM my_table ORDER BY first_name, last_name`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
				OrderBy: &sql.OrderByClause{
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
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
				Where: &sql.WhereClause{
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
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
				Where: &sql.WhereClause{
					Predicate: &sql.BinaryExpr{
						LHS: &sql.Ident{Name: "first_name"},
						Op:  sql.EQ,
						RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
					},
				},
			},
		},

		// Select all statement
		// {
		// 	s: `SELECT a as b FROM my_table`,
		// 	stmt: &sql.SelectStmt{
		// 		Fields:    []sql.Expr{&sql.AliasExpr{Expr: &sql.Ident{Name: "a"}, Alias: sql.Ident{Name: "b"}}},
		// 		From: sql.Ident{Name: "my_table"},
		// 	},
		// },

		// Join statement
		{
			s: `SELECT t1.a, t2.b FROM t1 JOIN t2 ON t1.a = t2.c`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "t1.a"}, {Name: "t2.b"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
					Join: &sql.JoinSubClause{
						TableName: &sql.Ident{Name: "t2"},
						Kind:      sql.InnerJoin,
						Criterion: sql.BinaryExpr{
							LHS: &sql.Ident{Name: "t1.a"},
							Op:  sql.EQ,
							RHS: &sql.Ident{Name: "t2.c"},
						},
					},
				},
			},
		},

		// all types of join statement
		{
			s: `SELECT t1.a, t2.b 
				FROM t1 
					JOIN t2 ON t1.a = t2.c
					LEFT JOIN t3 ON t2.d = t3.a
					FULL JOIN t4 ON t1.a = t4.b
					RIGHT OUTER JOIN t5 ON t3.c = t5.x`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "t1.a"}, {Name: "t2.b"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
					Join: &sql.JoinSubClause{
						TableName: &sql.Ident{Name: "t2"},
						Kind:      sql.InnerJoin,
						Criterion: sql.BinaryExpr{
							LHS: &sql.Ident{Name: "t1.a"},
							Op:  sql.EQ,
							RHS: &sql.Ident{Name: "t2.c"},
						},
						Join: &sql.JoinSubClause{
							TableName: &sql.Ident{Name: "t3"},
							Kind:      sql.LeftOuterJoin,
							Criterion: sql.BinaryExpr{
								LHS: &sql.Ident{Name: "t2.d"},
								Op:  sql.EQ,
								RHS: &sql.Ident{Name: "t3.a"},
							},
							Join: &sql.JoinSubClause{
								TableName: &sql.Ident{Name: "t4"},
								Kind:      sql.FullOuterJoin,
								Criterion: sql.BinaryExpr{
									LHS: &sql.Ident{Name: "t1.a"},
									Op:  sql.EQ,
									RHS: &sql.Ident{Name: "t4.b"},
								},
								Join: &sql.JoinSubClause{
									TableName: &sql.Ident{Name: "t5"},
									Kind:      sql.RightOuterJoin,
									Criterion: sql.BinaryExpr{
										LHS: &sql.Ident{Name: "t3.c"},
										Op:  sql.EQ,
										RHS: &sql.Ident{Name: "t5.x"},
									},
								},
							},
						},
					},
				},
			},
		},

		// Select all statement
		{
			s: `SELECT * FROM my_table`,
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "*"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
			},
		},

		// Errors
		{s: `foo`, err: `found "foo", expected SELECT`},
		{s: `SELECT !`, err: `found "!", expected field`},
		{s: `SELECT field xxx`, err: `found "xxx", expected FROM`},
		{s: `SELECT field FROM *`, err: `found "*", expected table name`},
		{s: `SELECT field FROM table OFFSET 1`, err: `found "OFFSET", invalid after FROM <table>`},
		{s: `SELECT field FROM table ORDER BY field OFFSET 1.5`, err: `found "1.5", expected INT offset value`},
		{s: `SELECT field FROM table ORDER BY field OFFSET -1`, err: `found "-1", expected nonnegative INT`},
		{s: `SELECT field FROM table LIMIT -1`, err: `found "-1", expected nonnegative INT`},
		{s: `SELECT field FROM table1 JOIN table2 LIMIT -1`, err: `found "LIMIT", expected ON keyword`},
		{s: `SELECT field FROM table1 JOIN table2`, err: `found "", expected ON keyword`},
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
