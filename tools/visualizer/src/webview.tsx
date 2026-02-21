import React from 'react'
import ReactDOM from 'react-dom/client'
import { WorkflowCanvas } from './components/WorkflowCanvas'
import { StyleGuide } from './components/StyleGuide'
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
  const [showStyleGuide, setShowStyleGuide] = React.useState(false)

  // Ctrl+Shift+G toggles style guide
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.shiftKey && e.key === 'G') {
        e.preventDefault()
        setShowStyleGuide(prev => !prev)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

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

  // Request focus return to the editor after user interaction
  const requestRefocus = React.useCallback(() => {
    vscode.postMessage({ type: 'refocus' })
  }, [])

  // Open a file in the editor when the file filter narrows to one
  const openFile = React.useCallback((file: string) => {
    vscode.postMessage({ type: 'openFile', file })
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

  if (showStyleGuide) {
    return <StyleGuide onClose={() => setShowStyleGuide(false)} />
  }

  return (
    <div onClick={requestRefocus}>
      <WorkflowCanvas ast={ast} onOpenFile={openFile} />
    </div>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <WebviewApp />
  </React.StrictMode>,
)
