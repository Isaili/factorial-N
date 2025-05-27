package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var sqliteReservedWords = map[string]bool{
	"ABORT": true, "ACTION": true, "ADD": true, "AFTER": true,
	"ALL": true, "ALTER": true, "ANALYZE": true, "AND": true,
	"AS": true, "ASC": true, "ATTACH": true, "AUTOINCREMENT": true,
	"BEFORE": true, "BEGIN": true, "BETWEEN": true, "BY": true,
	"CASCADE": true, "CASE": true, "CAST": true, "CHECK": true,
	"COLLATE": true, "COLUMN": true, "COMMIT": true, "CONFLICT": true,
	"CONSTRAINT": true, "CREATE": true, "CROSS": true, "CURRENT_DATE": true,
	"CURRENT_TIME": true, "CURRENT_TIMESTAMP": true, "DATABASE": true, "DEFAULT": true,
	"DEFERRABLE": true, "DEFERRED": true, "DELETE": true, "DESC": true,
	"DETACH": true, "DISTINCT": true, "DROP": true, "EACH": true,
	"ELSE": true, "END": true, "ESCAPE": true, "EXCEPT": true,
	"EXCLUSIVE": true, "EXISTS": true, "EXPLAIN": true, "FAIL": true,
	"FOR": true, "FOREIGN": true, "FROM": true, "FULL": true,
	"GLOB": true, "GROUP": true, "HAVING": true, "IF": true,
	"IGNORE": true, "IMMEDIATE": true, "IN": true, "INDEX": true,
	"INDEXED": true, "INITIALLY": true, "INNER": true, "INSERT": true,
	"INSTEAD": true, "INTERSECT": true, "INTO": true, "IS": true,
	"ISNULL": true, "JOIN": true, "KEY": true, "LEFT": true,
	"LIKE": true, "LIMIT": true, "MATCH": true, "NATURAL": true,
	"NO": true, "NOT": true, "NOTNULL": true, "NULL": true,
	"OF": true, "OFFSET": true, "ON": true, "OR": true,
	"ORDER": true, "OUTER": true, "PLAN": true, "PRAGMA": true,
	"PRIMARY": true, "QUERY": true, "RAISE": true, "RECURSIVE": true,
	"REFERENCES": true, "REGEXP": true, "REINDEX": true, "RELEASE": true,
	"RENAME": true, "REPLACE": true, "RESTRICT": true, "RIGHT": true,
	"ROLLBACK": true, "ROW": true, "SAVEPOINT": true, "SELECT": true,
	"SET": true, "TABLE": true, "TEMP": true, "TEMPORARY": true,
	"THEN": true, "TO": true, "TRANSACTION": true, "TRIGGER": true,
	"UNION": true, "UNIQUE": true, "UPDATE": true, "USING": true,
	"VACUUM": true, "VALUES": true, "VIEW": true, "VIRTUAL": true,
	"WHEN": true, "WHERE": true, "WITH": true, "WITHOUT": true,
}


type Request struct {
	Code string `json:"code"`
}

type Response struct {
	ReservedWords []string          `json:"reservedWords"`
	Operators     []string          `json:"operators"`
	Numbers       []string          `json:"numbers"`
	Symbols       []string          `json:"symbols"`
	Strings       []string          `json:"strings"`
	Comments      []string          `json:"comments"`
	Totals        map[string]int    `json:"totals"`
}


var (
	reStringDouble = regexp.MustCompile(`"([^"]*)"`)
	reStringSingle = regexp.MustCompile(`'([^']*)'`)
	reComment      = regexp.MustCompile(`//.*`)
	reNumber       = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	reOperator     = regexp.MustCompile(`<>|<=|>=|=|\+|-|\*|/`)
	reSymbol       = regexp.MustCompile(`[=+\-*/;,()]`)
	reWord         = regexp.MustCompile(`\b\w+\b`)
)

func analyzeCode(code string) Response {
	foundReserved := map[string]bool{}
	foundOperators := map[string]bool{}
	foundNumbers := map[string]bool{}
	foundSymbols := map[string]bool{}
	foundStrings := map[string]bool{}
	foundComments := map[string]bool{}


	comments := reComment.FindAllString(code, -1)
	for _, c := range comments {
		foundComments[c] = true
	}
	codeWithoutComments := reComment.ReplaceAllString(code, " ")

	
	stringsDouble := reStringDouble.FindAllStringSubmatch(codeWithoutComments, -1)
	for _, m := range stringsDouble {
		if len(m) > 1 {
			foundStrings[m[1]] = true
		}
	}
	codeWithoutStrings := reStringDouble.ReplaceAllString(codeWithoutComments, " ")

	
	stringsSingle := reStringSingle.FindAllStringSubmatch(codeWithoutStrings, -1)
	for _, m := range stringsSingle {
		if len(m) > 1 {
			foundStrings[m[1]] = true
		}
	}
	codeClean := reStringSingle.ReplaceAllString(codeWithoutStrings, " ")

	
	operators := reOperator.FindAllString(codeClean, -1)
	for _, op := range operators {
		foundOperators[op] = true
	}

	
	symbols := reSymbol.FindAllString(codeClean, -1)
	for _, sym := range symbols {
		foundSymbols[sym] = true
	}

	
	numbers := reNumber.FindAllString(codeClean, -1)
	for _, num := range numbers {
		foundNumbers[num] = true
	}

	
	words := reWord.FindAllString(codeClean, -1)
	for _, w := range words {
		upper := strings.ToUpper(w)
		if sqliteReservedWords[upper] {
			foundReserved[upper] = true
		}
	}

	for str := range foundStrings {
		words := reWord.FindAllString(str, -1)
		for _, w := range words {
			upper := strings.ToUpper(w)
			if sqliteReservedWords[upper] {
				foundReserved[upper] = true
			}
		}
	}

	reservedSlice := mapKeysToSlice(foundReserved)
	operatorSlice := mapKeysToSlice(foundOperators)
	numberSlice := mapKeysToSlice(foundNumbers)
	symbolSlice := mapKeysToSlice(foundSymbols)
	stringSlice := mapKeysToSlice(foundStrings)
	commentSlice := mapKeysToSlice(foundComments)


	totals := map[string]int{
		"identifiers": len(words) - len(foundReserved), // aproximación
		"operators":   len(operatorSlice),
		"numbers":     len(numberSlice),
		"symbols":     len(symbolSlice),
		"comments":    len(commentSlice),
	}

	return Response{
		ReservedWords: reservedSlice,
		Operators:     operatorSlice,
		Numbers:       numberSlice,
		Symbols:       symbolSlice,
		Strings:       stringSlice,
		Comments:      commentSlice,
		Totals:        totals,
	}
}

func mapKeysToSlice(m map[string]bool) []string {
	s := []string{}
	for k := range m {
		s = append(s, k)
	}
	return s
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	// Headers CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error en JSON", http.StatusBadRequest)
		return
	}

	res := analyzeCode(req.Code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	fmt.Println("Servidor corriendo en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
