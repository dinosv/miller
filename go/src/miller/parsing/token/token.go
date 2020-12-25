// Code generated by gocc; DO NOT EDIT.

package token

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

type Token struct {
	Type
	Lit []byte
	Pos
}

type Type int

const (
	INVALID Type = iota
	EOF
)

type Pos struct {
	Offset int
	Line   int
	Column int
}

func (p Pos) String() string {
	return fmt.Sprintf("Pos(offset=%d, line=%d, column=%d)", p.Offset, p.Line, p.Column)
}

type TokenMap struct {
	typeMap []string
	idMap   map[string]Type
}

func (m TokenMap) Id(tok Type) string {
	if int(tok) < len(m.typeMap) {
		return m.typeMap[tok]
	}
	return "unknown"
}

func (m TokenMap) Type(tok string) Type {
	if typ, exist := m.idMap[tok]; exist {
		return typ
	}
	return INVALID
}

func (m TokenMap) TokenString(tok *Token) string {
	//TODO: refactor to print pos & token string properly
	return fmt.Sprintf("%s(%d,%s)", m.Id(tok.Type), tok.Type, tok.Lit)
}

func (m TokenMap) StringType(typ Type) string {
	return fmt.Sprintf("%s(%d)", m.Id(typ), typ)
}

// CharLiteralValue returns the string value of the char literal.
func (t *Token) CharLiteralValue() string {
	return string(t.Lit[1 : len(t.Lit)-1])
}

// Float32Value returns the float32 value of the token or an error if the token literal does not
// denote a valid float32.
func (t *Token) Float32Value() (float32, error) {
	if v, err := strconv.ParseFloat(string(t.Lit), 32); err != nil {
		return 0, err
	} else {
		return float32(v), nil
	}
}

// Float64Value returns the float64 value of the token or an error if the token literal does not
// denote a valid float64.
func (t *Token) Float64Value() (float64, error) {
	return strconv.ParseFloat(string(t.Lit), 64)
}

// IDValue returns the string representation of an identifier token.
func (t *Token) IDValue() string {
	return string(t.Lit)
}

// Int32Value returns the int32 value of the token or an error if the token literal does not
// denote a valid float64.
func (t *Token) Int32Value() (int32, error) {
	if v, err := strconv.ParseInt(string(t.Lit), 10, 64); err != nil {
		return 0, err
	} else {
		return int32(v), nil
	}
}

// Int64Value returns the int64 value of the token or an error if the token literal does not
// denote a valid float64.
func (t *Token) Int64Value() (int64, error) {
	return strconv.ParseInt(string(t.Lit), 10, 64)
}

// UTF8Rune decodes the UTF8 rune in the token literal. It returns utf8.RuneError if
// the token literal contains an invalid rune.
func (t *Token) UTF8Rune() (rune, error) {
	r, _ := utf8.DecodeRune(t.Lit)
	if r == utf8.RuneError {
		err := fmt.Errorf("Invalid rune")
		return r, err
	}
	return r, nil
}

// StringValue returns the string value of the token literal.
func (t *Token) StringValue() string {
	return string(t.Lit[1 : len(t.Lit)-1])
}

var TokMap = TokenMap{
	typeMap: []string{
		"INVALID",
		"$",
		"empty",
		";",
		"{",
		"}",
		"=",
		"unset",
		"filter",
		"emit",
		"emitp",
		"emitf",
		"dump",
		"edump",
		"print",
		"eprint",
		"printn",
		"eprintn",
		"field_name",
		"$[",
		"]",
		"braced_field_name",
		"$[[",
		"$[[[",
		"full_srec",
		"oosvar_name",
		"@[",
		"braced_oosvar_name",
		"full_oosvar",
		"non_sigil_name",
		"arr",
		"bool",
		"float",
		"int",
		"map",
		"num",
		"str",
		"var",
		"||=",
		"^^=",
		"&&=",
		"??=",
		"???=",
		"|=",
		"&=",
		"^=",
		"<<=",
		">>=",
		">>>=",
		"+=",
		".=",
		"-=",
		"*=",
		"/=",
		"//=",
		"%=",
		"**=",
		"?",
		":",
		"||",
		"^^",
		"&&",
		"??",
		"???",
		"=~",
		"!=~",
		"==",
		"!=",
		">",
		">=",
		"<",
		"<=",
		"|",
		"^",
		"&",
		"<<",
		">>",
		">>>",
		"+",
		"-",
		".+",
		".-",
		".",
		"*",
		"/",
		"//",
		"%",
		".*",
		"./",
		".//",
		"!",
		"~",
		"**",
		"(",
		")",
		"string_literal",
		"int_literal",
		"float_literal",
		"boolean_literal",
		"const_M_PI",
		"const_M_E",
		"panic",
		"[",
		",",
		"ctx_IPS",
		"ctx_IFS",
		"ctx_IRS",
		"ctx_OPS",
		"ctx_OFS",
		"ctx_ORS",
		"ctx_NF",
		"ctx_NR",
		"ctx_FNR",
		"ctx_FILENAME",
		"ctx_FILENUM",
		"env",
		"[[",
		"[[[",
		"call",
		"begin",
		"end",
		"if",
		"elif",
		"else",
		"while",
		"do",
		"for",
		"in",
		"break",
		"continue",
		"func",
		"subr",
		"return",
	},

	idMap: map[string]Type{
		"INVALID":            0,
		"$":                  1,
		"empty":              2,
		";":                  3,
		"{":                  4,
		"}":                  5,
		"=":                  6,
		"unset":              7,
		"filter":             8,
		"emit":               9,
		"emitp":              10,
		"emitf":              11,
		"dump":               12,
		"edump":              13,
		"print":              14,
		"eprint":             15,
		"printn":             16,
		"eprintn":            17,
		"field_name":         18,
		"$[":                 19,
		"]":                  20,
		"braced_field_name":  21,
		"$[[":                22,
		"$[[[":               23,
		"full_srec":          24,
		"oosvar_name":        25,
		"@[":                 26,
		"braced_oosvar_name": 27,
		"full_oosvar":        28,
		"non_sigil_name":     29,
		"arr":                30,
		"bool":               31,
		"float":              32,
		"int":                33,
		"map":                34,
		"num":                35,
		"str":                36,
		"var":                37,
		"||=":                38,
		"^^=":                39,
		"&&=":                40,
		"??=":                41,
		"???=":               42,
		"|=":                 43,
		"&=":                 44,
		"^=":                 45,
		"<<=":                46,
		">>=":                47,
		">>>=":               48,
		"+=":                 49,
		".=":                 50,
		"-=":                 51,
		"*=":                 52,
		"/=":                 53,
		"//=":                54,
		"%=":                 55,
		"**=":                56,
		"?":                  57,
		":":                  58,
		"||":                 59,
		"^^":                 60,
		"&&":                 61,
		"??":                 62,
		"???":                63,
		"=~":                 64,
		"!=~":                65,
		"==":                 66,
		"!=":                 67,
		">":                  68,
		">=":                 69,
		"<":                  70,
		"<=":                 71,
		"|":                  72,
		"^":                  73,
		"&":                  74,
		"<<":                 75,
		">>":                 76,
		">>>":                77,
		"+":                  78,
		"-":                  79,
		".+":                 80,
		".-":                 81,
		".":                  82,
		"*":                  83,
		"/":                  84,
		"//":                 85,
		"%":                  86,
		".*":                 87,
		"./":                 88,
		".//":                89,
		"!":                  90,
		"~":                  91,
		"**":                 92,
		"(":                  93,
		")":                  94,
		"string_literal":     95,
		"int_literal":        96,
		"float_literal":      97,
		"boolean_literal":    98,
		"const_M_PI":         99,
		"const_M_E":          100,
		"panic":              101,
		"[":                  102,
		",":                  103,
		"ctx_IPS":            104,
		"ctx_IFS":            105,
		"ctx_IRS":            106,
		"ctx_OPS":            107,
		"ctx_OFS":            108,
		"ctx_ORS":            109,
		"ctx_NF":             110,
		"ctx_NR":             111,
		"ctx_FNR":            112,
		"ctx_FILENAME":       113,
		"ctx_FILENUM":        114,
		"env":                115,
		"[[":                 116,
		"[[[":                117,
		"call":               118,
		"begin":              119,
		"end":                120,
		"if":                 121,
		"elif":               122,
		"else":               123,
		"while":              124,
		"do":                 125,
		"for":                126,
		"in":                 127,
		"break":              128,
		"continue":           129,
		"func":               130,
		"subr":               131,
		"return":             132,
	},
}
