import React, { useState } from "react";
import "./App.css";

function App() {
  const [code, setCode] = useState(`package main

import "fmt"

func main() {
  query := "SELECT * FROM users WHERE age > 18"
  fmt.Println(query)
}`);

  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleAnalyze = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const res = await fetch("http://localhost:8080/analyze", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }),
      });

      if (!res.ok) {
        throw new Error(`Error en la petición: ${res.statusText}`);
      }

      const data = await res.json();
      setResult(data);
    } catch (error) {
      console.error("Error al analizar:", error);
      setError("Error al analizar el código");
    }

    setLoading(false);
  };

  return (
    <div className="container">
      <h1>Analizador SQL - Tokens</h1>
      <textarea
        value={code}
        onChange={(e) => setCode(e.target.value)}
        rows={10}
        style={{ width: "100%", fontFamily: "monospace" }}
      ></textarea>

      <button onClick={handleAnalyze} disabled={loading}>
        {loading ? "Analizando..." : "Analizar"}
      </button>

      {error && <p style={{ color: "red" }}>{error}</p>}

      {result && (
        <>
          <h2>Tokens encontrados</h2>

          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={{ border: "1px solid black", padding: "8px" }}>Categoría</th>
                <th style={{ border: "1px solid black", padding: "8px" }}>Tokens</th>
                <th style={{ border: "1px solid black", padding: "8px" }}>Total</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>Palabras Reservadas</td>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  {(result.reservedWords && result.reservedWords.length > 0)
                    ? result.reservedWords.join(", ")
                    : "-"}
                </td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.reservedWords ? result.reservedWords.length : 0}
                </td>
              </tr>

              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>Operadores</td>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  {(result.operators && result.operators.length > 0)
                    ? result.operators.join(", ")
                    : "-"}
                </td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.operators ? result.operators.length : 0}
                </td>
              </tr>

              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>Números</td>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  {(result.numbers && result.numbers.length > 0)
                    ? result.numbers.join(", ")
                    : "-"}
                </td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.numbers ? result.numbers.length : 0}
                </td>
              </tr>

              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>Símbolos</td>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  {(result.symbols && result.symbols.length > 0)
                    ? result.symbols.join(", ")
                    : "-"}
                </td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.symbols ? result.symbols.length : 0}
                </td>
              </tr>

              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>Cadenas</td>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  {(result.strings && result.strings.length > 0)
                    ? result.strings.map(s => `"${s}"`).join(", ")
                    : "-"}
                </td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.strings ? result.strings.length : 0}
                </td>
              </tr>

              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>Comentarios</td>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  {(result.comments && result.comments.length > 0)
                    ? result.comments.join(", ")
                    : "-"}
                </td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.comments ? result.comments.length : 0}
                </td>
              </tr>

              <tr>
                <td style={{ border: "1px solid black", padding: "8px" }}>
                  <strong>Identificadores (aprox.)</strong>
                </td>
                <td style={{ border: "1px solid black", padding: "8px" }}>-</td>
                <td style={{ border: "1px solid black", padding: "8px", textAlign: "center" }}>
                  {result.totals && result.totals.identifiers ? result.totals.identifiers : 0}
                </td>
              </tr>
            </tbody>
          </table>
        </>
      )}
    </div>
  );
}

export default App;
