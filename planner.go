package sql

import (
	"errors"
)

/*
 Logical Plan Actions:
-
*/
type ParentRelationship int

const (
	OUTER_REL ParentRelationship = iota
	INNER_REL
)

type PlanNode interface {
	Output() []string
	ParentRelationship() ParentRelationship
	Plans() []PlanNode
}

func planNodePrecedence(node PlanNode) int {
	var p int
	switch node.(type) {
	case *LimitNode:
		p = 0
	case *OffsetNode:
		p = 1
	case *SortNode:
		p = 2
	case *NestedLoopNode:
		p = 3
	case *FilterNode:
		p = 4
	case *TableScanNode:
		p = 5
	default:
		p = -1
	}
	return p
}

// TableScanNode is a full table scan
type TableScanNode struct {
	Schema       string
	RelationName string
}

func (t *TableScanNode) Output() []string {
	panic("implement me")
}

func (t *TableScanNode) ParentRelationship() ParentRelationship {
	panic("implement me")
}

func (t *TableScanNode) Plans() []PlanNode {
	panic("implement me")
}

// SortNode is an in memory sort of the working set
type SortNode struct {
	Keys   []string
	Method string
}

func (s *SortNode) Output() []string {
	panic("implement me")
}

func (s *SortNode) ParentRelationship() ParentRelationship {
	panic("implement me")
}

func (s *SortNode) Plans() []PlanNode {
	panic("implement me")
}

// NestedLoopNode is a join without indexes
type NestedLoopNode struct {
	JoinType string
}

func (n *NestedLoopNode) Output() []string {
	panic("implement me")
}

func (n *NestedLoopNode) ParentRelationship() ParentRelationship {
	panic("implement me")
}

func (n *NestedLoopNode) Plans() []PlanNode {
	panic("implement me")
}

// LimitNode is a limit to the number of rows returned.
type LimitNode struct {
	Rows int
}

func (l *LimitNode) Output() []string {
	panic("implement me")
}

func (l *LimitNode) ParentRelationship() ParentRelationship {
	panic("implement me")
}

func (l *LimitNode) Plans() []PlanNode {
	panic("implement me")
}

// OffsetNode is an offset of the working set.
type OffsetNode struct {
	Rows int
}

func (o *OffsetNode) Output() []string {
	panic("implement me")
}

func (o *OffsetNode) ParentRelationship() ParentRelationship {
	panic("implement me")
}

func (o *OffsetNode) Plans() []PlanNode {
	panic("implement me")
}

// FilterNode is a filter based on conditions in the where clause.
type FilterNode struct {
	Filter string
}

func (f *FilterNode) Output() []string {
	panic("implement me")
}

func (f *FilterNode) ParentRelationship() ParentRelationship {
	panic("implement me")
}

func (f *FilterNode) Plans() []PlanNode {
	panic("implement me")
}

type Catalog interface {
	GetRelation(string) (*Relation, error)
}

type Relation struct {
	Name     string
	Location interface{}
	Schema   map[string]Column
}

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
		return p.planSelect(s)
	default:
		return nil, errors.New("unknown statement type")
	}
	// _, err := p.findRelations(stmt)
	// if err != nil {
	// 	return nil, err
	// }
	// return nil, nil
}

// findRelations walks the Stmt and find any relation referenced.
// It returns an error if a relation is missing.
func (p Planner) findRelations(stmt Stmt) ([]*Relation, error) {
	var relations []*Relation
	switch v := stmt.(type) {
	case *SelectStmt:
		n := v.TableName.Name
		r, err := p.c.GetRelation(n)
		if err != nil {
			return nil, err
		}
		relations = append(relations, r)
		return relations, nil
	default:
		return nil, errors.New("unknown statement type")
	}
}

func (p Planner) planSelect(stmt *SelectStmt) (PlanNode, error) {
	/*
		Plan todolist
		1. Check all relations exists
		2. Check all fields referenced in SELECT, WHERE, GROUPBY exist in one of the relations
		3. Create Nodes starting from last operation:
			1. Limit
			2. Offset
			3. Sort
			4. NestedLoop
			5. Filter
			6. Scan
	*/
	n := stmt.TableName.Name
	_, err := p.c.GetRelation(n)
	if err != nil {
		return nil, err
	}
	// for _, field := range stmt.Fields {
	//
	// }

	return nil, nil
}
