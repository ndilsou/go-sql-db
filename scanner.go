package sql

import (
	"bufio"
	"strconv"
	"strings"
	"unicode"
)

var eof = rune(0)
var bufSizeHint = 32

type Lexeme struct {
	Token Token
	Lit   string
}

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r *strings.Reader) *Scanner {
	s := Scanner{r: bufio.NewReader(r)}
	return &s
}

func (s *Scanner) Scan() Lexeme {
	var ch rune
	ch = s.read()
	var lex Lexeme
	switch {
	case isWhitespace(ch):
		s.unread()
		s.skipWhitespaces()
		lex = s.Scan()
	case unicode.IsLetter(ch):
		s.unread()
		lit := s.scanLiterals()
		tok := tokenizeLiteral(lit)
		lex = Lexeme{tok, lit}
	case isComparisonOperator(ch):
		s.unread()
		lit := s.scanOperators()
		tok := tokenizeOperators(lit)
		lex = Lexeme{tok, lit}
	case unicode.IsDigit(ch) || isSign(ch):
		s.unread()
		lit := s.scanNumerics()
		tok := tokenizeNumerics(lit)
		lex = Lexeme{tok, lit}
	case isQuotationMark(ch):
		s.unread()
		lit := s.scanStringLiterals()
		lex = Lexeme{STRING, lit}
	case ch == eof:
		lex = Lexeme{EOF, ""}
	case ch == '*':
		lex = Lexeme{ASTERISK, "*"}
	case ch == ',':
		lex = Lexeme{COMMA, ","}
	case ch == ';':
		lex = Lexeme{SEMICOLON, ";"}
	default:
		lex = Lexeme{ILLEGAL, string(ch)}
	}
	return lex
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) peek() rune {
	ch := s.read()
	s.unread()
	return ch
}

func (s *Scanner) scanLiterals() string {
	var sb strings.Builder
	sb.Grow(bufSizeHint)
	for {
		ch := s.read()
		if !isAlphanumeric(ch) {
			// Special cases:
			// - underscore in identifiers names are acceptable
			// - dot for table qualified name, ex. db.schema.table
			if (ch == '_') || (ch == '.' && s.peek() != '.') {
				sb.WriteRune(ch)
				continue
			}
			s.unread()
			return sb.String()
		}

		sb.WriteRune(ch)
	}
}

func (s *Scanner) scanWhitespaces() Lexeme {
	var sb strings.Builder
	sb.Grow(bufSizeHint)
	for {
		ch := s.read()
		if !isWhitespace(ch) {
			s.unread()
			return Lexeme{WS, sb.String()}
		}

		sb.WriteRune(ch)
	}
}

func (s *Scanner) scanNumerics() string {
	var sb strings.Builder
	sb.Grow(bufSizeHint)

	ch := s.read()
	sb.WriteRune(ch)

	for {
		ch := s.read()
		if !unicode.IsDigit(ch) {
			// Special cases:
			// SI notation exponent
			// floats dot decimal separator
			if (unicode.ToUpper(ch) == 'E') || (ch == '.' && s.peek() != '.') {
				sb.WriteRune(ch)
				continue
			}
			s.unread()

			return sb.String()
		}

		sb.WriteRune(ch)
	}
}

func (s *Scanner) scanStringLiterals() string {
	var sb strings.Builder
	sb.Grow(bufSizeHint)
	ch := s.read()
	sb.WriteRune(ch)

	for {
		ch := s.read()
		sb.WriteRune(ch)
		if isQuotationMark(ch) {
			return sb.String()
		}
	}
}

func (s *Scanner) scanOperators() string {
	var sb strings.Builder
	sb.Grow(bufSizeHint)
	ch1 := s.read()
	sb.WriteRune(ch1)

	ch2 := s.read()
	if (ch1 == '<' && (ch2 == '>' || ch2 == '=')) || (ch1 == '>' && ch2 == '=') {
		sb.WriteRune(ch2)
	} else {
		s.unread()
	}

	return sb.String()
}

func (s *Scanner) skipWhitespaces() {
	for {
		ch := s.read()
		if !isWhitespace(ch) {
			s.unread()
			return
		}
	}
}

func isWhitespace(ch rune) bool {
	return unicode.Is(unicode.White_Space, ch)
}

func isQuotationMark(ch rune) bool {
	return unicode.Is(unicode.Quotation_Mark, ch)
}

func isAlphanumeric(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

func isSign(ch rune) bool {
	return ch == '-' || ch == '+'
}

func isComparisonOperator(ch rune) bool {
	return ch == '=' || ch == '>' || ch == '<'

}

func tokenizeLiteral(lit string) Token {
	if kw, ok := keywords[strings.ToUpper(lit)]; ok {
		return kw
	}
	return IDENT
}

func tokenizeOperators(lit string) Token {
	var tok Token
	switch lit {
	case "=":
		tok = EQ
	case "<>":
		tok = NEQ
	case ">":
		tok = GT
	case ">=":
		tok = GTE
	case "<":
		tok = LT
	case "<=":
		tok = LTE
	default:
		tok = ILLEGAL
	}
	return tok
}

func tokenizeNumerics(lit string) Token {
	if _, err := strconv.ParseInt(lit, 10, 64); err == nil {
		return INT
	}
	if _, err := strconv.ParseFloat(lit, 64); err == nil {
		return FLOAT
	}
	return ILLEGAL
}
