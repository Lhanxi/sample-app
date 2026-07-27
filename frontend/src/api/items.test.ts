import { afterEach, describe, expect, it, vi } from 'vitest'
import { createItem, deleteItem, getItems, updateItem } from './items'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('items API', () => {
  it('retrieves items', async () => {
    const items = [{ id: '1', name: 'Desk', description: '' }]
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(items), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    vi.stubGlobal('fetch', fetchMock)

    await expect(getItems()).resolves.toEqual(items)
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/items',
      expect.objectContaining({
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
        }),
      }),
    )
  })

  it('creates and updates an item with JSON requests', async () => {
    const item = { id: '1', name: 'Desk', description: 'Standing' }
    const fetchMock = vi.fn().mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify(item), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )
    vi.stubGlobal('fetch', fetchMock)

    await createItem({ name: 'Desk', description: 'Standing' })
    await updateItem('1', { name: 'Desk', description: 'Updated' })

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      'http://localhost:8080/api/v1/items',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ name: 'Desk', description: 'Standing' }),
      }),
    )
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      'http://localhost:8080/api/v1/items/1',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ name: 'Desk', description: 'Updated' }),
      }),
    )
  })

  it('handles a no-content deletion response', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(deleteItem('1')).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/items/1',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('uses the API error message when a request fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: 'item not found' }), {
          status: 404,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    await expect(getItems()).rejects.toThrow('item not found')
  })
})
