import type { ReactNode } from 'react'
import { Button } from './Button'

interface StatusPanelProps {
  title?: string
  description: string
  icon?: ReactNode
  loading?: boolean
  actionLabel?: string
  onAction?: () => void
  role?: 'alert' | 'status'
}

export function StatusPanel({
  title,
  description,
  icon,
  loading = false,
  actionLabel,
  onAction,
  role = 'status',
}: StatusPanelProps) {
  return (
    <div className="status-panel" role={role}>
      {loading && <span className="spinner" aria-hidden="true" />}
      {icon && <div className="status-panel__icon">{icon}</div>}
      {title && <h3>{title}</h3>}
      <p>{description}</p>
      {actionLabel && onAction && (
        <Button onClick={onAction}>{actionLabel}</Button>
      )}
    </div>
  )
}
