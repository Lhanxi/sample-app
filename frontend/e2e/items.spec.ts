import { expect, test, type Page } from '@playwright/test'

interface Item {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

async function mockItemsAPI(page: Page) {
  let items: Item[] = []

  await page.route('**/api/v1/items**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const id = url.pathname.split('/').at(-1)

    if (request.method() === 'GET') {
      await route.fulfill({ json: items })
      return
    }

    if (request.method() === 'POST') {
      const input = request.postDataJSON() as {
        name: string
        description: string
      }
      const now = new Date().toISOString()
      const item: Item = {
        id: 'b8f574d6-1d0d-4f63-b4a8-4ec847bd9f1d',
        ...input,
        created_at: now,
        updated_at: now,
      }
      items = [item, ...items]
      await route.fulfill({ status: 201, json: item })
      return
    }

    if (request.method() === 'PUT') {
      const input = request.postDataJSON() as {
        name: string
        description: string
      }
      const item = items.find((candidate) => candidate.id === id)
      if (!item) {
        await route.fulfill({ status: 404, json: { error: 'item not found' } })
        return
      }
      Object.assign(item, input, { updated_at: new Date().toISOString() })
      await route.fulfill({ json: item })
      return
    }

    if (request.method() === 'DELETE') {
      items = items.filter((candidate) => candidate.id !== id)
      await route.fulfill({ status: 204, body: '' })
      return
    }

    await route.fallback()
  })
}

test('creates, edits, and deletes an item', async ({ page }) => {
  await mockItemsAPI(page)
  await page.goto('/')

  await expect(page.getByText('No items yet')).toBeVisible()

  await page.getByLabel('Name').fill('Standing desk')
  await page.getByLabel('Description').fill('A desk for focused work')
  await page.getByRole('button', { name: 'Create item' }).click()

  await expect(
    page.getByRole('heading', { name: 'Standing desk' }),
  ).toBeVisible()
  await expect(page.getByText('A desk for focused work')).toBeVisible()

  await page.getByRole('button', { name: 'Edit' }).click()
  await page.getByLabel('Name').fill('Updated desk')
  await page.getByRole('button', { name: 'Save changes' }).click()

  await expect(
    page.getByRole('heading', { name: 'Updated desk' }),
  ).toBeVisible()

  page.once('dialog', (dialog) => dialog.accept())
  await page.getByRole('button', { name: 'Delete' }).click()

  await expect(page.getByText('No items yet')).toBeVisible()
})
