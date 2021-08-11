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
		p.stmt.Fields = append(p.stmt.Fields, l.Lit)
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
		p.stmt.TableName = l.Lit
	} else {
		p.err = fmt.Errorf("found \"%s\", expected table name", l.Lit)
		return nil
	}
	if p.peek().Token.IsTerminal() {
		return parseTerminalLexeme
	}

	var next parseFunc
	switch n := p.peek(); n.Token {
	case EOF, SEMICOLON:
		next = parseTerminalLexeme
	case WHERE:
		next = parseWhere
	case GROUP:
		next = parseGroupBy
	case LIMIT:
		next = parseLimit
	case OFFSET:
		next = parseOffset
	default:
		p.err = fmt.Errorf("found \"%s\", invalid after FROM <table>", n.Lit)
		next = nil
	}
	return next
}

func parseLimit(p *Parser) parseFunc {
	if p.stmt.LimitClause != nil {
		p.err = errors.New("LIMIT already defined in statement")
		return nil
	}

	if l := p.scan(); l.Token != LIMIT {
		p.err = fmt.Errorf("found \"%s\", expected LIMIT", l.Lit)
		return nil
	}
	p.stmt.LimitClause = new(LimitClause)

	if l := p.scan(); l.Token == INT {
		v, err := strconv.Atoi(l.Lit)
		if err != nil {
			p.err = fmt.Errorf("cannot parse limit, literal \"%s\" is not INT", l.Lit)
		}
		p.stmt.LimitClause.Value = v
	} else {
		p.err = fmt.Errorf("found \"%s\", expected INT limit value", l.Lit)
		return nil
	}

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

func parseGroupBy(p *Parser) parseFunc {
	if p.stmt.GroupByClause != nil {
		p.err = errors.New("GROUP BY already defined in statement")
		return nil
	}

	l1 := p.scan()
	l2 := p.scan()
	if l1.Token != GROUP || l2.Token != BY {
		p.err = fmt.Errorf("found \"%s %s\", expected GROUP BY", l1.Lit, l2.Lit)
		return nil
	}
	p.stmt.GroupByClause = new(GroupByClause)

	return parseGroupByFields
}

func parseGroupByFields(p *Parser) parseFunc {
	if l := p.scan(); l.Token == IDENT {
		p.stmt.GroupByClause.Fields = append(p.stmt.GroupByClause.Fields, l.Lit)
	} else {
		p.err = fmt.Errorf("found \"%s\", expected field", l.Lit)
		return nil
	}

	var next parseFunc
	if l := p.scan(); l.Token == COMMA {
		next = parseGroupByFields(p)
	} else {
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
	}

	return next
}

func parseWhere(p *Parser) parseFunc {
	l := p.scan()
	if l.Token != WHERE {
		p.err = fmt.Errorf("found \"%s\", expected WHERE", l.Lit)
		return nil
	}
	p.stmt.WhereClause = new(WhereClause)

	if p.peek().Token.IsTerminal() {
		return parseTerminalLexeme
	}
	return parseWherePredicates
}

func parseWherePredicates(p *Parser) parseFunc {
	panic("not implemented")
}

func parseOffset(p *Parser) parseFunc {
	if p.stmt.OffsetClause != nil {
		p.err = errors.New("OFFSET already defined in statement")
		return nil
	}

	l := p.scan()
	if l.Token != OFFSET {
		p.err = fmt.Errorf("found \"%s\", expected OFFSET", l.Lit)
		return nil
	}
	p.stmt.OffsetClause = new(OffsetClause)

	l = p.scan()
	if l.Token != INT {
		p.err = fmt.Errorf("found \"%s\", expected INT offset value", l.Lit)
		return nil
	}
	v, err := strconv.Atoi(l.Lit)
	if err != nil {
		p.err = fmt.Errorf("cannot parse offset, literal \"%s\" is not INT", l.Lit)
	}
	p.stmt.OffsetClause.Value = v

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
	if l.Token != EOF {
		p.err = fmt.Errorf("found \"%s\", expected EOF", l.Lit)
	}

	return nil
}
