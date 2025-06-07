import { useState } from "react";
import "./App.css"; // si usas Tailwind, puedes ignorar esto

export default function App() {
  const [code, setCode] = useState("");
  const [result, setResult] = useState(null);

  const analyzeCode = async () => {
    const response = await fetch("http://localhost:8080/analyze", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
    });

    const data = await response.json();
    setResult(data);
  };

  return (
    <div style={styles.page}>
      <h1 style={styles.title}>Analizador Léxico, Sintáctico y Semántico</h1>

      <textarea
        placeholder="Escribe tu código PHP aquí..."
        value={code}
        onChange={(e) => setCode(e.target.value)}
        style={styles.textarea}
      />

      <button onClick={analyzeCode} style={styles.button}>
        Analizar
      </button>

      {result && (
        <div style={styles.results}>
          {renderTable("Palabras Reservadas", result.reservedWords)}
          {renderTable("Operadores", result.operators)}
          {renderTable("Números", result.numbers)}
          {renderTable("Símbolos", result.symbols)}
          {renderTable("Cadenas", result.strings)}
          {renderTable("Comentarios", result.comments)}
          {renderTable("Errores Sintácticos", result.syntaxErrors?.map(e => `Línea ${e.line}: ${e.message}`))}
          {renderTable("Errores Semánticos", result.semanticErrors?.map(e => `Línea ${e.line}: ${e.message}`))}
          {renderTable("Sugerencias", result.suggestions)}

          <div style={styles.tableWrapper}>
            <h3>Totales</h3>
            <table style={styles.table}>
              <tbody>
                {Object.entries(result.totals).map(([key, value]) => (
                  <tr key={key}>
                    <td style={styles.cell}>{key}</td>
                    <td style={styles.cell}>{value}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

const renderTable = (title, items) => {
  if (!items || items.length === 0) return null;
  return (
    <div style={styles.tableWrapper}>
      <h3>{title}</h3>
      <table style={styles.table}>
        <tbody>
          {items.map((item, i) => (
            <tr key={i}>
              <td style={styles.cell}>{item}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

const styles = {
  page: {
    fontFamily: "Arial, sans-serif",
    padding: "40px 20px",
    maxWidth: "900px",
    margin: "auto",
    textAlign: "center",
  },
  title: {
    fontSize: "2rem",
    marginBottom: "30px",
  },
  textarea: {
    width: "100%",
    height: "200px",
    padding: "15px",
    fontSize: "16px",
    borderRadius: "8px",
    border: "1px solid #ccc",
    marginBottom: "20px",
    fontFamily: "monospace",
  },
  button: {
    padding: "10px 20px",
    fontSize: "16px",
    borderRadius: "6px",
    border: "none",
    backgroundColor: "#007bff",
    color: "white",
    cursor: "pointer",
    marginBottom: "30px",
  },
  results: {
    marginTop: "40px",
    textAlign: "left",
  },
  tableWrapper: {
    marginBottom: "30px",
  },
  table: {
    width: "100%",
    borderCollapse: "collapse",
    marginTop: "10px",
  },
  cell: {
    border: "1px solid #ddd",
    padding: "8px",
  },
};
