import { uiText } from '../content/uiText'
import type { Item } from '../types/item'
import { StatusPanel } from './common/StatusPanel'
import { ItemRow } from './ItemRow'

interface ItemListProps {
  items: Item[]
  deletingItemId: string | null
  onEdit: (id: string) => void
  onDelete: (id: string) => void
}

export function ItemList({
  items,
  deletingItemId,
  onEdit,
  onDelete,
}: ItemListProps) {
  if (items.length === 0) {
    return (
      <StatusPanel
        icon="+"
        title={uiText.list.emptyTitle}
        description={uiText.list.emptyDescription}
      />
    )
  }

  return (
    <div className="item-list">
      {items.map((item) => (
        <ItemRow
          key={item.id}
          item={item}
          isDeleting={deletingItemId === item.id}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </div>
  )
}
