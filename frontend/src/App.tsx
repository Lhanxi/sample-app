import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { createItem, deleteItem, getItems, updateItem } from './api/items'
import { ItemForm } from './components/ItemForm'
import { ItemList } from './components/ItemList'
import { Button } from './components/common/Button'
import { StatusPanel } from './components/common/StatusPanel'
import { uiText } from './content/uiText'
import { useAppDispatch, useAppSelector } from './store/hooks'
import { startEditing, stopEditing } from './store/uiSlice'
import type { ItemInput } from './types/item'
import './App.css'

function App() {
  const dispatch = useAppDispatch()
  const queryClient = useQueryClient()
  const editingItemId = useAppSelector((state) => state.ui.editingItemId)
  const [deletingItemId, setDeletingItemId] = useState<string | null>(null)
  const [actionError, setActionError] = useState('')

  const itemsQuery = useQuery({
    queryKey: ['items'],
    queryFn: getItems,
  })

  const createMutation = useMutation({
    mutationFn: createItem,
    onSuccess: async () => {
      setActionError('')
      await queryClient.invalidateQueries({ queryKey: ['items'] })
    },
    onError: (error) => setActionError(error.message),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: ItemInput }) =>
      updateItem(id, input),
    onSuccess: async () => {
      dispatch(stopEditing())
      setActionError('')
      await queryClient.invalidateQueries({ queryKey: ['items'] })
    },
    onError: (error) => setActionError(error.message),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteItem,
    onMutate: (id) => {
      setDeletingItemId(id)
      setActionError('')
    },
    onSuccess: async (_, id) => {
      if (editingItemId === id) {
        dispatch(stopEditing())
      }
      await queryClient.invalidateQueries({ queryKey: ['items'] })
    },
    onError: (error) => setActionError(error.message),
    onSettled: () => setDeletingItemId(null),
  })

  const editingItem = itemsQuery.data?.find((item) => item.id === editingItemId)

  async function handleSubmit(input: ItemInput) {
    setActionError('')
    if (editingItem) {
      await updateMutation.mutateAsync({ id: editingItem.id, input })
      return
    }
    await createMutation.mutateAsync(input)
  }

  function handleDelete(id: string) {
    if (window.confirm(uiText.list.deleteConfirmation)) {
      deleteMutation.mutate(id)
    }
  }

  const isSubmitting = createMutation.isPending || updateMutation.isPending

  return (
    <main>
      <header className="page-header">
        <div className="brand-mark" aria-hidden="true">
          S
        </div>
        <div>
          <p className="eyebrow">{uiText.app.brand}</p>
          <h1>{uiText.app.title}</h1>
          <p className="subtitle">{uiText.app.subtitle}</p>
        </div>
      </header>

      <div className="content-grid">
        <ItemForm
          item={editingItem}
          isSubmitting={isSubmitting}
          onSubmit={handleSubmit}
          onCancel={() => dispatch(stopEditing())}
        />

        <section className="list-card" aria-labelledby="items-heading">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{uiText.list.eyebrow}</p>
              <h2 id="items-heading">
                {itemsQuery.data
                  ? `${itemsQuery.data.length} ${
                      itemsQuery.data.length === 1
                        ? uiText.list.singular
                        : uiText.list.plural
                    }`
                  : uiText.list.title}
              </h2>
            </div>
            <Button
              onClick={() => itemsQuery.refetch()}
              disabled={itemsQuery.isFetching}
            >
              {itemsQuery.isFetching
                ? uiText.list.refreshing
                : uiText.list.refresh}
            </Button>
          </div>

          {actionError && (
            <div className="error-banner" role="alert">
              <span>{actionError}</span>
              <Button variant="text" onClick={() => setActionError('')}>
                {uiText.list.dismiss}
              </Button>
            </div>
          )}

          {itemsQuery.isPending && (
            <StatusPanel loading description={uiText.list.loading} />
          )}

          {itemsQuery.isError && (
            <StatusPanel
              title={uiText.list.loadErrorTitle}
              description={itemsQuery.error.message}
              actionLabel={uiText.list.retry}
              onAction={() => itemsQuery.refetch()}
              role="alert"
            />
          )}

          {itemsQuery.data && (
            <ItemList
              items={itemsQuery.data}
              deletingItemId={deletingItemId}
              onEdit={(id) => dispatch(startEditing(id))}
              onDelete={handleDelete}
            />
          )}
        </section>
      </div>
    </main>
  )
}

export default App
