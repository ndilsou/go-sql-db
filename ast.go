package sql

type Stmt interface {
	stmtNode()
}

type SelectStmt struct {
	TableName     Ident
	Fields        []*Ident
	WhereClause   *WhereClause
	GroupByClause *GroupByClause
	LimitClause   *LimitClause
	OffsetClause  *OffsetClause
}

func (*SelectStmt) stmtNode() {}

type Expr interface {
	exprNode()
}

type Ident struct {
	Name string
}

type BasicLit struct {
	Kind  Token
	Value string
}

type UnaryExpr struct {
	Op Token
	X  Expr
}

type BinaryExpr struct {
	LHS Expr
	Op  Token
	RHS Expr
}

type LogicalExpr struct {
	Expr      Expr
	IsNegated bool
	Op        *Token
	Then      *LogicalExpr
}

func (*Ident) exprNode()       {}
func (*BasicLit) exprNode()    {}
func (*UnaryExpr) exprNode()   {}
func (*BinaryExpr) exprNode()  {}
func (*LogicalExpr) exprNode() {}

type WhereClause struct {
	// Predicate LogicalExpr
	Predicate Expr
}
type GroupByClause struct {
	Fields []*Ident
}
type LimitClause struct{ Value int }
type OffsetClause struct{ Value int }

func (id *Ident) String() string {
	if id == nil {
		return "NULL"
	}
	return id.Name
}

func (l *BasicLit) String() string {
	return l.Value
}
