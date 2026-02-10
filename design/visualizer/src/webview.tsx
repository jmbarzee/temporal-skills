import React from 'react'
import ReactDOM from 'react-dom/client'
import { WorkflowCanvas } from './components/WorkflowCanvas'
import type { TWFFile } from './types/ast'
import './styles/index.css'

// VSCode webview entry point
declare const acquireVsCodeApi: () => {
  postMessage: (msg: unknown) => void
  getState: () => unknown
  setState: (state: unknown) => void
}

const vscode = acquireVsCodeApi()

function WebviewApp() {
  const [ast, setAst] = React.useState<TWFFile | null>(null)
  const [error, setError] = React.useState<string | null>(null)

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

    // Request initial data
    vscode.postMessage({ type: 'ready' })

    return () => window.removeEventListener('message', handleMessage)
  }, [])

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
      <div className="loading-container">
        <p>Loading workflow...</p>
      </div>
    )
  }

  return <WorkflowCanvas ast={ast} />
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <WebviewApp />
  </React.StrictMode>,
)
