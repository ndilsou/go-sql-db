package sql

type SelectStmt struct {
	TableName     string
	Fields        []string
	WhereClause   *WhereClause
	GroupByClause *GroupByClause
	LimitClause   *LimitClause
	OffsetClause  *OffsetClause
}

type Expr interface {
	exprNode()
}

type UnaryExpr struct {
	Op Token
	X  Expr
}

type BinaryExpr struct {
	X  Expr
	Op Token
	Y  Expr
}

type WhereClause struct {
	Predicates []*Expr
}
type GroupByClause struct {
	Fields []string
}
type LimitClause struct{ Value int }
type OffsetClause struct{ Value int }
