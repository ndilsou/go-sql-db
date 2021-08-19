package sql

import (
	"errors"
	"fmt"
)

// PlanNode is a node in the query execution plan.
type PlanNode interface {
	planNode()
}

// ProjectionNode represents the columns to keep after a stage.
type ProjectionNode struct {
	Columns []Ident
	From    PlanNode
}

// TableScanNode is a full table scan
type TableScanNode struct {
	Schema       string
	RelationName string
}

// SortNode is an in memory sort of the working set
type SortNode struct {
	Keys   []string
	Method string
	From   PlanNode
}

// NestedLoopNode is a join without indexes
type NestedLoopNode struct {
	JoinType string
	From     PlanNode
}

// LimitNode is a limit to the number of rows returned.
type LimitNode struct {
	Value int
	From  PlanNode
}

// OffsetNode discards a number of rows from the top of the set.
type OffsetNode struct {
	Value int
	From  PlanNode
}

// FilterNode is a filter based on conditions in the where clause.
type FilterNode struct {
	Filter Expr
	From   PlanNode
}

func (*ProjectionNode) planNode() {}
func (*TableScanNode) planNode()  {}
func (*SortNode) planNode()       {}
func (*NestedLoopNode) planNode() {}
func (*LimitNode) planNode()      {}
func (*OffsetNode) planNode()     {}
func (*FilterNode) planNode()     {}

type Catalog interface {
	GetRelation(string) (Relation, error)
	// HasColumn checks if a column of this name exists in any relations of the catalog.
	HasColumn(string) bool
}

type NodeSchema struct {
	Relations map[string]Relation
}

type Relation struct {
	Name     string
	Location interface{}
	Schema   map[string]Column
}

func (r Relation) HasColumn(name string) bool {
	if _, ok := r.Schema[name]; ok {
		return true
	}
	return false
}

// Column is the metadata about a relation's column
type Column struct {
	Name     string
	Type     DataType
	Location interface{}
}

type Planner struct {
	c Catalog
}

func NewPlanner(c Catalog) *Planner {
	return &Planner{c: c}
}

func (p Planner) Plan(stmt Stmt) (PlanNode, error) {
	switch s := stmt.(type) {
	case *SelectStmt:
		return planSelect(p.c, s)
	default:
		return nil, errors.New("unknown statement type")
	}
}

type planFunc func(Catalog, *SelectStmt) (PlanNode, error)

func planSelect(catalog Catalog, stmt *SelectStmt) (PlanNode, error) {
	if stmt.Limit == nil && stmt.Offset != nil {
		return nil, errors.New("invalid SELECT: OFFSET without LIMIT")
	}

	var planFn planFunc
	if stmt.Limit != nil {
		planFn = planLimit
	} else if stmt.Offset != nil {
		planFn = planOffset

	} else if stmt.OrderBy != nil {
		planFn = planSort
	} else {
		planFn = planProjection
	}
	plan, err := planFn(catalog, stmt)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func planFilter(catalog Catalog, stmt *SelectStmt) (PlanNode, error) {
	if err := validateExpr(catalog, stmt.Where.Predicate); err != nil {
		return nil, err

	}
	plan := FilterNode{
		Filter: stmt.Where.Predicate,
	}

	n, ok := stmt.From.TableName.(*Ident)
	if !ok {
		return nil, errors.New("invalid expression in FROM clause")
	}
	relation, err := catalog.GetRelation(n.Name)
	if err != nil {
		return nil, err
	}
	from, err := planTableScan(relation)
	if err != nil {
		return nil, err
	}
	plan.From = from

	return &plan, nil
}

func planProjection(catalog Catalog, stmt *SelectStmt) (PlanNode, error) {
	if len(stmt.Fields) == 0 {
		return nil, errors.New("invalid statement: no columns in select")
	}

	n, ok := stmt.From.TableName.(*Ident)
	if !ok {
		return nil, errors.New("invalid expression in FROM clause")
	}
	relation, err := catalog.GetRelation(n.Name)
	if err != nil {
		return nil, err
	}

	var cols []Ident
	for _, field := range stmt.Fields {
		if !relation.HasColumn(field.Name) {
			return nil, fmt.Errorf("unknown column in statement: %s", field.Name)
		}
		cols = append(cols, field)
	}
	plan := ProjectionNode{
		Columns: cols,
	}

	var from PlanNode
	if stmt.Where != nil {
		from, err = planFilter(catalog, stmt)
		if err != nil {
			return nil, err
		}
	} else {
		from, err = planTableScan(relation)
		if err != nil {
			return nil, err
		}
	}
	plan.From = from

	return &plan, nil
}

func planLimit(catalog Catalog, stmt *SelectStmt) (PlanNode, error) {
	plan := LimitNode{
		Value: stmt.Limit.Value,
	}

	var planFn planFunc
	if stmt.Offset != nil {
		planFn = planOffset
	} else if stmt.OrderBy != nil {
		planFn = planSort
	} else {
		planFn = planProjection
	}

	from, err := planFn(catalog, stmt)
	if err != nil {
		return nil, err
	}
	plan.From = from

	return &plan, nil
}

func planOffset(catalog Catalog, stmt *SelectStmt) (PlanNode, error) {
	plan := OffsetNode{
		Value: stmt.Offset.Value,
	}

	var planFn planFunc
	if stmt.OrderBy != nil {
		planFn = planSort
	} else {
		planFn = planProjection
	}

	from, err := planFn(catalog, stmt)
	if err != nil {
		return nil, err
	}
	plan.From = from

	return &plan, nil
}

func planSort(catalog Catalog, stmt *SelectStmt) (PlanNode, error) {
	var keys []string
	for _, id := range stmt.OrderBy.Fields {
		keys = append(keys, id.Name)
	}
	plan := SortNode{
		Keys: keys,
	}

	from, err := planProjection(catalog, stmt)
	if err != nil {
		return nil, err
	}
	plan.From = from

	return &plan, nil
}

func planTableScan(r Relation) (PlanNode, error) {
	return &TableScanNode{
		RelationName: r.Name,
	}, nil
}

// validateExpr walks through an expression and ensure all identifiers present exist in the catalog.
func validateExpr(catalog Catalog, expr Expr) error {
	switch e := expr.(type) {
	case *Ident:
		if catalog.HasColumn(e.Name) {
			return nil
		}
		return fmt.Errorf("unknown column, \"%s\" in statement", e.Name)
	case *BasicLit:
		return nil
	case *UnaryExpr:
		return validateExpr(catalog, e.X)
	case *BinaryExpr:
		var err error
		if err = validateExpr(catalog, e.LHS); err != nil {
			return err
		} else if err = validateExpr(catalog, e.RHS); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid expression")
	}
}
