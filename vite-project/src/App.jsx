import React, { useState } from 'react';
import { Play, Code, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';

const FactorialAnalyzer = () => {
  const [code, setCode] = useState(`def factorial(n):
    if n <= 1:
        return 1
    else:
        return n * factorial(n - 1)

x = 5
print("El factorial de", x, "es", factorial(x))`);
  
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const analyzeCode = async () => {
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch('http://localhost:8080/analyze', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code }),
      });
      
      if (!response.ok) {
        throw new Error('Error en la comunicación con el servidor');
      }
      
      const data = await response.json();
      setResult(data);
    } catch (err) {
      setError('Error: ' + err.message + '. Asegúrate de que el servidor Go esté ejecutándose en puerto 8080.');
    } finally {
      setLoading(false);
    }
  };

  const getTokenStyle = (tokenType) => {
    const styles = {
      'KEYWORD': {
        backgroundColor: '#f3e8ff',
        color: '#6b21a8',
        border: '1px solid #e9d5ff'
      },
      'IDENTIFIER': {
        backgroundColor: '#dbeafe',
        color: '#1e40af',
        border: '1px solid #bfdbfe'
      },
      'NUMBER': {
        backgroundColor: '#dcfce7',
        color: '#166534',
        border: '1px solid #bbf7d0'
      },
      'STRING': {
        backgroundColor: '#fef3c7',
        color: '#92400e',
        border: '1px solid #fde68a'
      },
      'OPERATOR': {
        backgroundColor: '#fee2e2',
        color: '#991b1b',
        border: '1px solid #fecaca'
      },
      'PARENTHESIS': {
        backgroundColor: '#f3f4f6',
        color: '#374151',
        border: '1px solid #d1d5db'
      },
      'COMMA': {
        backgroundColor: '#f3f4f6',
        color: '#374151',
        border: '1px solid #d1d5db'
      },
      'COLON': {
        backgroundColor: '#f3f4f6',
        color: '#374151',
        border: '1px solid #d1d5db'
      },
      'UNKNOWN': {
        backgroundColor: '#fecaca',
        color: '#7f1d1d',
        border: '1px solid #f87171'
      }
    };
    return styles[tokenType] || styles['UNKNOWN'];
  };

  const containerStyle = {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #1e293b 0%, #334155 100%)',
    padding: '24px',
    fontFamily: 'system-ui, -apple-system, sans-serif'
  };

  const cardStyle = {
    backgroundColor: 'white',
    borderRadius: '12px',
    boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
    overflow: 'hidden',
    marginBottom: '24px'
  };

  const headerStyle = {
    backgroundColor: '#475569',
    padding: '16px 24px',
    color: 'white',
    fontWeight: '600',
    display: 'flex',
    alignItems: 'center',
    gap: '8px'
  };

  const buttonStyle = {
    backgroundColor: '#2563eb',
    color: 'white',
    padding: '8px 24px',
    borderRadius: '8px',
    border: 'none',
    fontWeight: '500',
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    cursor: 'pointer',
    transition: 'background-color 0.2s'
  };

  const textareaStyle = {
    width: '100%',
    height: '300px',
    padding: '16px',
    fontFamily: 'monospace',
    fontSize: '14px',
    backgroundColor: '#f8fafc',
    border: '1px solid #e2e8f0',
    borderRadius: '8px',
    resize: 'none',
    outline: 'none'
  };

  const tokenStyle = {
    display: 'inline-block',
    padding: '4px 12px',
    margin: '2px',
    borderRadius: '16px',
    fontSize: '12px',
    fontWeight: '500',
    cursor: 'default'
  };

  const errorStyle = {
    backgroundColor: '#fef2f2',
    border: '1px solid #fecaca',
    borderRadius: '8px',
    padding: '16px',
    marginTop: '16px'
  };

  const successStyle = {
    backgroundColor: '#f0fdf4',
    border: '2px solid #bbf7d0',
    borderRadius: '8px',
    padding: '16px'
  };

  const failStyle = {
    backgroundColor: '#fef2f2',
    border: '2px solid #fecaca',
    borderRadius: '8px',
    padding: '16px'
  };

  return (
    <div style={containerStyle}>
      <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
        {/* Título */}
        <div style={{ textAlign: 'center', marginBottom: '32px' }}>
          <h1 style={{ 
            fontSize: '2.5rem', 
            fontWeight: 'bold', 
            color: 'white', 
            marginBottom: '8px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '12px'
          }}>
            <Code size={40} style={{ color: '#60a5fa' }} />
            Analizador de Factorial
          </h1>
          <p style={{ color: '#cbd5e1' }}>Análisis Léxico, Sintáctico y Semántico</p>
        </div>

        {/* Panel de Código */}
        <div style={cardStyle}>
          <div style={headerStyle}>
            <Code size={20} />
            Editor de Código
          </div>
          <div style={{ padding: '24px' }}>
            <textarea
              value={code}
              onChange={(e) => setCode(e.target.value)}
              style={textareaStyle}
              placeholder="Ingresa tu código aquí..."
            />
            <div style={{ marginTop: '16px' }}>
              <button
                onClick={analyzeCode}
                disabled={loading}
                style={{
                  ...buttonStyle,
                  backgroundColor: loading ? '#94a3b8' : '#2563eb'
                }}
                onMouseOver={(e) => {
                  if (!loading) e.target.style.backgroundColor = '#1d4ed8';
                }}
                onMouseOut={(e) => {
                  if (!loading) e.target.style.backgroundColor = '#2563eb';
                }}
              >
                <Play size={16} />
                {loading ? 'Analizando...' : 'Analizar Código'}
              </button>
            </div>
            
            {error && (
              <div style={errorStyle}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#b91c1c' }}>
                  <XCircle size={16} />
                  <span style={{ fontWeight: '500' }}>Error</span>
                </div>
                <p style={{ color: '#dc2626', marginTop: '4px', margin: 0 }}>{error}</p>
              </div>
            )}
          </div>
        </div>

        {result && (
          <>
            {/* Resumen del Análisis */}
            <div style={cardStyle}>
              <div style={headerStyle}>
                Resumen del Análisis
              </div>
              <div style={{ padding: '24px' }}>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                  <div style={result.syntaxValid ? successStyle : failStyle}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                      {result.syntaxValid ? 
                        <CheckCircle style={{ color: '#16a34a' }} size={20} /> : 
                        <XCircle style={{ color: '#dc2626' }} size={20} />
                      }
                      <span style={{ 
                        fontWeight: '500', 
                        color: result.syntaxValid ? '#15803d' : '#b91c1c' 
                      }}>
                        Análisis Sintáctico
                      </span>
                    </div>
                    <p style={{ 
                      fontSize: '14px', 
                      color: result.syntaxValid ? '#16a34a' : '#dc2626',
                      margin: 0
                    }}>
                      {result.syntaxValid ? 'Sintaxis correcta' : `${result.syntaxErrors.length} errores`}
                    </p>
                  </div>
                  
                  <div style={result.semanticValid ? successStyle : failStyle}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                      {result.semanticValid ? 
                        <CheckCircle style={{ color: '#16a34a' }} size={20} /> : 
                        <XCircle style={{ color: '#dc2626' }} size={20} />
                      }
                      <span style={{ 
                        fontWeight: '500', 
                        color: result.semanticValid ? '#15803d' : '#b91c1c' 
                      }}>
                        Análisis Semántico
                      </span>
                    </div>
                    <p style={{ 
                      fontSize: '14px', 
                      color: result.semanticValid ? '#16a34a' : '#dc2626',
                      margin: 0
                    }}>
                      {result.semanticValid ? 'Semántica correcta' : `${result.semanticErrors.length} errores`}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Tokens */}
            <div style={cardStyle}>
              <div style={headerStyle}>
                Tokens Identificados ({result.tokens.length})
              </div>
              <div style={{ padding: '24px' }}>
                <div style={{ 
                  maxHeight: '300px', 
                  overflowY: 'auto',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                  padding: '16px'
                }}>
                  {result.tokens.map((token, index) => (
                    <span
                      key={index}
                      style={{
                        ...tokenStyle,
                        ...getTokenStyle(token.type)
                      }}
                      title={`Tipo: ${token.type} | Línea: ${token.line} | Columna: ${token.col}`}
                    >
                      {token.value}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            {/* Errores */}
            {(result.syntaxErrors.length > 0 || result.semanticErrors.length > 0) && (
              <div style={cardStyle}>
                <div style={{
                  ...headerStyle,
                  backgroundColor: '#dc2626'
                }}>
                  <AlertTriangle size={20} />
                  Errores Encontrados
                </div>
                <div style={{ padding: '24px' }}>
                  {result.syntaxErrors.length > 0 && (
                    <div style={{ marginBottom: '16px' }}>
                      <h3 style={{ fontWeight: '500', color: '#b91c1c', marginBottom: '8px' }}>
                        Errores Sintácticos:
                      </h3>
                      <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        {result.syntaxErrors.map((error, index) => (
                          <li key={index} style={{ 
                            color: '#dc2626', 
                            fontSize: '14px',
                            marginBottom: '4px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px'
                          }}>
                            <XCircle size={14} />
                            {error}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                  
                  {result.semanticErrors.length > 0 && (
                    <div>
                      <h3 style={{ fontWeight: '500', color: '#b91c1c', marginBottom: '8px' }}>
                        Errores Semánticos:
                      </h3>
                      <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        {result.semanticErrors.map((error, index) => (
                          <li key={index} style={{ 
                            color: '#dc2626', 
                            fontSize: '14px',
                            marginBottom: '4px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px'
                          }}>
                            <XCircle size={14} />
                            {error}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              </div>
            )}
          </>
        )}

        {/* Leyenda de Tokens */}
        <div style={cardStyle}>
          <div style={headerStyle}>
            Leyenda de Tokens
          </div>
          <div style={{ padding: '24px' }}>
            <div style={{ 
              display: 'grid', 
              gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', 
              gap: '16px' 
            }}>
              {[
                { type: 'KEYWORD', desc: 'Palabras clave (def, if, else, return, print)' },
                { type: 'IDENTIFIER', desc: 'Identificadores (nombres de variables y funciones)' },
                { type: 'NUMBER', desc: 'Números literales' },
                { type: 'STRING', desc: 'Cadenas de texto' },
                { type: 'OPERATOR', desc: 'Operadores (<=, *, -, =)' },
                { type: 'PARENTHESIS', desc: 'Paréntesis ( )' },
                { type: 'COMMA', desc: 'Comas' },
                { type: 'COLON', desc: 'Dos puntos :' }
              ].map(({ type, desc }) => (
                <div key={type} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{
                    ...tokenStyle,
                    ...getTokenStyle(type)
                  }}>
                    {type}
                  </span>
                  <span style={{ fontSize: '14px', color: '#6b7280' }}>{desc}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default FactorialAnalyzer;