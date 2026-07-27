import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { Provider } from 'react-redux'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import * as itemsAPI from './api/items'
import App from './App'
import { uiText } from './content/uiText'
import { store } from './store/store'

vi.mock('./api/items', () => ({
  getItems: vi.fn(),
  createItem: vi.fn(),
  updateItem: vi.fn(),
  deleteItem: vi.fn(),
}))

function renderApp() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </Provider>,
  )
}

describe('App query states', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows a loading state', () => {
    vi.mocked(itemsAPI.getItems).mockReturnValue(new Promise(() => {}))

    renderApp()

    expect(screen.getByText(uiText.list.loading)).toBeInTheDocument()
  })

  it('shows the empty state after loading', async () => {
    vi.mocked(itemsAPI.getItems).mockResolvedValue([])

    renderApp()

    expect(await screen.findByText(uiText.list.emptyTitle)).toBeInTheDocument()
  })

  it('shows an API error state', async () => {
    vi.mocked(itemsAPI.getItems).mockRejectedValue(
      new Error('backend unavailable'),
    )

    renderApp()

    expect(
      await screen.findByText(uiText.list.loadErrorTitle),
    ).toBeInTheDocument()
    expect(screen.getByText('backend unavailable')).toBeInTheDocument()
  })
})
