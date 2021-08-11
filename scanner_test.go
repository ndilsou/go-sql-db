package sql_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ndilsou/go-rdbms-playground"
)

// Ensure the scanner can scan tokens correctly.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s    string
		item sql.Lexeme
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, item: sql.Lexeme{Token: sql.EOF}},
		{s: `#`, item: sql.Lexeme{Token: sql.ILLEGAL, Lit: `#`}},
		{s: ` `, item: sql.Lexeme{Token: sql.WS, Lit: " "}},
		{s: `,`, item: sql.Lexeme{Token: sql.COMMA, Lit: ","}},
		{s: "\t", item: sql.Lexeme{Token: sql.EOF, Lit: ""}},
		{s: "\n", item: sql.Lexeme{Token: sql.EOF, Lit: ""}},
		{s: " \t\n    ", item: sql.Lexeme{Token: sql.EOF, Lit: ""}},

		// Misc characters
		{s: `*`, item: sql.Lexeme{Token: sql.ASTERISK, Lit: "*"}},

		// Identifiers
		{s: `foo`, item: sql.Lexeme{Token: sql.IDENT, Lit: `foo`}},
		{s: `foo.bar.baz`, item: sql.Lexeme{Token: sql.IDENT, Lit: `foo.bar.baz`}},
		{s: `Zx12_3U_-`, item: sql.Lexeme{Token: sql.IDENT, Lit: `Zx12_3U_`}},

		// String Literals
		{s: `"yolo"`, item: sql.Lexeme{Token: sql.STRING, Lit: `"yolo"`}},
		{s: `'yolo'`, item: sql.Lexeme{Token: sql.STRING, Lit: `'yolo'`}},
		{s: `'this is a test'`, item: sql.Lexeme{Token: sql.STRING, Lit: `'this is a test'`}},

		// Comparison Operators
		{s: `=`, item: sql.Lexeme{Token: sql.EQ, Lit: `=`}},
		{s: `>=`, item: sql.Lexeme{Token: sql.GTE, Lit: `>=`}},
		{s: `<=`, item: sql.Lexeme{Token: sql.LTE, Lit: `<=`}},
		{s: `<>`, item: sql.Lexeme{Token: sql.NEQ, Lit: `<>`}},

		// Numerics
		{s: `1`, item: sql.Lexeme{Token: sql.INT, Lit: `1`}},
		{s: `0.5`, item: sql.Lexeme{Token: sql.FLOAT, Lit: `0.5`}},
		{s: `-1.1`, item: sql.Lexeme{Token: sql.FLOAT, Lit: `-1.1`}},
		{s: `+9`, item: sql.Lexeme{Token: sql.INT, Lit: `+9`}},

		// Keywords
		{s: `FROM`, item: sql.Lexeme{Token: sql.FROM, Lit: "FROM"}},
		{s: `INNER`, item: sql.Lexeme{Token: sql.INNER, Lit: "INNER"}},
		{s: `OUTER`, item: sql.Lexeme{Token: sql.OUTER, Lit: "OUTER"}},
		{s: `LEFT`, item: sql.Lexeme{Token: sql.LEFT, Lit: "LEFT"}},
		{s: `JOIN`, item: sql.Lexeme{Token: sql.JOIN, Lit: "JOIN"}},
		{s: `BY`, item: sql.Lexeme{Token: sql.BY, Lit: "BY"}},
		{s: `GROUP`, item: sql.Lexeme{Token: sql.GROUP, Lit: "GROUP"}},
		{s: `ORDER`, item: sql.Lexeme{Token: sql.ORDER, Lit: "ORDER"}},
		{s: `SELECT`, item: sql.Lexeme{Token: sql.SELECT, Lit: "SELECT"}},
		{s: `HAVING`, item: sql.Lexeme{Token: sql.HAVING, Lit: "HAVING"}},
		{s: `WHERE`, item: sql.Lexeme{Token: sql.WHERE, Lit: "WHERE"}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			s := sql.NewScanner(strings.NewReader(tt.s))
			it := s.Scan()
			if tt.item.Token != it.Token {
				t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tt.s, tt.item.Token, it.Token, it.Lit)
			} else if tt.item.Lit != it.Lit {
				t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s, tt.item.Lit, it.Lit)
			}
		})
	}
}

func TestScanner_Scan_Sequence(t *testing.T) {
	var tests = []struct {
		s     string
		items []sql.Lexeme
	}{
		{
			s: "SELECT *\nFROM yolo",
			items: []sql.Lexeme{
				{Token: sql.SELECT, Lit: "SELECT"},
				{Token: sql.ASTERISK, Lit: "*"},
				{Token: sql.FROM, Lit: "FROM"},
				{Token: sql.IDENT, Lit: "yolo"},
				{Token: sql.EOF, Lit: ""},
			},
		},
		{
			s: "SELECT *\nFROM yolo WHERE iam",
			items: []sql.Lexeme{
				{Token: sql.SELECT, Lit: "SELECT"},
				{Token: sql.ASTERISK, Lit: "*"},
				{Token: sql.FROM, Lit: "FROM"},
				{Token: sql.IDENT, Lit: "yolo"},
				{Token: sql.WHERE, Lit: "WHERE"},
				{Token: sql.IDENT, Lit: "iam"},
				{Token: sql.EOF, Lit: ""},
			},
		},
		{
			s:     "",
			items: []sql.Lexeme{{Token: sql.EOF, Lit: ""}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			s := sql.NewScanner(strings.NewReader(tt.s))
			var items []sql.Lexeme
			for {
				item := s.Scan()
				items = append(items, item)
				if item.Token == sql.EOF {
					break
				}
			}
			if !reflect.DeepEqual(tt.items, items) {
				t.Errorf("%d. %q tokens mismatch: exp=%q got=%q", i, tt.s, tt.items, items)
			}
		})
	}
}
