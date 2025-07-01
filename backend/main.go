package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	
	"strings"
)

// Estructuras para los análisis
type Token struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Position int    `json:"position"`
}

type SyntaxNode struct {
	Type     string        `json:"type"`
	Value    string        `json:"value"`
	Children []*SyntaxNode `json:"children,omitempty"`
	Line     int           `json:"line"`
}

type SemanticError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Line        int    `json:"line"`
	Severity    string `json:"severity"`
	Variable    string `json:"variable,omitempty"`
	ExpectedType string `json:"expectedType,omitempty"`
	ActualType   string `json:"actualType,omitempty"`
}

type AnalysisResult struct {
	Tokens         []Token         `json:"tokens"`
	SyntaxTree     []*SyntaxNode   `json:"syntaxTree"`
	SemanticErrors []SemanticError `json:"semanticErrors"`
	SymbolTable    []Symbol        `json:"symbolTable"`
}

type Symbol struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Line     int    `json:"line"`
	Scope    string `json:"scope"`
	Used     bool   `json:"used"`
}

type Analyzer struct {
	code         string
	tokens       []Token
	current      int
	line         int
	column       int
	symbols      map[string]Symbol
	errors       []SemanticError
}

func NewAnalyzer(code string) *Analyzer {
	return &Analyzer{
		code:    code,
		tokens:  []Token{},
		current: 0,
		line:    1,
		column:  1,
		symbols: make(map[string]Symbol),
		errors:  []SemanticError{},
	}
}

// Analizador Léxico
func (a *Analyzer) lexicalAnalysis() {
	keywords := map[string]bool{
		"def": true, "if": true, "else": true, "elif": true,
		"while": true, "for": true, "in": true, "return": true,
		"print": true, "True": true, "False": true, "None": true,
		"and": true, "or": true, "not": true, "import": true,
		"from": true, "class": true, "try": true, "except": true,
	}

	operators := []string{
		"==", "!=", "<=", ">=", "<", ">", "=", "+", "-", "*", "/", "%",
	}

	delimiters := []string{
		"(", ")", "[", "]", "{", "}", ",", ":", ";", ".", "\"", "'",
	}

	i := 0
	for i < len(a.code) {
		startLine := a.line
		startColumn := a.column
		startPos := i

		// Espacios en blanco y saltos de línea
		if a.code[i] == ' ' || a.code[i] == '\t' {
			i++
			a.column++
			continue
		}

		if a.code[i] == '\n' {
			a.addToken("NEWLINE", "\\n", startLine, startColumn, startPos)
			i++
			a.line++
			a.column = 1
			continue
		}

		// Comentarios
		if a.code[i] == '#' {
			start := i
			for i < len(a.code) && a.code[i] != '\n' {
				i++
			}
			a.addToken("COMMENT", a.code[start:i], startLine, startColumn, startPos)
			continue
		}

		// Strings
		if a.code[i] == '"' || a.code[i] == '\'' {
			quote := a.code[i]
			start := i
			i++
			a.column++
			for i < len(a.code) && a.code[i] != quote {
				if a.code[i] == '\n' {
					a.line++
					a.column = 1
				} else {
					a.column++
				}
				i++
			}
			if i < len(a.code) {
				i++
				a.column++
			}
			a.addToken("STRING", a.code[start:i], startLine, startColumn, startPos)
			continue
		}

		// Números
		if a.isDigit(a.code[i]) {
			start := i
			for i < len(a.code) && (a.isDigit(a.code[i]) || a.code[i] == '.') {
				i++
				a.column++
			}
			a.addToken("NUMBER", a.code[start:i], startLine, startColumn, startPos)
			continue
		}

		// Operadores de múltiples caracteres
		found := false
		for _, op := range operators {
			if i+len(op) <= len(a.code) && a.code[i:i+len(op)] == op {
				a.addToken("OPERATOR", op, startLine, startColumn, startPos)
				i += len(op)
				a.column += len(op)
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Delimitadores
		found = false
		for _, del := range delimiters {
			if i+len(del) <= len(a.code) && a.code[i:i+len(del)] == del {
				a.addToken("DELIMITER", del, startLine, startColumn, startPos)
				i += len(del)
				a.column += len(del)
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Identificadores y palabras clave
		if a.isLetter(a.code[i]) || a.code[i] == '_' {
			start := i
			for i < len(a.code) && (a.isAlphanumeric(a.code[i]) || a.code[i] == '_') {
				i++
				a.column++
			}
			word := a.code[start:i]
			tokenType := "IDENTIFIER"
			if keywords[word] {
				tokenType = "KEYWORD"
			}
			a.addToken(tokenType, word, startLine, startColumn, startPos)
			continue
		}

		// Carácter desconocido
		a.addToken("UNKNOWN", string(a.code[i]), startLine, startColumn, startPos)
		i++
		a.column++
	}
}

func (a *Analyzer) addToken(tokenType, value string, line, column, position int) {
	a.tokens = append(a.tokens, Token{
		Type:     tokenType,
		Value:    value,
		Line:     line,
		Column:   column,
		Position: position,
	})
}

func (a *Analyzer) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (a *Analyzer) isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (a *Analyzer) isAlphanumeric(c byte) bool {
	return a.isLetter(c) || a.isDigit(c)
}

// Analizador Sintáctico
func (a *Analyzer) syntaxAnalysis() []*SyntaxNode {
	a.current = 0
	var statements []*SyntaxNode

	for a.current < len(a.tokens) {
		if a.tokens[a.current].Type == "NEWLINE" {
			a.current++
			continue
		}
		stmt := a.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements
}

func (a *Analyzer) parseStatement() *SyntaxNode {
	if a.current >= len(a.tokens) {
		return nil
	}

	token := a.tokens[a.current]

	switch token.Value {
	case "def":
		return a.parseFunctionDef()
	case "if":
		return a.parseIfStatement()
	default:
		if a.current+1 < len(a.tokens) && a.tokens[a.current+1].Value == "=" {
			return a.parseAssignment()
		}
		return a.parseExpressionStatement()
	}
}

func (a *Analyzer) parseFunctionDef() *SyntaxNode {
	node := &SyntaxNode{Type: "FUNCTION_DEF", Line: a.tokens[a.current].Line}
	a.current++ // consumir 'def'

	if a.current < len(a.tokens) && a.tokens[a.current].Type == "IDENTIFIER" {
		node.Value = a.tokens[a.current].Value
		a.current++
	}

	// Consumir '('
	if a.current < len(a.tokens) && a.tokens[a.current].Value == "(" {
		a.current++
	}

	// Consumir ')'
	if a.current < len(a.tokens) && a.tokens[a.current].Value == ")" {
		a.current++
	}

	// Consumir ':'
	if a.current < len(a.tokens) && a.tokens[a.current].Value == ":" {
		a.current++
	}

	// Consumir salto de línea
	if a.current < len(a.tokens) && a.tokens[a.current].Type == "NEWLINE" {
		a.current++
	}

	// Parsear cuerpo de la función
	for a.current < len(a.tokens) && a.tokens[a.current].Type != "KEYWORD" {
		if a.tokens[a.current].Type == "NEWLINE" {
			a.current++
			continue
		}
		stmt := a.parseStatement()
		if stmt != nil {
			node.Children = append(node.Children, stmt)
		}
	}

	return node
}

func (a *Analyzer) parseIfStatement() *SyntaxNode {
	node := &SyntaxNode{Type: "IF_STATEMENT", Line: a.tokens[a.current].Line}
	a.current++ // consumir 'if'

	// Parsear condición
	condition := a.parseExpression()
	if condition != nil {
		node.Children = append(node.Children, condition)
	}

	// Consumir ':'
	if a.current < len(a.tokens) && a.tokens[a.current].Value == ":" {
		a.current++
	}

	// Consumir salto de línea
	if a.current < len(a.tokens) && a.tokens[a.current].Type == "NEWLINE" {
		a.current++
	}

	// Parsear cuerpo del if
	for a.current < len(a.tokens) && 
		a.tokens[a.current].Type != "KEYWORD" && 
		a.tokens[a.current].Value != "if" {
		if a.tokens[a.current].Type == "NEWLINE" {
			a.current++
			continue
		}
		stmt := a.parseStatement()
		if stmt != nil {
			node.Children = append(node.Children, stmt)
		}
	}

	return node
}

func (a *Analyzer) parseAssignment() *SyntaxNode {
	node := &SyntaxNode{Type: "ASSIGNMENT", Line: a.tokens[a.current].Line}
	
	// Variable
	if a.current < len(a.tokens) && a.tokens[a.current].Type == "IDENTIFIER" {
		varNode := &SyntaxNode{
			Type:  "IDENTIFIER",
			Value: a.tokens[a.current].Value,
			Line:  a.tokens[a.current].Line,
		}
		node.Children = append(node.Children, varNode)
		a.current++
	}

	// Operador '='
	if a.current < len(a.tokens) && a.tokens[a.current].Value == "=" {
		a.current++
	}

	// Expresión
	expr := a.parseExpression()
	if expr != nil {
		node.Children = append(node.Children, expr)
	}

	return node
}

func (a *Analyzer) parseExpressionStatement() *SyntaxNode {
	expr := a.parseExpression()
	if expr != nil {
		return &SyntaxNode{
			Type:     "EXPRESSION_STATEMENT",
			Children: []*SyntaxNode{expr},
			Line:     expr.Line,
		}
	}
	return nil
}

func (a *Analyzer) parseExpression() *SyntaxNode {
	if a.current >= len(a.tokens) {
		return nil
	}

	token := a.tokens[a.current]

	switch token.Type {
	case "NUMBER":
		a.current++
		return &SyntaxNode{Type: "NUMBER", Value: token.Value, Line: token.Line}
	case "STRING":
		a.current++
		return &SyntaxNode{Type: "STRING", Value: token.Value, Line: token.Line}
	case "IDENTIFIER":
		a.current++
		node := &SyntaxNode{Type: "IDENTIFIER", Value: token.Value, Line: token.Line}
		
		// Verificar si es una llamada a función o método
		if a.current < len(a.tokens) && a.tokens[a.current].Value == "(" {
			return a.parseFunctionCall(node)
		} else if a.current < len(a.tokens) && a.tokens[a.current].Value == "." {
			return a.parseMethodCall(node)
		}
		
		return node
	case "KEYWORD":
		if token.Value == "print" {
			return a.parsePrintStatement()
		}
		a.current++
		return &SyntaxNode{Type: "KEYWORD", Value: token.Value, Line: token.Line}
	default:
		// Manejar operadores y expresiones complejas
		if a.current+2 < len(a.tokens) && a.tokens[a.current+1].Type == "OPERATOR" {
			return a.parseBinaryExpression()
		}
		a.current++
		return &SyntaxNode{Type: token.Type, Value: token.Value, Line: token.Line}
	}
}

func (a *Analyzer) parseFunctionCall(funcNode *SyntaxNode) *SyntaxNode {
	node := &SyntaxNode{Type: "FUNCTION_CALL", Value: funcNode.Value, Line: funcNode.Line}
	a.current++ // consumir '('

	// Parsear argumentos
	for a.current < len(a.tokens) && a.tokens[a.current].Value != ")" {
		if a.tokens[a.current].Value == "," {
			a.current++
			continue
		}
		arg := a.parseExpression()
		if arg != nil {
			node.Children = append(node.Children, arg)
		}
	}

	if a.current < len(a.tokens) && a.tokens[a.current].Value == ")" {
		a.current++
	}

	return node
}

func (a *Analyzer) parseMethodCall(objNode *SyntaxNode) *SyntaxNode {
	node := &SyntaxNode{Type: "METHOD_CALL", Line: objNode.Line}
	node.Children = append(node.Children, objNode)
	
	a.current++ // consumir '.'
	
	if a.current < len(a.tokens) && a.tokens[a.current].Type == "IDENTIFIER" {
		methodNode := &SyntaxNode{
			Type:  "IDENTIFIER",
			Value: a.tokens[a.current].Value,
			Line:  a.tokens[a.current].Line,
		}
		node.Children = append(node.Children, methodNode)
		a.current++
		
		// Si hay paréntesis, parsear argumentos
		if a.current < len(a.tokens) && a.tokens[a.current].Value == "(" {
			a.current++
			for a.current < len(a.tokens) && a.tokens[a.current].Value != ")" {
				if a.tokens[a.current].Value == "," {
					a.current++
					continue
				}
				arg := a.parseExpression()
				if arg != nil {
					node.Children = append(node.Children, arg)
				}
			}
			if a.current < len(a.tokens) && a.tokens[a.current].Value == ")" {
				a.current++
			}
		}
	}
	
	return node
}

func (a *Analyzer) parsePrintStatement() *SyntaxNode {
	node := &SyntaxNode{Type: "PRINT_STATEMENT", Line: a.tokens[a.current].Line}
	a.current++ // consumir 'print'

	if a.current < len(a.tokens) && a.tokens[a.current].Value == "(" {
		a.current++
		for a.current < len(a.tokens) && a.tokens[a.current].Value != ")" {
			if a.tokens[a.current].Value == "," {
				a.current++
				continue
			}
			arg := a.parseExpression()
			if arg != nil {
				node.Children = append(node.Children, arg)
			}
		}
		if a.current < len(a.tokens) && a.tokens[a.current].Value == ")" {
			a.current++
		}
	}

	return node
}

func (a *Analyzer) parseBinaryExpression() *SyntaxNode {
	left := &SyntaxNode{
		Type:  a.tokens[a.current].Type,
		Value: a.tokens[a.current].Value,
		Line:  a.tokens[a.current].Line,
	}
	a.current++

	operator := &SyntaxNode{
		Type:  "OPERATOR",
		Value: a.tokens[a.current].Value,
		Line:  a.tokens[a.current].Line,
	}
	a.current++

	right := a.parseExpression()

	node := &SyntaxNode{
		Type:     "BINARY_EXPRESSION",
		Children: []*SyntaxNode{left, operator, right},
		Line:     left.Line,
	}

	return node
}

// Analizador Semántico
func (a *Analyzer) semanticAnalysis(syntaxTree []*SyntaxNode) {
	for _, node := range syntaxTree {
		a.analyzeNode(node)
	}
}

func (a *Analyzer) analyzeNode(node *SyntaxNode) {
	if node == nil {
		return
	}

	switch node.Type {
	case "ASSIGNMENT":
		a.analyzeAssignment(node)
	case "IDENTIFIER":
		a.analyzeIdentifier(node)
	case "FUNCTION_CALL":
		a.analyzeFunctionCall(node)
	case "BINARY_EXPRESSION":
		a.analyzeBinaryExpression(node)
	}

	for _, child := range node.Children {
		a.analyzeNode(child)
	}
}

func (a *Analyzer) analyzeAssignment(node *SyntaxNode) {
	if len(node.Children) >= 2 {
		varNode := node.Children[0]
		valueNode := node.Children[1]

		varName := varNode.Value
		varType := a.inferType(valueNode)

		// Agregar al tabla de símbolos
		a.symbols[varName] = Symbol{
			Name:  varName,
			Type:  varType,
			Value: a.getNodeValue(valueNode),
			Line:  node.Line,
			Scope: "global",
			Used:  false,
		}
	}
}

func (a *Analyzer) analyzeIdentifier(node *SyntaxNode) {
	varName := node.Value
	if symbol, exists := a.symbols[varName]; exists {
		// Marcar como usado
		symbol.Used = true
		a.symbols[varName] = symbol
	} else {
		// Variable no declarada
		a.errors = append(a.errors, SemanticError{
			Type:     "UNDEFINED_VARIABLE",
			Message:  fmt.Sprintf("Variable '%s' no está definida", varName),
			Line:     node.Line,
			Severity: "error",
			Variable: varName,
		})
	}
}

func (a *Analyzer) analyzeFunctionCall(node *SyntaxNode) {
	funcName := node.Value
	
	// Verificar funciones built-in
	builtinFunctions := map[string]bool{
		"print": true,
		"len":   true,
		"str":   true,
		"int":   true,
		"float": true,
	}

	if !builtinFunctions[funcName] {
		if _, exists := a.symbols[funcName]; !exists {
			a.errors = append(a.errors, SemanticError{
				Type:     "UNDEFINED_FUNCTION",
				Message:  fmt.Sprintf("Función '%s' no está definida", funcName),
				Line:     node.Line,
				Severity: "error",
				Variable: funcName,
			})
		}
	}
}

func (a *Analyzer) analyzeBinaryExpression(node *SyntaxNode) {
	if len(node.Children) >= 3 {
		left := node.Children[0]
		operator := node.Children[1]
		right := node.Children[2]

		leftType := a.inferType(left)
		rightType := a.inferType(right)

		// Verificar compatibilidad de tipos
		if !a.areTypesCompatible(leftType, rightType, operator.Value) {
			a.errors = append(a.errors, SemanticError{
				Type:         "TYPE_MISMATCH",
				Message:      fmt.Sprintf("Tipos incompatibles: %s %s %s", leftType, operator.Value, rightType),
				Line:         node.Line,
				Severity:     "error",
				ExpectedType: leftType,
				ActualType:   rightType,
			})
		}
	}
}

func (a *Analyzer) inferType(node *SyntaxNode) string {
	if node == nil {
		return "unknown"
	}

	switch node.Type {
	case "NUMBER":
		if strings.Contains(node.Value, ".") {
			return "float"
		}
		return "int"
	case "STRING":
		return "string"
	case "IDENTIFIER":
		if symbol, exists := a.symbols[node.Value]; exists {
			return symbol.Type
		}
		return "unknown"
	default:
		return "unknown"
	}
}

func (a *Analyzer) areTypesCompatible(leftType, rightType, operator string) bool {
	// Operadores de comparación
	if operator == "==" || operator == "!=" || operator == "<" || 
	   operator == ">" || operator == "<=" || operator == ">=" {
		return leftType == rightType
	}

	// Operadores aritméticos
	if operator == "+" || operator == "-" || operator == "*" || operator == "/" {
		return (leftType == "int" || leftType == "float") && 
			   (rightType == "int" || rightType == "float")
	}

	return true
}

func (a *Analyzer) getNodeValue(node *SyntaxNode) string {
	if node == nil {
		return ""
	}
	return node.Value
}

func (a *Analyzer) getSymbolTable() []Symbol {
	var symbols []Symbol
	for _, symbol := range a.symbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// Funciones HTTP
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func analyzeCodeHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	analyzer := NewAnalyzer(request.Code)
	
	// Análisis léxico
	analyzer.lexicalAnalysis()
	
	// Análisis sintáctico
	syntaxTree := analyzer.syntaxAnalysis()
	
	// Análisis semántico
	analyzer.semanticAnalysis(syntaxTree)

	result := AnalysisResult{
		Tokens:         analyzer.tokens,
		SyntaxTree:     syntaxTree,
		SemanticErrors: analyzer.errors,
		SymbolTable:    analyzer.getSymbolTable(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/analyze", analyzeCodeHandler)
	
	fmt.Println("Servidor iniciado en el puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}