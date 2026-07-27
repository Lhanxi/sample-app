import { useEffect, useState, type FormEvent } from 'react'
import { uiText } from '../content/uiText'
import type { Item, ItemInput } from '../types/item'
import { Button } from './common/Button'

interface ItemFormProps {
  item?: Item
  isSubmitting: boolean
  onSubmit: (input: ItemInput) => Promise<void>
  onCancel: () => void
}

export function ItemForm({
  item,
  isSubmitting,
  onSubmit,
  onCancel,
}: ItemFormProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [validationError, setValidationError] = useState('')

  useEffect(() => {
    setName(item?.name ?? '')
    setDescription(item?.description ?? '')
    setValidationError('')
  }, [item])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const trimmedName = name.trim()
    const trimmedDescription = description.trim()

    if (!trimmedName) {
      setValidationError(uiText.form.nameRequired)
      return
    }

    setValidationError('')
    try {
      await onSubmit({
        name: trimmedName,
        description: trimmedDescription,
      })
    } catch {
      return
    }

    if (!item) {
      setName('')
      setDescription('')
    }
  }

  return (
    <section className="form-card" aria-labelledby="item-form-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">
            {item ? uiText.form.editEyebrow : uiText.form.newEyebrow}
          </p>
          <h2 id="item-form-title">
            {item ? uiText.form.editTitle : uiText.form.createTitle}
          </h2>
        </div>
        {item && (
          <Button variant="text" onClick={onCancel}>
            {uiText.form.cancel}
          </Button>
        )}
      </div>

      <form onSubmit={handleSubmit} noValidate>
        <label htmlFor="item-name">{uiText.form.nameLabel}</label>
        <input
          id="item-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder={uiText.form.namePlaceholder}
          disabled={isSubmitting}
          aria-describedby={validationError ? 'name-error' : undefined}
          aria-invalid={Boolean(validationError)}
          autoFocus
        />

        <label htmlFor="item-description">{uiText.form.descriptionLabel}</label>
        <textarea
          id="item-description"
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          placeholder={uiText.form.descriptionPlaceholder}
          disabled={isSubmitting}
          rows={4}
        />

        {validationError && (
          <p className="field-error" id="name-error" role="alert">
            {validationError}
          </p>
        )}

        <Button variant="primary" type="submit" disabled={isSubmitting}>
          {isSubmitting
            ? uiText.form.saving
            : item
              ? uiText.form.update
              : uiText.form.create}
        </Button>
      </form>
    </section>
  )
}
