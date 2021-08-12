package sql

type Token int

const (
	// Special tokens
	EOF Token = iota
	ILLEGAL
	WS

	misc_begin
	// Misc characters
	COMMA
	SEMICOLON

	misc_end

	// Identifiers
	IDENT
	ASTERISK

	literal_begin
	FLOAT
	INT
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
	AND
	BY
	DISTINCT
	FROM
	GROUP
	HAVING
	INNER
	JOIN
	LEFT
	LIMIT
	OFFSET
	OR
	ORDER
	OUTER
	SELECT
	WHERE

	keyword_end
)

var tokens = map[Token]string{
	OR:        "OR",
	AND:       "AND",
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
	"OR":       OR,
	"AND":      AND,
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

func (t Token) IsComparisonOperator() bool { return t >= operator_begin && t <= operator_end }

func (t Token) IsLiteral() bool { return t >= literal_begin && t <= literal_end }

func (t Token) IsKeyword() bool { return t >= keyword_begin && t <= keyword_end }

func (t Token) IsTerminal() bool { return t == EOF || t == SEMICOLON }
