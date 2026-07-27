import type { Item, ItemInput } from '../types/item'

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as {
      error?: string
    } | null

    throw new Error(
      payload?.error ?? `Request failed with status ${response.status}`,
    )
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}

export function getItems(): Promise<Item[]> {
  return request<Item[]>('/api/v1/items')
}

export function createItem(input: ItemInput): Promise<Item> {
  return request<Item>('/api/v1/items', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateItem(id: string, input: ItemInput): Promise<Item> {
  return request<Item>(`/api/v1/items/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export function deleteItem(id: string): Promise<void> {
  return request<void>(`/api/v1/items/${id}`, {
    method: 'DELETE',
  })
}
