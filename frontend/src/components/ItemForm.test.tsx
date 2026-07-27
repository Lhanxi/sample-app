import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { uiText } from '../content/uiText'
import { ItemForm } from './ItemForm'

describe('ItemForm', () => {
  it('rejects a blank name', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()

    render(
      <ItemForm isSubmitting={false} onSubmit={onSubmit} onCancel={vi.fn()} />,
    )

    await user.click(screen.getByRole('button', { name: uiText.form.create }))

    expect(screen.getByRole('alert')).toHaveTextContent(
      uiText.form.nameRequired,
    )
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('trims and submits valid input', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn().mockResolvedValue(undefined)

    render(
      <ItemForm isSubmitting={false} onSubmit={onSubmit} onCancel={vi.fn()} />,
    )

    await user.type(screen.getByLabelText(uiText.form.nameLabel), '  Desk  ')
    await user.type(
      screen.getByLabelText(uiText.form.descriptionLabel),
      '  Standing desk  ',
    )
    await user.click(screen.getByRole('button', { name: uiText.form.create }))

    expect(onSubmit).toHaveBeenCalledWith({
      name: 'Desk',
      description: 'Standing desk',
    })
  })

  it('populates edit values and cancels editing', async () => {
    const user = userEvent.setup()
    const onCancel = vi.fn()

    render(
      <ItemForm
        item={{
          id: '1',
          name: 'Desk',
          description: 'Standing desk',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
        }}
        isSubmitting={false}
        onSubmit={vi.fn()}
        onCancel={onCancel}
      />,
    )

    expect(screen.getByLabelText(uiText.form.nameLabel)).toHaveValue('Desk')
    await user.click(screen.getByRole('button', { name: uiText.form.cancel }))
    expect(onCancel).toHaveBeenCalledOnce()
  })
})
