import { uiText } from '../content/uiText'
import type { Item } from '../types/item'
import { Button } from './common/Button'

interface ItemRowProps {
  item: Item
  isDeleting: boolean
  onEdit: (id: string) => void
  onDelete: (id: string) => void
}

export function ItemRow({ item, isDeleting, onEdit, onDelete }: ItemRowProps) {
  return (
    <article className="item-row">
      <div className="item-copy">
        <div className="item-title-line">
          <h3>{item.name}</h3>
          <span>
            {uiText.list.updated}{' '}
            {new Intl.DateTimeFormat(undefined, {
              dateStyle: 'medium',
            }).format(new Date(item.updated_at))}
          </span>
        </div>
        <p>{item.description || uiText.list.noDescription}</p>
      </div>

      <div className="item-actions" aria-label={`Actions for ${item.name}`}>
        <Button onClick={() => onEdit(item.id)}>{uiText.list.edit}</Button>
        <Button
          variant="danger"
          onClick={() => onDelete(item.id)}
          disabled={isDeleting}
        >
          {isDeleting ? uiText.list.deleting : uiText.list.delete}
        </Button>
      </div>
    </article>
  )
}
