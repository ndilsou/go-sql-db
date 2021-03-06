package sql

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	s    Scanner
	stmt SelectStmt
	err  error
	buf  struct {
		l Lexeme
		n int
	}
}

func NewParser(r *strings.Reader) *Parser {
	return &Parser{s: *NewScanner(r)}
}

func (p *Parser) Parse() (*SelectStmt, error) {
	// TODO: Add support for aliases
	for next := parseStmtInit(p); next != nil; {
		next = next(p)
	}
	return &p.stmt, p.err
}

func (p *Parser) scan() Lexeme {
	if p.buf.n == 1 {
		p.buf.n = 0
		return p.buf.l
	}

	l := p.s.Scan()
	p.buf.l = l
	return l
}

func (p *Parser) unscan() {
	p.buf.n = 1
}

func (p *Parser) peek() Lexeme {
	l := p.scan()
	p.unscan()
	return l
}

func (p *Parser) scanIgnoreWhitespace() Lexeme {
	l := p.scan()
	if l.Token == WS {
		l = p.scan()
	}
	return l
}

type parseFunc func(p *Parser) parseFunc

func parseStmtInit(p *Parser) parseFunc {
	l := p.scan()
	if l.Token != SELECT {
		p.err = fmt.Errorf("found \"%s\", expected SELECT", l.Lit)
		return nil
	}
	return parseSelectFields
}

func parseSelectFields(p *Parser) parseFunc {
	if l := p.scan(); l.Token == IDENT || l.Token == ASTERISK {
		p.stmt.Fields = append(p.stmt.Fields, Ident{Name: l.Lit})
	} else {
		p.err = fmt.Errorf("found \"%s\", expected field", l.Lit)
		return nil
	}

	if l := p.scan(); l.Token == COMMA {
		return parseSelectFields(p)
	} else {
		p.unscan()
		return parseFrom
	}
}

func parseFrom(p *Parser) parseFunc {
	if l := p.scan(); l.Token != FROM {
		p.err = fmt.Errorf("found \"%s\", expected FROM", l.Lit)
		return nil
	}

	if l := p.scan(); l.Token == IDENT {
		p.stmt.From = FromClause{TableName: &Ident{Name: l.Lit}}
	} else {
		p.err = fmt.Errorf("found \"%s\", expected table name", l.Lit)
		return nil
	}
	// if p.peek().Token.IsTerminal() {
	// 	return parseTerminalLexeme
	// }

	var next parseFunc
	switch n := p.peek(); n.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case JOIN, LEFT, RIGHT, FULL:
		next = parseJoin
	case WHERE:
		next = parseWhere
	case GROUP:
		next = parseGroupBy
	case LIMIT:
		next = parseLimit
	case ORDER:
		next = parseOrderBy
	default:
		p.err = fmt.Errorf(`found "%s", invalid after FROM <table>`, n.Lit)
		next = nil
	}
	return next
}

func parseJoin(p *Parser) parseFunc {
	l := p.scan()
	var kind JoinKind
	switch l.Token {
	case JOIN:
		kind = InnerJoin
	case INNER:
		if n := p.scan(); n.Token != JOIN {
			p.err = fmt.Errorf(`found "%s", expected JOIN`, n.Lit)
			return nil
		}
		kind = InnerJoin
	case LEFT:
		if err := skipOuterJoinKeywords(p); err != nil {
			p.err = err
			return nil
		}
		kind = LeftOuterJoin
	case RIGHT:
		if err := skipOuterJoinKeywords(p); err != nil {
			p.err = err
			return nil
		}
		kind = RightOuterJoin
	case FULL:
		if err := skipOuterJoinKeywords(p); err != nil {
			p.err = err
			return nil
		}
		kind = FullOuterJoin
	}

	l = p.scan()
	if l.Token != IDENT {
		p.err = fmt.Errorf("found \"%s\", expected table name", l.Lit)
		return nil
	}
	t := Ident{Name: l.Lit}

	l = p.scan()
	if l.Token != ON {
		p.err = fmt.Errorf("found \"%s\", expected ON keyword", l.Lit)
		return nil
	}

	expr, err := extractBinaryExpr(p)
	if err != nil {
		p.err = err
		return nil
	}

	join := JoinSubClause{
		TableName: &t,
		Kind:      kind,
		Criterion: expr,
	}
	p.stmt.From.Join = appendJoinSubClause(p.stmt.From.Join, join)

	var next parseFunc
	switch n := p.peek(); n.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case JOIN, LEFT, RIGHT, FULL:
		next = parseJoin
	case WHERE:
		next = parseWhere
	case GROUP:
		next = parseGroupBy
	case LIMIT:
		next = parseLimit
	case ORDER:
		next = parseOrderBy
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after FROM <table>", n.Lit)
		next = nil
	}
	return next
}

func appendJoinSubClause(join *JoinSubClause, next JoinSubClause) *JoinSubClause {
	if join == nil {
		return &next
	}

	j := join
	for {
		if j.Join == nil {
			j.Join = &next
			break
		}

		j = j.Join
	}
	return join
}

func parseLimit(p *Parser) parseFunc {
	if p.stmt.Limit != nil {
		p.err = errors.New("LIMIT already defined in statement")
		return nil
	}

	l := p.scan()
	if l.Token != LIMIT {
		p.err = fmt.Errorf("found \"%s\", expected LIMIT", l.Lit)
		return nil
	}
	p.stmt.Limit = new(LimitClause)

	l = p.scan()
	if l.Token != INT {
		p.err = fmt.Errorf("found \"%s\", expected INT offset value", l.Lit)
		return nil
	}

	v, err := strconv.Atoi(l.Lit)
	if err != nil {
		p.err = fmt.Errorf("cannot parse limit, literal \"%s\" is not INT", l.Lit)
		return nil
	}
	if v < 0 {
		p.err = fmt.Errorf("found \"%s\", expected nonnegative INT", l.Lit)
		return nil
	}
	p.stmt.Limit.Value = v

	var next parseFunc
	switch n := p.peek(); n.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case OFFSET:
		next = parseOffset
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after OFFSET <offset>", n.Lit)
		next = nil
	}
	return next
}

func parseOrderBy(p *Parser) parseFunc {
	if p.stmt.OrderBy != nil {
		p.err = errors.New("ORDER BY already defined in statement")
		return nil
	}

	l1 := p.scan()
	l2 := p.scan()
	if l1.Token != ORDER || l2.Token != BY {
		p.err = fmt.Errorf("found \"%s %s\", expected ORDER BY", l1.Lit, l2.Lit)
		return nil
	}
	p.stmt.OrderBy = new(OrderByClause)

	return parseOrderByFields
}

func parseOrderByFields(p *Parser) parseFunc {
	if l := p.scan(); l.Token == IDENT {
		p.stmt.OrderBy.Fields = append(p.stmt.OrderBy.Fields, &Ident{Name: l.Lit})
	} else {
		p.err = fmt.Errorf("found \"%s\", expected field", l.Lit)
		return nil
	}

	l := p.scan()
	if l.Token == COMMA {
		return parseOrderByFields(p)
	}

	var next parseFunc
	switch l.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case OFFSET:
		next = parseOffset
	case LIMIT:
		next = parseLimit
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after OFFSET <offset>", l.Lit)
		next = nil
	}
	p.unscan()

	return next
}

func parseGroupBy(p *Parser) parseFunc {
	if p.stmt.GroupBy != nil {
		p.err = errors.New("GROUP BY already defined in statement")
		return nil
	}

	l1 := p.scan()
	l2 := p.scan()
	if l1.Token != GROUP || l2.Token != BY {
		p.err = fmt.Errorf("found \"%s %s\", expected GROUP BY", l1.Lit, l2.Lit)
		return nil
	}
	p.stmt.GroupBy = new(GroupByClause)

	return parseGroupByFields
}

func parseGroupByFields(p *Parser) parseFunc {
	if l := p.scan(); l.Token == IDENT {
		p.stmt.GroupBy.Fields = append(p.stmt.GroupBy.Fields, &Ident{Name: l.Lit})
	} else {
		p.err = fmt.Errorf("found \"%s\", expected field", l.Lit)
		return nil
	}

	l := p.scan()
	if l.Token == COMMA {
		return parseGroupByFields(p)
	}

	var next parseFunc
	switch l.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case ORDER:
		next = parseOrderBy
	case OFFSET:
		next = parseOffset
	case LIMIT:
		next = parseLimit
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after OFFSET <offset>", l.Lit)
		next = nil
	}
	p.unscan()

	return next
}

func parseWhere(p *Parser) parseFunc {
	l := p.scan()
	if l.Token != WHERE {
		p.err = fmt.Errorf("found \"%s\", expected WHERE", l.Lit)
		return nil
	}
	p.stmt.Where = new(WhereClause)

	if p.peek().Token.IsTerminal() {
		return parseTerminalLexeme
	}
	return parseWherePredicates
}

func parseWherePredicates(p *Parser) parseFunc {
	predicate, err := extractLogicalExpr(p)
	if err != nil {
		p.err = err
		return nil
	}

	p.stmt.Where.Predicate = predicate

	var next parseFunc
	switch nextLex := p.peek(); nextLex.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case GROUP:
		next = parseGroupBy
	case OFFSET:
		next = parseOffset
	case LIMIT:
		next = parseLimit
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after WHERE <predicate>", nextLex.Lit)
		next = nil
	}

	return next
}

func parseOffset(p *Parser) parseFunc {
	if p.stmt.OrderBy == nil {
		p.err = errors.New("OFFSET can only be defined for statement with ORDER BY")
		return nil
	}
	if p.stmt.Offset != nil {
		p.err = errors.New("OFFSET already defined in statement")
		return nil
	}

	l := p.scan()
	if l.Token != OFFSET {
		p.err = fmt.Errorf("found \"%s\", expected OFFSET", l.Lit)
		return nil
	}
	p.stmt.Offset = new(OffsetClause)

	l = p.scan()
	if l.Token != INT {
		p.err = fmt.Errorf("found \"%s\", expected INT offset value", l.Lit)
		return nil
	}
	v, err := strconv.Atoi(l.Lit)
	if err != nil {
		p.err = fmt.Errorf("cannot parse offset, literal \"%s\" is not INT", l.Lit)
		return nil
	}
	if v < 0 {
		p.err = fmt.Errorf("found \"%s\", expected nonnegative INT", l.Lit)
		return nil
	}
	p.stmt.Offset.Value = v

	var next parseFunc
	switch n := p.peek(); n.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case LIMIT:
		next = parseLimit
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after OFFSET <offset>", n.Lit)
		next = nil
	}
	return next
}

func parseTerminalLexeme(p *Parser) parseFunc {
	l := p.scan()
	if l.Token != EOF && l.Token != SEMICOLON {
		p.err = fmt.Errorf("found \"%s\", expected EOF", l.Lit)
	}

	return nil
}

func toLiteralExpr(l Lexeme) (Expr, error) {
	var expr Expr
	if l.Token == IDENT {
		expr = &Ident{Name: l.Lit}
	} else if l.Token.IsLiteral() {
		expr = &BasicLit{
			Kind:  l.Token,
			Value: l.Lit,
		}
	} else {
		return nil, fmt.Errorf("found \"%s\", expected literal", l.Lit)
	}
	return expr, nil
}

func extractLogicalExpr(p *Parser) (Expr, error) {
	// TODO: add support for UnaryExpr
	expr, err := extractBinaryExpr(p)
	if err != nil {
		return nil, err
	}

	var predicate Expr
	if op := p.scan(); op.Token == AND || op.Token == OR {
		rhs, err := extractLogicalExpr(p)
		if err != nil {
			return nil, err
		}
		predicate = &BinaryExpr{LHS: &expr, Op: op.Token, RHS: rhs}
	} else {
		p.unscan()
		predicate = &expr
	}
	return predicate, nil
}

func extractBinaryExpr(p *Parser) (BinaryExpr, error) {
	var expr BinaryExpr
	l := p.scan()
	if lhs, err := toLiteralExpr(l); err == nil {
		expr.LHS = lhs
	} else {
		return expr, err
	}

	l = p.scan()
	if !l.Token.IsComparisonOperator() {
		return expr, fmt.Errorf("found \"%s\", expected comparison operator", l.Lit)
	}
	expr.Op = l.Token

	l = p.scan()
	rhs, err := toLiteralExpr(l)
	if err != nil {
		return expr, err
	}
	expr.RHS = rhs

	return expr, nil
}

func skipOuterJoinKeywords(p *Parser) error {
	if n := p.scan(); n.Token != OUTER {
		p.unscan()
	}

	if n := p.scan(); n.Token != JOIN {
		return fmt.Errorf(`found "%s", expected JOIN`, n.Lit)
	}
	return nil
}
