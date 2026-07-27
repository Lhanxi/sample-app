import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { uiText } from '../content/uiText'
import { ItemList } from './ItemList'

describe('ItemList', () => {
  it('renders the empty state', () => {
    render(
      <ItemList
        items={[]}
        deletingItemId={null}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    )

    expect(screen.getByText(uiText.list.emptyTitle)).toBeInTheDocument()
  })

  it('renders items and exposes edit and delete actions', async () => {
    const user = userEvent.setup()
    const onEdit = vi.fn()
    const onDelete = vi.fn()

    render(
      <ItemList
        items={[
          {
            id: '1',
            name: 'Desk',
            description: 'Standing desk',
            created_at: '2026-01-01T00:00:00Z',
            updated_at: '2026-01-01T00:00:00Z',
          },
        ]}
        deletingItemId={null}
        onEdit={onEdit}
        onDelete={onDelete}
      />,
    )

    expect(screen.getByText('Desk')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: uiText.list.edit }))
    await user.click(screen.getByRole('button', { name: uiText.list.delete }))

    expect(onEdit).toHaveBeenCalledWith('1')
    expect(onDelete).toHaveBeenCalledWith('1')
  })
})
