import React from 'react'
import { WorkflowCanvas } from './components/WorkflowCanvas'
import type { TWFFile } from './types/ast'

// Standalone app - for development/testing
// Load AST from URL query param: ?ast=/path/to/file.json

function App() {
  const [ast, setAst] = React.useState<TWFFile | null>(null)
  const [error, setError] = React.useState<string | null>(null)
  const [loading, setLoading] = React.useState(false)

  // Load AST from query param or postMessage
  React.useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const message = event.data
      if (message.type === 'ast') {
        setAst(message.data)
        setError(null)
      } else if (message.type === 'error') {
        setError(message.message)
        setAst(null)
      }
    }

    window.addEventListener('message', handleMessage)

    // Check for ?ast= query param
    const params = new URLSearchParams(window.location.search)
    const astPath = params.get('ast')
    if (astPath) {
      setLoading(true)
      fetch(astPath)
        .then(res => res.json())
        .then(data => {
          setAst(data)
          setLoading(false)
        })
        .catch(err => {
          setError(err.message)
          setLoading(false)
        })
    }

    return () => window.removeEventListener('message', handleMessage)
  }, [])

  // File input handler for manual loading
  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const json = JSON.parse(e.target?.result as string)
        setAst(json)
        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to parse JSON')
      }
    }
    reader.readAsText(file)
  }

  if (loading) {
    return (
      <div className="loading-container">
        <p>Loading workflow...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="error-container">
        <h2>Error parsing workflow</h2>
        <pre>{error}</pre>
      </div>
    )
  }

  if (!ast) {
    return (
      <div className="empty-container">
        <div className="empty-content">
          <h2>TWF Workflow Visualizer</h2>
          <p>Load an AST JSON file to visualize:</p>
          <label className="file-upload-btn">
            Choose File
            <input type="file" accept=".json" onChange={handleFileUpload} />
          </label>
          <p className="hint">
            Generate AST with: <code>parse --json file.twf &gt; ast.json</code>
          </p>
        </div>
      </div>
    )
  }

  return <WorkflowCanvas ast={ast} />
}

export default App
