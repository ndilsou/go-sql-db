package sql

type Token int

const (
	// Special tokens
	EOF Token = iota
	ILLEGAL
	WS

	misc_begin
	// Misc characters
	ASTERISK

	misc_end

	literal_begin
	// Identifiers
	COMMA
	FLOAT
	IDENT
	INT
	SEMICOLON
	STRING

	literal_end

	operator_begin
	// Comparison Operators and delimiters
	EQ  // =
	NEQ // <>
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=

	operator_end

	keyword_begin
	// Keywords
	BY
	DISTINCT
	FROM
	JOIN
	INNER
	OUTER
	LEFT
	GROUP
	HAVING
	LIMIT
	OFFSET
	ORDER
	SELECT
	WHERE

	keyword_end
)

var tokens = map[Token]string{
	ASTERISK:  "ASTERISK",
	COMMA:     "COMMA",
	DISTINCT:  "DISTINCT",
	EOF:       "EOF",
	FLOAT:     "FLOAT",
	FROM:      "FROM",
	GROUP:     "GROUP BY",
	HAVING:    "HAVING",
	IDENT:     "IDENT",
	ILLEGAL:   "ILLEGAL",
	INNER:     "INNER",
	INT:       "INT",
	JOIN:      "JOIN",
	LEFT:      "LEFT",
	LIMIT:     "LIMIT",
	OFFSET:    "OFFSET",
	ORDER:     "ORDER BY",
	OUTER:     "OUTER",
	SELECT:    "SELECT",
	SEMICOLON: "SEMICOLON",
	STRING:    "STRING",
	WHERE:     "WHERE",
	WS:        "WS",
	EQ:        "EQ",
	GT:        "GT",
	GTE:       "GTE",
	LT:        "LT",
	LTE:       "LTE",
	NEQ:       "NEQ",
}
var keywords = map[string]Token{
	"BY":       BY,
	"DISTINCT": DISTINCT,
	"FROM":     FROM,
	"GROUP":    GROUP,
	"HAVING":   HAVING,
	"INNER":    INNER,
	"JOIN":     JOIN,
	"LEFT":     LEFT,
	"LIMIT":    LIMIT,
	"OFFSET":   OFFSET,
	"ORDER":    ORDER,
	"OUTER":    OUTER,
	"SELECT":   SELECT,
	"WHERE":    WHERE,
}

func (t Token) String() string {
	str, ok := tokens[t]
	if !ok {
		return ""
	}
	return str
}

func (t Token) IsMisc() bool { return t >= misc_begin && t <= misc_end }

func (t Token) IsOperator() bool { return t >= operator_begin && t <= operator_end }

func (t Token) IsIdentifier() bool { return t >= literal_begin && t <= literal_end }

func (t Token) IsKeyword() bool { return t >= keyword_begin && t <= keyword_end }

func (t Token) IsTerminal() bool { return t == EOF || t == SEMICOLON }
