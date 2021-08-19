package sql

type JoinKind int

const (
	InnerJoin JoinKind = iota
	LeftOuterJoin
	RightOuterJoin
	FullOuterJoin
)

type Stmt interface {
	stmtNode()
}

type SelectStmt struct {
	From    FromClause
	Fields  []Ident
	Where   *WhereClause
	GroupBy *GroupByClause
	OrderBy *OrderByClause
	Limit   *LimitClause
	Offset  *OffsetClause
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

type AliasExpr struct {
	Expr  Expr
	Alias Ident
}

func (*Ident) exprNode()      {}
func (*BasicLit) exprNode()   {}
func (*UnaryExpr) exprNode()  {}
func (*BinaryExpr) exprNode() {}
func (*AliasExpr) exprNode()  {}

type WhereClause struct {
	Predicate Expr
}

type OrderByClause struct {
	Fields []*Ident
}

type GroupByClause struct {
	Fields []*Ident
}

type FromClause struct {
	TableName Expr
	Join      *JoinSubClause
}

type JoinSubClause struct {
	TableName Expr
	Kind      JoinKind
	Criterion BinaryExpr
	Join      *JoinSubClause
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
