import React, { useState } from 'react';
import { Play, Code, AlertTriangle, CheckCircle, Table, FileText } from 'lucide-react';

const CodeAnalyzer = () => {
  const [code, setCode] = useState(`def main():
    edad = 22
    escuela = "upchiapas"
    if edad > 18:
        print("Mayor de edad")
    if escuela.lower() == "upchiapas":
        print("Bienvenido a UPChiapas")

if __name__ == "__main__":
    main()`);
  
  const [analysisResult, setAnalysisResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('lexical');

  const analyzeCode = async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8080/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
      });
      if (!response.ok) throw new Error('Error en el análisis');
      const result = await response.json();
      setAnalysisResult(result);
    } catch (error) {
      alert('Error al analizar el código. Asegúrate de que el servidor esté ejecutándose.');
    }
    setLoading(false);
  };

  const getTokenTypeColor = (type) => {
    const colors = {
      'KEYWORD': 'blue',
      'IDENTIFIER': 'green',
      'NUMBER': 'purple',
      'STRING': 'orange',
      'OPERATOR': 'red',
      'DELIMITER': 'gray',
      'NEWLINE': 'indigo',
      'COMMENT': 'darkorange',
      'UNKNOWN': 'darkred'
    };
    return colors[type] || 'black';
  };

  const getTypeColor = (type) => {
    const colors = {
      'int': 'purple',
      'float': 'pink',
      'string': 'orange',
      'bool': 'green',
      'function': 'blue'
    };
    return colors[type] || 'gray';
  };

  const renderTokensTable = () => {
    if (!analysisResult?.tokens) return null;
    return (
      <table className="table">
        <thead>
          <tr>
            <th>Posición</th>
            <th>Tipo</th>
            <th>Valor</th>
            <th>Línea</th>
            <th>Columna</th>
          </tr>
        </thead>
        <tbody>
          {analysisResult.tokens.map((token, index) => (
            <tr key={index}>
              <td>{token.position}</td>
              <td style={{ color: getTokenTypeColor(token.type) }}>{token.type}</td>
              <td><code>{token.value === '\n' ? '\\n' : token.value}</code></td>
              <td>{token.line}</td>
              <td>{token.column}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  };

  const renderSyntaxTree = () => {
    if (!analysisResult?.syntaxTree) return null;

    const renderNode = (node, depth = 0) => (
      <div key={Math.random()} style={{ marginLeft: `${depth * 20}px` }} className="syntax-node">
        <span className="syntax-type">{node.type}</span>
        {node.value && <span className="syntax-value"> {node.value}</span>}
        <span className="syntax-line"> (Línea {node.line})</span>
        {node.children && node.children.map(child => renderNode(child, depth + 1))}
      </div>
    );

    return (
      <div>
        <h3><FileText size={20} /> Árbol Sintáctico</h3>
        <div>{analysisResult.syntaxTree.map((node) => renderNode(node))}</div>
      </div>
    );
  };

  const renderSemanticErrors = () => {
    if (!analysisResult?.semanticErrors) return null;
    return (
      <div className="semantic-errors">
        <h3><AlertTriangle size={20} /> Errores Semánticos</h3>
        {analysisResult.semanticErrors.length === 0 ? (
          <div className="no-errors">
            <CheckCircle size={20} /> No se encontraron errores
          </div>
        ) : (
          analysisResult.semanticErrors.map((error, index) => (
            <div key={index} className="semantic-error">
              <strong>{error.type}</strong> (Línea {error.line})
              <p>{error.message}</p>
              {error.variable && <p>Variable: {error.variable}</p>}
              {error.expectedType && error.actualType && (
                <p>Esperado: {error.expectedType}, Actual: {error.actualType}</p>
              )}
            </div>
          ))
        )}
      </div>
    );
  };

  const renderSymbolTable = () => {
    if (!analysisResult?.symbolTable) return null;
    return (
      <table className="table">
        <thead>
          <tr>
            <th>Nombre</th>
            <th>Tipo</th>
            <th>Valor</th>
            <th>Línea</th>
            <th>Ámbito</th>
            <th>Usado</th>
          </tr>
        </thead>
        <tbody>
          {analysisResult.symbolTable.map((symbol, index) => (
            <tr key={index}>
              <td>{symbol.name}</td>
              <td style={{ color: getTypeColor(symbol.type) }}>{symbol.type}</td>
              <td><code>{symbol.value}</code></td>
              <td>{symbol.line}</td>
              <td>{symbol.scope}</td>
              <td>{symbol.used ? 'Sí' : 'No'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  };

  const tabs = [
    { id: 'lexical', label: 'Análisis Léxico', icon: Code },
    { id: 'syntax', label: 'Análisis Sintáctico', icon: FileText },
    { id: 'semantic', label: 'Análisis Semántico', icon: AlertTriangle },
    { id: 'symbols', label: 'Tabla de Símbolos', icon: Table }
  ];

  return (
    <div className="container">
      <h1>Analizador Léxico, Sintáctico y Semántico</h1>

      <div className="grid">
        <div>
          <h2>Código Python</h2>
          <textarea value={code} onChange={(e) => setCode(e.target.value)} />
          <button onClick={analyzeCode} disabled={loading}>
            {loading ? 'Analizando...' : <><Play size={16} /> Analizar Código</>}
          </button>
        </div>

        <div>
          <h2>Resumen del Análisis</h2>
          {analysisResult ? (
            <ul>
              <li>Tokens: {analysisResult.tokens?.length || 0}</li>
              <li>Nodos: {analysisResult.syntaxTree?.length || 0}</li>
              <li>Símbolos: {analysisResult.symbolTable?.length || 0}</li>
              <li>Errores Semánticos: {analysisResult.semanticErrors?.length || 0}</li>
            </ul>
          ) : (
            <p>Ejecuta el análisis para ver resultados.</p>
          )}
        </div>
      </div>

      {analysisResult && (
        <div>
          <div className="tabs">
            {tabs.map(({ id, label, icon: Icon }) => (
              <button key={id} onClick={() => setActiveTab(id)} className={activeTab === id ? 'active' : ''}>
                <Icon size={16} /> {label}
              </button>
            ))}
          </div>

          <div className="tab-content">
            {activeTab === 'lexical' && renderTokensTable()}
            {activeTab === 'syntax' && renderSyntaxTree()}
            {activeTab === 'semantic' && renderSemanticErrors()}
            {activeTab === 'symbols' && renderSymbolTable()}
          </div>
        </div>
      )}

      {/* Estilos embebidos */}
      <style>{`
        .container {
          padding: 20px;
          font-family: Arial, sans-serif;
          background-color: #f0f2f5;
        }
        h1 {
          text-align: center;
          margin-bottom: 30px;
        }
        .grid {
          display: flex;
          gap: 40px;
          margin-bottom: 40px;
        }
        textarea {
          width: 100%;
          height: 250px;
          font-family: monospace;
          padding: 10px;
          font-size: 14px;
          resize: vertical;
          border: 1px solid #ccc;
          border-radius: 4px;
        }
        button {
          margin-top: 10px;
          padding: 8px 16px;
          font-weight: bold;
          background-color: #007bff;
          color: white;
          border: none;
          border-radius: 4px;
          cursor: pointer;
        }
        button:disabled {
          background-color: #ccc;
          cursor: not-allowed;
        }
        .table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 10px;
        }
        .table th, .table td {
          border: 1px solid #ddd;
          padding: 8px;
        }
        .table th {
          background-color: #f8f8f8;
          text-align: left;
        }
        .tabs {
          display: flex;
          gap: 10px;
          margin-bottom: 20px;
        }
        .tabs button {
          padding: 6px 12px;
          background: #eee;
          border: none;
          cursor: pointer;
        }
        .tabs .active {
          background: #007bff;
          color: white;
          font-weight: bold;
        }
        .syntax-node {
          padding: 4px;
          border-left: 3px solid #007bff;
          margin-bottom: 5px;
          background: #fff;
        }
        .syntax-type {
          font-weight: bold;
          font-size: 13px;
          color: #007bff;
        }
        .syntax-value {
          font-family: monospace;
          color: #333;
          margin-left: 5px;
        }
        .syntax-line {
          font-size: 11px;
          color: #888;
          margin-left: 10px;
        }
        .semantic-errors {
          padding: 10px;
        }
        .semantic-error {
          border-left: 4px solid red;
          background: #ffe5e5;
          padding: 10px;
          margin-bottom: 10px;
        }
        .no-errors {
          color: green;
          font-weight: bold;
        }
      `}</style>
    </div>
  );
};

export default CodeAnalyzer;
