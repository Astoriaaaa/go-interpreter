package parser

import (
	"fmt"
	"interpreter/ast"
	"interpreter/lexer"
	"interpreter/token"
	"strconv"
)

type (
	prefixParseFns func() ast.Expression
	infixParseFns  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX
)

var precedences = map[token.TokenType]int{
	token.EQ:    EQUALS,
	token.NEQ:   EQUALS,
	token.LE:    LESSGREATER,
	token.GR:    LESSGREATER,
	token.PLUS:  SUM,
	token.MINUS: SUM,
	token.SLASH: PRODUCT,
	token.STAR:  PRODUCT,
	token.LP:    CALL,
	token.LSB:   INDEX,
}

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peakToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFns
	infixParseFns  map[token.TokenType]infixParseFns
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFns)
	p.infixParseFns = make(map[token.TokenType]infixParseFns)
	p.registerPrefix(token.LB, p.parseHashExpression)
	p.registerPrefix(token.LSB, p.parseArrayExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.FUNC, p.parseFunction)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.LP, p.parseGroupExpressions)
	p.registerPrefix(token.IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.EXCLA, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerInfix(token.LP, p.parseCallExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.STAR, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LE, p.parseInfixExpression)
	p.registerInfix(token.GR, p.parseInfixExpression)
	p.registerInfix(token.LSB, p.parseIndexExpression)

	return p

}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpreesionStatement()
	}
}

func (p *Parser) parseExpreesionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peakToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RP)
	return exp
}

func (p *Parser) parseHashExpression() ast.Expression {
	hash := &ast.HashExpression{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RB) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		val := p.parseExpression(LOWEST)
		hash.Pairs[key] = val

		if !p.peekTokenIs(token.RB) && !p.expectPeek(token.COMMA) {
			return nil
		}

	}

	if !p.expectPeek(token.RB) {
		return nil
	}

	return hash

}

func (p *Parser) parseExpressionList(expect token.TokenType) []ast.Expression {
	expressions := []ast.Expression{}
	if p.peekTokenIs(expect) {
		p.nextToken()
		return expressions
	}
	p.nextToken()
	expressions = append(expressions, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		expressions = append(expressions, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(expect) {
		return nil
	}
	return expressions

}

func (p *Parser) parseFunction() ast.Expression {
	exp := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LP) {
		return nil
	}
	exp.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LB) {
		return nil
	}
	exp.Body = p.parseBlockStatement()
	return exp
}
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	idents := []*ast.Identifier{}
	p.nextToken()
	if p.curTokenIs(token.RP) {
		return idents
	}

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	idents = append(idents, ident)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		idents = append(idents, ident)
	}

	if !p.expectPeek(token.RP) {
		return nil
	}
	return idents
}

func (p *Parser) parseIfExpression() ast.Expression {
	stmt := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LP) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RP) {
		return nil
	}
	if !p.expectPeek(token.LB) {
		return nil
	}
	stmt.Consequence = p.parseBlockStatement()
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LB) {
			return nil
		}
		stmt.Alternatives = p.parseBlockStatement()

	}
	return stmt

}
func (p *Parser) parseBlockStatement() *ast.BlockStatements {
	block := &ast.BlockStatements{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RB) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseGroupExpressions() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RP) {
		return nil
	}
	return exp
}

func (p *Parser) parseBoolean() ast.Expression {
	stmt := &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
	return stmt
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit

}

func (p *Parser) parseIdentifier() ast.Expression {
	stmt := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return stmt
}

func (p *Parser) parseReturnStatement() ast.Statement {
	r := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	r.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return r
}

func (p *Parser) parseLetStatement() ast.Statement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayExpression() ast.Expression {
	array := &ast.Array{Token: p.curToken}
	array.Items = p.parseExpressionList(token.RSB)
	return array

}

func (p *Parser) parseIndexExpression(leftExp ast.Expression) ast.Expression {
	p.nextToken()
	exp := &ast.IndexExpression{Token: p.curToken, LeftExpression: leftExp}
	index := p.parseExpression(LOWEST)
	exp.Index = index
	if !p.expectPeek(token.RSB) {
		return nil
	}
	return exp
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peakToken.Type == t
}
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.PeekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) PeekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peakToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peakToken
	p.peakToken = p.l.NextToken()
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFns) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFns) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peakToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
