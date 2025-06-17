package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"strconv"
	"unicode"
)

type AnalysisRequest struct {
	Code string `json:"code"`
}

type ErrorDetail struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Type    string `json:"type"` // "syntax", "semantic", "lexical"
}

type AnalysisResult struct {
	ReservedWords  []string          `json:"reservedWords"`
	Operators      []string          `json:"operators"`
	Numbers        []string          `json:"numbers"`
	Symbols        []string          `json:"symbols"`
	Strings        []string          `json:"strings"`
	Comments       []string          `json:"comments"`
	Totals         map[string]int    `json:"totals"`
	Suggestions    []string          `json:"suggestions"`
	SyntaxErrors   []ErrorDetail     `json:"syntaxErrors"`
	SemanticErrors []ErrorDetail     `json:"semanticErrors"`
	LexicalErrors  []ErrorDetail     `json:"lexicalErrors"`
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)

	port := ":8080"
	println("Servidor iniciado en http://localhost" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		println("Error al iniciar el servidor:", err.Error())
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req AnalysisRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	code := req.Code
	result := analyzeCode(code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func analyzeCode(code string) AnalysisResult {
	lines := strings.Split(code, "\n")

	result := AnalysisResult{
		ReservedWords:  []string{},
		Operators:      []string{},
		Numbers:        []string{},
		Symbols:        []string{},
		Strings:        []string{},
		Comments:       []string{},
		Totals:         map[string]int{},
		Suggestions:    []string{},
		SyntaxErrors:   []ErrorDetail{},
		SemanticErrors: []ErrorDetail{},
		LexicalErrors:  []ErrorDetail{},
	}

	// Listas de tokens válidos
	reservedWordsList := []string{"function", "return", "if", "else", "for", "while", "echo", "class", "public", "private", "protected", "static", "var", "const", "let", "int", "string", "bool", "float", "double", "void", "true", "false", "null", "undefined", "new", "this", "extends", "implements", "try", "catch", "finally", "throw", "switch", "case", "default", "break", "continue", "do", "typeof", "instanceof"}
	operatorsList := []string{"+", "-", "*", "/", "%", "=", "==", "===", "!=", "!==", "<", ">", "<=", ">=", "&&", "||", "!", "++", "--", "+=", "-=", "*=", "/=", "%=", "?", ":", "&", "|", "^", "~", "<<", ">>", ">>>"}

	// Expresiones regulares mejoradas
	wordRegex := regexp.MustCompile(`[a-zA-Z_$][a-zA-Z0-9_$]*`)
	numberRegex := regexp.MustCompile(`\b\d+(\.\d+)?([eE][+-]?\d+)?\b`)
	stringRegex := regexp.MustCompile(`"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|` + "`" + `(?:[^` + "`" + `\\]|\\.)*` + "`")
	commentRegex := regexp.MustCompile(`//.*|/\*[\s\S]*?\*/`)

	// Sets para evitar duplicados
	reservedSet := map[string]bool{}
	operatorSet := map[string]bool{}
	numberSet := map[string]bool{}
	symbolSet := map[string]bool{}
	stringSet := map[string]bool{}
	commentSet := map[string]bool{}

	// Contadores para balanceo de símbolos
	bracketStack := []rune{}

	// Errores comunes de escritura
	commonMistakes := map[string]string{
		"fuction":    "function",
		"funtion":    "function",
		"funciton":   "function",
		"fucntion":   "function",
		"retrn":      "return",
		"retrun":     "return",
		"retur":      "return",
		"pubic":      "public",
		"publc":      "public",
		"pritvate":   "private",
		"privte":     "private",
		"proteted":   "protected",
		"protcted":   "protected",
		"satic":      "static",
		"statc":      "static",
		"clas":       "class",
		"clss":       "class",
		"consle":     "console",
		"conole":     "console",
		"consol":     "console",
		"consoe":     "console",
		"documnet":   "document",
		"documen":    "document",
		"docuemnt":   "document",
		"lenght":     "length",
		"lenth":      "length",
		"undefineed": "undefined",
		"undefind":   "undefined",
		"udefined":   "undefined",
		"tru":        "true",
		"flase":      "false",
		"fals":       "false",
		"nul":        "null",
		"iff":        "if",
		"esle":       "else",
		"els":        "else",
		"whle":       "while",
		"wile":       "while",
		"fo":         "for",
		"forr":       "for",
		"swich":      "switch",
		"swicth":     "switch",
		"cas":        "case",
		"cse":        "case",
		"defalt":     "default",
		"deafult":    "default",
		"brek":       "break",
		"breka":      "break",
		"continu":    "continue",
		"contineu":   "continue",
		"tyr":        "try",
		"ctach":      "catch",
		"catc":       "catch",
		"finaly":     "finally",
		"finall":     "finally",
		"thow":       "throw",
		"throW":      "throw",
		"ne":         "new",
		"nw":         "new",
		"ths":        "this",
		"tihs":       "this",
		"var":        "var",
		"vr":         "var",
		"le":         "let",
		"lte":        "let",
		"cons":       "const",
		"cnst":       "const",
		"conts":      "const",
	}

	for lineNum, line := range lines {
		lineIndex := lineNum + 1
		
		// Verificar strings sin cerrar
		checkUnclosedStrings(line, lineIndex, &result)
		
		// Verificar comentarios sin cerrar
		checkUnclosedComments(line, lineIndex, &result)
		
		// Verificar balanceo de paréntesis/llaves/corchetes
		checkBracketBalance(line, lineIndex, &bracketStack, &result)
		
		// Verificar punto y coma faltante
		checkMissingSemicolon(line, lineIndex, &result)
		
		// Verificar operadores mal formados
		checkMalformedOperators(line, lineIndex, &result)
		
		// Verificar números mal formados
		checkMalformedNumbers(line, lineIndex, &result)

		// Extraer comentarios
		comments := commentRegex.FindAllString(line, -1)
		for _, c := range comments {
			if !commentSet[c] {
				commentSet[c] = true
				result.Comments = append(result.Comments, c)
			}
		}

		// Extraer strings
		stringsFound := stringRegex.FindAllString(line, -1)
		for _, s := range stringsFound {
			if !stringSet[s] {
				stringSet[s] = true
				result.Strings = append(result.Strings, s)
			}
		}

		// Extraer palabras
		words := wordRegex.FindAllString(line, -1)
		for _, w := range words {
			lw := strings.ToLower(w)
			
			// Verificar palabras reservadas
			if contains(reservedWordsList, lw) && !reservedSet[lw] {
				reservedSet[lw] = true
				result.ReservedWords = append(result.ReservedWords, lw)
			}
			
			// Verificar errores comunes de escritura
			if correction, exists := commonMistakes[lw]; exists {
				col := strings.Index(line, w) + 1
				result.SemanticErrors = append(result.SemanticErrors, ErrorDetail{
					Line:    lineIndex,
					Column:  col,
					Message: "Se esperaba '" + correction + "' pero se encontró '" + w + "'",
					Type:    "semantic",
				})
				result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineIndex)+": Corrige '"+w+"' por '"+correction+"'")
			}
		}

		// Extraer operadores
		for _, op := range operatorsList {
			if strings.Contains(line, op) && !operatorSet[op] {
				operatorSet[op] = true
				result.Operators = append(result.Operators, op)
			}
		}

		// Extraer números
		numbers := numberRegex.FindAllString(line, -1)
		for _, n := range numbers {
			if !numberSet[n] {
				numberSet[n] = true
				result.Numbers = append(result.Numbers, n)
			}
		}

		// Extraer símbolos
		symbols := []string{"(", ")", "{", "}", "[", "]", ";", ",", ".", ":", "?"}
		for _, sym := range symbols {
			if strings.Contains(line, sym) && !symbolSet[sym] {
				symbolSet[sym] = true
				result.Symbols = append(result.Symbols, sym)
			}
		}
	}

	// Verificar balance final de brackets
	checkFinalBracketBalance(&bracketStack, &result)

	// Calcular totales
	result.Totals["reservedWords"] = len(result.ReservedWords)
	result.Totals["operators"] = len(result.Operators)
	result.Totals["numbers"] = len(result.Numbers)
	result.Totals["symbols"] = len(result.Symbols)
	result.Totals["strings"] = len(result.Strings)
	result.Totals["comments"] = len(result.Comments)
	result.Totals["syntaxErrors"] = len(result.SyntaxErrors)
	result.Totals["semanticErrors"] = len(result.SemanticErrors)
	result.Totals["lexicalErrors"] = len(result.LexicalErrors)

	return result
}

func checkUnclosedStrings(line string, lineNum int, result *AnalysisResult) {
	inString := false
	stringChar := byte(0)
	escaped := false
	
	for _, char := range []byte(line) {
		if escaped {
			escaped = false
			continue
		}
		
		if char == '\\' {
			escaped = true
			continue
		}
		
		if (char == '"' || char == '\'' || char == '`') && !inString {
			inString = true
			stringChar = char
		} else if char == stringChar && inString {
			inString = false
			stringChar = 0
		}
	}
	
	if inString {
		result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
			Line:    lineNum,
			Column:  len(line),
			Message: "String sin cerrar (falta comilla " + string(stringChar) + ")",
			Type:    "syntax",
		})
		result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Cierra la cadena con comilla "+string(stringChar))
	}
}

func checkUnclosedComments(line string, lineNum int, result *AnalysisResult) {
	if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
		result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
			Line:    lineNum,
			Column:  strings.Index(line, "/*") + 1,
			Message: "Comentario de bloque sin cerrar (falta */)",
			Type:    "syntax",
		})
		result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Cierra el comentario con */")
	}
}

func checkBracketBalance(line string, lineNum int, bracketStack *[]rune, result *AnalysisResult) {
	bracketMap := map[rune]rune{')': '(', '}': '{', ']': '['}
	openBrackets := map[rune]bool{'(': true, '{': true, '[': true}
	
	for i, char := range line {
		if openBrackets[char] {
			*bracketStack = append(*bracketStack, char)
		} else if expectedOpen, isClosing := bracketMap[char]; isClosing {
			if len(*bracketStack) == 0 {
				result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
					Line:    lineNum,
					Column:  i + 1,
					Message: "Símbolo de cierre '" + string(char) + "' sin apertura correspondiente",
					Type:    "syntax",
				})
				result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Elimina el símbolo '"+string(char)+"' o agrega '"+string(expectedOpen)+"' antes")
			} else {
				lastOpen := (*bracketStack)[len(*bracketStack)-1]
				if lastOpen != expectedOpen {
					result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
						Line:    lineNum,
						Column:  i + 1,
						Message: "Símbolo de cierre '" + string(char) + "' no coincide con apertura '" + string(lastOpen) + "'",
						Type:    "syntax",
					})
					result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Cambia '"+string(char)+"' por '"+string(bracketMap[expectedOpen])+"' o corrige la apertura")
				} else {
					*bracketStack = (*bracketStack)[:len(*bracketStack)-1]
				}
			}
		}
	}
}

func checkFinalBracketBalance(bracketStack *[]rune, result *AnalysisResult) {
	for _, openBracket := range *bracketStack {
		var closeBracket rune
		switch openBracket {
		case '(':
			closeBracket = ')'
		case '{':
			closeBracket = '}'
		case '[':
			closeBracket = ']'
		}
		
		result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
			Line:    -1,
			Column:  -1,
			Message: "Símbolo de apertura '" + string(openBracket) + "' sin cierre correspondiente '" + string(closeBracket) + "'",
			Type:    "syntax",
		})
		result.Suggestions = append(result.Suggestions, "Agrega el símbolo de cierre '"+string(closeBracket)+"' al final del código")
	}
}

func checkMissingSemicolon(line string, lineNum int, result *AnalysisResult) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return
	}
	
	needsSemicolon := []string{"return", "break", "continue", "var", "let", "const"}
	endsWithKeyword := false
	
	for _, keyword := range needsSemicolon {
		if strings.Contains(trimmed, keyword) && !strings.HasSuffix(trimmed, ";") && !strings.HasSuffix(trimmed, "{") && !strings.HasSuffix(trimmed, "}") {
			endsWithKeyword = true
			break
		}
	}
	
	if endsWithKeyword {
		result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
			Line:    lineNum,
			Column:  len(line),
			Message: "Falta punto y coma (;) al final de la línea",
			Type:    "syntax",
		})
		result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Agrega punto y coma (;) al final")
	}
}

func checkMalformedOperators(line string, lineNum int, result *AnalysisResult) {
	// Buscar operadores mal formados como = en lugar de == o ===
	for i := 0; i < len(line)-1; i++ {
		// Detectar posible = en lugar de ==
		if line[i] == '=' && i > 0 && i+1 < len(line) {
			// Si hay una letra antes y después del =, podría ser asignación válida
			if unicode.IsLetter(rune(line[i-1])) && unicode.IsSpace(rune(line[i+1])) {
				continue // Probablemente asignación válida
			}
			
			// Si hay espacio antes del = y un valor después, podría ser comparación mal formada
			if i > 0 && unicode.IsSpace(rune(line[i-1])) && i+1 < len(line) && line[i+1] != '=' {
				// Buscar si hay contexto de comparación (if, while, etc.)
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, "if") || strings.Contains(trimmed, "while") || strings.Contains(trimmed, "for") {
					result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
						Line:    lineNum,
						Column:  i + 1,
						Message: "Posible operador de comparación mal formado (¿quisiste decir '==' o '==='?)",
						Type:    "syntax",
					})
					result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Verifica si necesitas '==' para comparación en lugar de '='")
				}
			}
		}
		
		// Detectar && mal escrito como &
		if line[i] == '&' && i+1 < len(line) && line[i+1] != '&' {
			result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
				Line:    lineNum,
				Column:  i + 1,
				Message: "Posible operador lógico mal formado (¿quisiste decir '&&'?)",
				Type:    "syntax",
			})
			result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Usa '&&' para operador lógico AND")
		}
		
		// Detectar || mal escrito como |
		if line[i] == '|' && i+1 < len(line) && line[i+1] != '|' {
			result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
				Line:    lineNum,
				Column:  i + 1,
				Message: "Posible operador lógico mal formado (¿quisiste decir '||'?)",
				Type:    "syntax",
			})
			result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Usa '||' para operador lógico OR")
		}
	}
}

func checkMalformedNumbers(line string, lineNum int, result *AnalysisResult) {
	// Buscar números mal formados como 123.45.67 o 123..45
	numberRegex := regexp.MustCompile(`\d+\.+\d+\.+\d*|\d+\.\.+\d*`)
	matches := numberRegex.FindAllString(line, -1)
	
	for _, match := range matches {
		col := strings.Index(line, match) + 1
		result.LexicalErrors = append(result.LexicalErrors, ErrorDetail{
			Line:    lineNum,
			Column:  col,
			Message: "Número mal formado: '" + match + "'",
			Type:    "lexical",
		})
		result.Suggestions = append(result.Suggestions, "Línea "+strconv.Itoa(lineNum)+": Corrige el formato del número '"+match+"'")
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}