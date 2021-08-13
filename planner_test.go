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
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "my_table"},
			},
			wantErr: true,
		},
		{
			name: "no column",
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "t1"},
			},
			wantErr: false,
		},
		{
			name: "has relation",
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "t1"},
			},
			wantErr: false,
		},
		{
			name: "simple plan",
			stmt: &sql.SelectStmt{
				Fields:    []*sql.Ident{{Name: "first_name"}, {Name: "last_name"}, {Name: "age"}},
				TableName: sql.Ident{Name: "t1"},
			},
			// want: &sql.,
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

var testRelations = map[string]*sql.Relation{
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

func (m *mockCatalog) GetRelation(s string) (*sql.Relation, error) {
	r, ok := testRelations[s]
	if !ok {
		return nil, errors.New("no relation")
	}
	return r, nil
}
