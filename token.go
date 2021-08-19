package sql

type Token int

const (
	// Special tokens
	EOF Token = iota + 1
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
	AS
	BY
	DISTINCT
	FROM
	GROUP
	HAVING
	INNER
	JOIN
	LEFT
	RIGHT
	FULL
	LIMIT
	OFFSET
	ON
	OR
	ORDER
	OUTER
	SELECT
	WHERE

	keyword_end
)

var tokens = map[Token]string{
	AND:       "AND",
	AS:        "AS",
	ASTERISK:  "ASTERISK",
	COMMA:     "COMMA",
	DISTINCT:  "DISTINCT",
	EOF:       "EOF",
	EQ:        "EQ",
	FLOAT:     "FLOAT",
	FROM:      "FROM",
	GROUP:     "GROUP BY",
	GT:        "GT",
	GTE:       "GTE",
	HAVING:    "HAVING",
	IDENT:     "IDENT",
	ILLEGAL:   "ILLEGAL",
	INNER:     "INNER",
	INT:       "INT",
	JOIN:      "JOIN",
	LEFT:      "LEFT",
	LIMIT:     "LIMIT",
	LT:        "LT",
	LTE:       "LTE",
	NEQ:       "NEQ",
	OFFSET:    "OFFSET",
	ON:        "ON",
	OR:        "OR",
	ORDER:     "ORDER BY",
	OUTER:     "OUTER",
	SELECT:    "SELECT",
	SEMICOLON: "SEMICOLON",
	STRING:    "STRING",
	WHERE:     "WHERE",
	WS:        "WS",
	RIGHT:     "RIGHT",
	FULL:      "FULL",
}
var keywords = map[string]Token{
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
	"ON":       ON,
	"OR":       OR,
	"ORDER":    ORDER,
	"OUTER":    OUTER,
	"SELECT":   SELECT,
	"WHERE":    WHERE,
	"RIGHT":    RIGHT,
	"FULL":     FULL,
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
