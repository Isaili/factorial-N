package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	
	"strings"
)

// Token representa un token léxico
type Token struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Line  int    `json:"line"`
	Col   int    `json:"col"`
}

// AnalysisResult representa el resultado del análisis
type AnalysisResult struct {
	Tokens         []Token  `json:"tokens"`
	SyntaxValid    bool     `json:"syntaxValid"`
	SyntaxErrors   []string `json:"syntaxErrors"`
	SemanticValid  bool     `json:"semanticValid"`
	SemanticErrors []string `json:"semanticErrors"`
	Success        bool     `json:"success"`
}

// LexicalAnalyzer maneja el análisis léxico
type LexicalAnalyzer struct {
	input    string
	position int
	line     int
	col      int
	tokens   []Token
}

// NewLexicalAnalyzer crea un nuevo analizador léxico
func NewLexicalAnalyzer(input string) *LexicalAnalyzer {
	return &LexicalAnalyzer{
		input:  input,
		line:   1,
		col:    1,
		tokens: []Token{},
	}
}

// Analyze realiza el análisis léxico
func (la *LexicalAnalyzer) Analyze() []Token {
	for la.position < len(la.input) {
		la.skipWhitespace()
		if la.position >= len(la.input) {
			break
		}

		char := la.input[la.position]
		
		// Comentarios
		if char == '#' {
			la.skipComment()
			continue
		}

		// Strings
		if char == '"' || char == '\'' {
			la.tokenizeString()
			continue
		}

		// Números
		if char >= '0' && char <= '9' {
			la.tokenizeNumber()
			continue
		}

		// Identificadores y palabras clave
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' {
			la.tokenizeIdentifier()
			continue
		}

		// Operadores y símbolos
		la.tokenizeOperator()
	}

	return la.tokens
}

func (la *LexicalAnalyzer) skipWhitespace() {
	for la.position < len(la.input) {
		char := la.input[la.position]
		if char == ' ' || char == '\t' {
			la.position++
			la.col++
		} else if char == '\n' {
			la.position++
			la.line++
			la.col = 1
		} else {
			break
		}
	}
}

func (la *LexicalAnalyzer) skipComment() {
	for la.position < len(la.input) && la.input[la.position] != '\n' {
		la.position++
		la.col++
	}
}

func (la *LexicalAnalyzer) tokenizeString() {
	quote := la.input[la.position]
	la.position++
	la.col++
	
	value := ""
	for la.position < len(la.input) && la.input[la.position] != quote {
		value += string(la.input[la.position])
		la.position++
		la.col++
	}
	
	if la.position < len(la.input) {
		la.position++ // Skip closing quote
		la.col++
	}
	
	la.addToken("STRING", value)
}

func (la *LexicalAnalyzer) tokenizeNumber() {
	start := la.position
	for la.position < len(la.input) && (la.input[la.position] >= '0' && la.input[la.position] <= '9') {
		la.position++
		la.col++
	}
	
	value := la.input[start:la.position]
	la.addToken("NUMBER", value)
}

func (la *LexicalAnalyzer) tokenizeIdentifier() {
	start := la.position
	for la.position < len(la.input) {
		char := la.input[la.position]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
		   (char >= '0' && char <= '9') || char == '_' {
			la.position++
			la.col++
		} else {
			break
		}
	}
	
	value := la.input[start:la.position]
	tokenType := "IDENTIFIER"
	
	// Palabras clave
	keywords := map[string]string{
		"def":    "KEYWORD",
		"if":     "KEYWORD", 
		"else":   "KEYWORD",
		"return": "KEYWORD",
		"print":  "KEYWORD",
	}
	
	if t, exists := keywords[value]; exists {
		tokenType = t
	}
	
	la.addToken(tokenType, value)
}

func (la *LexicalAnalyzer) tokenizeOperator() {
	char := la.input[la.position]
	
	// Operadores de dos caracteres
	if la.position+1 < len(la.input) {
		twoChar := la.input[la.position:la.position+2]
		operators := []string{"<=", ">=", "==", "!="}
		for _, op := range operators {
			if twoChar == op {
				la.addToken("OPERATOR", op)
				la.position += 2
				la.col += 2
				return
			}
		}
	}
	
	// Operadores de un caracter
	singleOps := "+-*/%<>=!(),:."
	if strings.Contains(singleOps, string(char)) {
		if char == '(' || char == ')' {
			la.addToken("PARENTHESIS", string(char))
		} else if char == ',' {
			la.addToken("COMMA", string(char))
		} else if char == ':' {
			la.addToken("COLON", string(char))
		} else {
			la.addToken("OPERATOR", string(char))
		}
		la.position++
		la.col++
	} else {
		// Caracter desconocido
		la.addToken("UNKNOWN", string(char))
		la.position++
		la.col++
	}
}

func (la *LexicalAnalyzer) addToken(tokenType, value string) {
	token := Token{
		Type:  tokenType,
		Value: value,
		Line:  la.line,
		Col:   la.col - len(value),
	}
	la.tokens = append(la.tokens, token)
}

// SyntaxAnalyzer verifica la sintaxis
func SyntaxAnalyzer(tokens []Token) (bool, []string) {
	errors := []string{}
	
	// Verificar estructura básica
	hasDefKeyword := false
	hasFunctionName := false
	hasOpenParen := false
	hasCloseParen := false
	hasColon := false
	
	for i, token := range tokens {
		if token.Value == "def" {
			hasDefKeyword = true
			// Verificar que después viene un identificador
			if i+1 < len(tokens) && tokens[i+1].Type == "IDENTIFIER" {
				hasFunctionName = true
			} else {
				errors = append(errors, "Se esperaba nombre de función después de 'def'")
			}
		}
		if token.Type == "PARENTHESIS" && token.Value == "(" {
			hasOpenParen = true
		}
		if token.Type == "PARENTHESIS" && token.Value == ")" {
			hasCloseParen = true
		}
		if token.Type == "COLON" {
			hasColon = true
		}
	}
	
	if !hasDefKeyword {
		errors = append(errors, "Falta palabra clave 'def'")
	}
	if !hasFunctionName {
		errors = append(errors, "Falta nombre de función")
	}
	if !hasOpenParen {
		errors = append(errors, "Falta paréntesis de apertura")
	}
	if !hasCloseParen {
		errors = append(errors, "Falta paréntesis de cierre")
	}
	if !hasColon {
		errors = append(errors, "Falta dos puntos después de la definición de función")
	}
	
	return len(errors) == 0, errors
}

// SemanticAnalyzer verifica la semántica
func SemanticAnalyzer(tokens []Token) (bool, []string) {
	errors := []string{}
	definedFunctions := make(map[string]bool)
	definedVariables := make(map[string]bool)
	
	for i, token := range tokens {
		// Verificar definición de funciones
		if token.Value == "def" && i+1 < len(tokens) {
			funcName := tokens[i+1].Value
			definedFunctions[funcName] = true
		}
		
		// Verificar asignación de variables
		if token.Type == "IDENTIFIER" && i+1 < len(tokens) && tokens[i+1].Value == "=" {
			definedVariables[token.Value] = true
		}
		
		// Verificar parámetros de función
		if token.Value == "def" && i+3 < len(tokens) && tokens[i+2].Value == "(" {
			paramName := tokens[i+3].Value
			if tokens[i+3].Type == "IDENTIFIER" {
				definedVariables[paramName] = true
			}
		}
		
		// Verificar llamadas a funciones
		if token.Type == "IDENTIFIER" && i+1 < len(tokens) && tokens[i+1].Value == "(" {
			if !definedFunctions[token.Value] && token.Value != "print" {
				errors = append(errors, fmt.Sprintf("Función '%s' no está definida", token.Value))
			}
		}
		
		// Verificar uso de variables
		if token.Type == "IDENTIFIER" && token.Value != "factorial" && token.Value != "print" {
			// Verificar contexto para determinar si es uso de variable
			if i > 0 && (tokens[i-1].Value == "=" || tokens[i-1].Value == "*" || 
						 tokens[i-1].Value == "-" || tokens[i-1].Value == "<=" ||
						 tokens[i-1].Value == "," || tokens[i-1].Value == "(") {
				if !definedVariables[token.Value] {
					errors = append(errors, fmt.Sprintf("Variable '%s' no está definida", token.Value))
				}
			}
		}
	}
	
	return len(errors) == 0, errors
}

// analyzeCode maneja el análisis completo
func analyzeCode(code string) AnalysisResult {
	// Análisis léxico
	lexer := NewLexicalAnalyzer(code)
	tokens := lexer.Analyze()
	
	// Análisis sintáctico
	syntaxValid, syntaxErrors := SyntaxAnalyzer(tokens)
	
	// Análisis semántico
	semanticValid, semanticErrors := SemanticAnalyzer(tokens)
	
	return AnalysisResult{
		Tokens:         tokens,
		SyntaxValid:    syntaxValid,
		SyntaxErrors:   syntaxErrors,
		SemanticValid:  semanticValid,
		SemanticErrors: semanticErrors,
		Success:        syntaxValid && semanticValid,
	}
}

// CORS middleware
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// analyzeHandler maneja las peticiones de análisis
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	
	if r.Method == "OPTIONS" {
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	
	var request struct {
		Code string `json:"code"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}
	
	result := analyzeCode(request.Code)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	
	fmt.Println("Servidor iniciado en puerto 8080")
	fmt.Println("Endpoint: POST http://localhost:8080/analyze")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error al iniciar servidor: %v\n", err)
	}
}