import React from 'react'

export function useToggle(
  initialState: boolean = false,
  canToggle: boolean = true,
): [boolean, () => void] {
  const [expanded, setExpanded] = React.useState(initialState)

  const toggle = () => {
    if (canToggle) { setExpanded(prev => !prev) }
  }

  return [expanded, toggle]
}
