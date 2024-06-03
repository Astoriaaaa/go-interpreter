package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"let":    LET,
	"fn":     FUNC,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// looks up if the string is LET FUNC or an IDENTIFIER
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENTIFIER
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENTIFIER = "IDENTIFIER"
	INT        = "INT"

	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"
	EQ     = "=="
	NEQ    = "!="
	STAR   = "*"
	GR     = ">"
	LE     = "<"
	SLASH  = "/"
	EXCLA  = "!"

	COMMA     = ","
	SEMICOLON = ";"
	LP        = "("
	RP        = ")"
	LB        = "{"
	RB        = "}"

	LET    = "LET"
	FUNC   = "FUNCTION"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	RETURN = "RETURN"
	IF     = "IF"
	ELSE   = "ELSE"
	STRING = "STRING"

	LSB = "["
	RSB = "]"
)
