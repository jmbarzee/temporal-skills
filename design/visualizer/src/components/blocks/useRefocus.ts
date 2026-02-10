import React from 'react'

/**
 * Context for requesting focus return to editor after webview interactions.
 * Provided by webview.tsx, defaults to no-op for standalone app.
 */
export const RefocusContext = React.createContext<() => void>(() => {})

/**
 * Hook to request focus return to the editor after webview interactions.
 * Returns a no-op function in standalone mode (when context is not provided).
 */
export function useRefocus(): () => void {
  return React.useContext(RefocusContext)
}
