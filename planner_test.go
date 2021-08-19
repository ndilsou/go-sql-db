package sql_test

import (
	"errors"
	"reflect"
	"testing"

	sql "github.com/ndilsou/go-rdbms-playground"
)

func TestPlanner_Plan(t *testing.T) {
	tests := []struct {
		name    string
		stmt    sql.Stmt
		want    sql.PlanNode
		wantErr bool
	}{
		{
			name: "no relation",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "my_table"},
				},
			},
			wantErr: true,
		},
		{
			name: "no column",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
			},
			wantErr: true,
		},
		{
			name: "simple plan",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
			},
			want: &sql.ProjectionNode{
				Columns: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: &sql.TableScanNode{
					RelationName: "t1",
				},
			},
		},
		{
			name: "plan with where",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				Where: &sql.WhereClause{
					Predicate: &sql.BinaryExpr{
						LHS: &sql.Ident{Name: "a"},
						Op:  sql.EQ,
						RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
					},
				},
			},
			want: &sql.ProjectionNode{
				Columns: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: &sql.FilterNode{
					Filter: &sql.BinaryExpr{
						LHS: &sql.Ident{Name: "a"},
						Op:  sql.EQ,
						RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
					},
					From: &sql.TableScanNode{
						RelationName: "t1",
					},
				},
			},
		},
		{
			name: "plan with limit",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				Limit: &sql.LimitClause{
					Value: 10,
				},
			},
			want: &sql.LimitNode{
				Value: 10,
				From: &sql.ProjectionNode{
					Columns: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
					From: &sql.TableScanNode{
						RelationName: "t1",
					},
				},
			},
		},
		{
			name: "plan with limit and offset and order by",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				Limit: &sql.LimitClause{
					Value: 10,
				},
				Offset:  &sql.OffsetClause{Value: 5},
				OrderBy: &sql.OrderByClause{Fields: []*sql.Ident{{Name: "a"}}},
			},
			want: &sql.LimitNode{
				Value: 10,
				From: &sql.OffsetNode{
					Value: 5,
					From: &sql.SortNode{
						Keys: []string{"a"},
						From: &sql.ProjectionNode{
							Columns: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
							From: &sql.TableScanNode{
								RelationName: "t1",
							},
						},
					},
				},
			},
		},
		{
			name: "plan with order by",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				OrderBy: &sql.OrderByClause{Fields: []*sql.Ident{{Name: "a"}}},
			},
			want: &sql.SortNode{
				Keys: []string{"a"},
				From: &sql.ProjectionNode{
					Columns: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
					From: &sql.TableScanNode{
						RelationName: "t1",
					},
				},
			},
		},
		{
			name: "plan with filter and unknown column",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				Where: &sql.WhereClause{
					Predicate: &sql.BinaryExpr{
						LHS: &sql.Ident{Name: "first_name"},
						Op:  sql.EQ,
						RHS: &sql.BasicLit{Kind: sql.INT, Value: "1"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "plan with offset and no limit",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				Offset: &sql.OffsetClause{Value: 10},
			},
			wantErr: true,
		},
		{
			name: "plan with limit and no projection",
			stmt: &sql.SelectStmt{
				Limit: &sql.LimitClause{Value: 10},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := sql.NewPlanner(&mockCatalog{})
			got, err := p.Plan(tt.stmt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Plan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Plan() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var testRelations = map[string]sql.Relation{
	"t1": {
		Name:     "t1",
		Location: nil,
		Schema: map[string]sql.Column{
			"a": {
				Name:     "a",
				Type:     sql.INTEGER,
				Location: nil,
			},
			"b": {
				Name:     "b",
				Type:     sql.REAL,
				Location: nil,
			},
			"c": {
				Name:     "b",
				Type:     sql.TEXT,
				Location: nil,
			},
		},
	},
}

type mockCatalog struct {
}

func (m *mockCatalog) HasColumn(columnName string) bool {
	for _, v := range testRelations {
		if v.HasColumn(columnName) {
			return true
		}
	}
	return false
}

func (m *mockCatalog) GetRelation(s string) (sql.Relation, error) {
	r, ok := testRelations[s]
	if !ok {
		return sql.Relation{}, errors.New("no relation")
	}
	return r, nil
}

func BenchmarkPlanner_Plan(b *testing.B) {
	benchmarks := []struct {
		name string
		stmt sql.Stmt
	}{
		{
			name: "simple plan",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
			},
		},
		{
			name: "plan with limit and offset and order by",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
				},
				Limit: &sql.LimitClause{
					Value: 10,
				},
				Offset:  &sql.OffsetClause{Value: 5},
				OrderBy: &sql.OrderByClause{Fields: []*sql.Ident{{Name: "a"}}},
			},
		},
		{
			name: "plan with filter and unknown column",
			stmt: &sql.SelectStmt{
				Fields: []sql.Ident{{Name: "a"}, {Name: "b"}, {Name: "c"}},
				From: sql.FromClause{
					TableName: &sql.Ident{Name: "t1"},
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
	}
	var plan sql.PlanNode
	var err error
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			p := sql.NewPlanner(&mockCatalog{})
			for i := 0; i < b.N; i++ {
				plan, err = p.Plan(bm.stmt)
			}
		})
	}
	ignoreBenchResult(plan, err)
}
