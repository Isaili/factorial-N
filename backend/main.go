package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"strconv"
)

type AnalysisRequest struct {
	Code string `json:"code"`
}

type ErrorDetail struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
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
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler) // <-- REGISTRA el handler

	port := ":8080"
	println("Servidor iniciado en http://localhost" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		println("Error al iniciar el servidor:", err.Error())
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // o especifica "http://localhost:3000" si usas React
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
	}

	reservedWordsList := []string{"function", "return", "if", "else", "for", "while", "echo", "class", "public", "private", "protected", "static", "var", "const"}
	operatorsList := []string{"+", "-", "*", "/", "=", "==", "===", "!=", "<", ">", "<=", ">="}

	wordRegex := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
	numberRegex := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	stringRegex := regexp.MustCompile(`"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'`)
	commentRegex := regexp.MustCompile(`//.*|/\*[\s\S]*?\*/`)

	reservedSet := map[string]bool{}
	operatorSet := map[string]bool{}
	numberSet := map[string]bool{}
	symbolSet := map[string]bool{}
	stringSet := map[string]bool{}
	commentSet := map[string]bool{}

	for i, line := range lines {
		lineNum := i + 1

		comments := commentRegex.FindAllString(line, -1)
		for _, c := range comments {
			if !commentSet[c] {
				commentSet[c] = true
				result.Comments = append(result.Comments, c)
			}
		}

		stringsFound := stringRegex.FindAllString(line, -1)
		for _, s := range stringsFound {
			if !stringSet[s] {
				stringSet[s] = true
				result.Strings = append(result.Strings, s)
			}
		}

		words := wordRegex.FindAllString(line, -1)
		for _, w := range words {
			lw := strings.ToLower(w)
			if contains(reservedWordsList, lw) && !reservedSet[lw] {
				reservedSet[lw] = true
				result.ReservedWords = append(result.ReservedWords, lw)
			}
		}

		if strings.Contains(line, "fuction") {
			result.SyntaxErrors = append(result.SyntaxErrors, ErrorDetail{
				Line:    lineNum,
				Message: "Se esperaba 'function' pero se encontró 'fuction'",
			})
			result.Suggestions = append(result.Suggestions, "Corrige 'fuction' por 'function' en línea "+itoa(lineNum))
		}

		for _, op := range operatorsList {
			if strings.Contains(line, op) && !operatorSet[op] {
				operatorSet[op] = true
				result.Operators = append(result.Operators, op)
			}
		}

		numbers := numberRegex.FindAllString(line, -1)
		for _, n := range numbers {
			if !numberSet[n] {
				numberSet[n] = true
				result.Numbers = append(result.Numbers, n)
			}
		}

		symbols := []string{"(", ")", "{", "}", ";", ","}
		for _, sym := range symbols {
			if strings.Contains(line, sym) && !symbolSet[sym] {
				symbolSet[sym] = true
				result.Symbols = append(result.Symbols, sym)
			}
		}
	}

	if strings.Contains(code, "retrn") {
		lineWithError := findLineWithSubstring(lines, "retrn")
		result.SemanticErrors = append(result.SemanticErrors, ErrorDetail{
			Line:    lineWithError,
			Message: "Se esperaba 'return' pero se encontró 'retrn'",
		})
		result.Suggestions = append(result.Suggestions, "Corrige 'retrn' por 'return' en línea "+itoa(lineWithError))
	}

	result.Totals["reservedWords"] = len(result.ReservedWords)
	result.Totals["operators"] = len(result.Operators)
	result.Totals["numbers"] = len(result.Numbers)
	result.Totals["symbols"] = len(result.Symbols)
	result.Totals["strings"] = len(result.Strings)
	result.Totals["comments"] = len(result.Comments)
	result.Totals["syntaxErrors"] = len(result.SyntaxErrors)
	result.Totals["semanticErrors"] = len(result.SemanticErrors)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func findLineWithSubstring(lines []string, substr string) int {
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i + 1
		}
	}
	return 0
}
