const PRIORITY_DOT = {
  low: 'priority-low',
  medium: 'priority-medium',
  high: 'priority-high',
  urgent: 'priority-urgent'
}

/** Maps a priority name to its colored-dot class (defaults to Low/medium/high; custom -> neutral). */
export function priorityDotClass (priorityName) {
  return PRIORITY_DOT[(priorityName || '').toLowerCase()] || 'priority-none'
}
