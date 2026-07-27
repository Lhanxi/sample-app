import { describe, expect, it } from 'vitest'
import reducer, { startEditing, stopEditing } from './uiSlice'

describe('UI state', () => {
  it('starts and stops item editing', () => {
    const editingState = reducer(undefined, startEditing('item-1'))
    expect(editingState.editingItemId).toBe('item-1')

    const idleState = reducer(editingState, stopEditing())
    expect(idleState.editingItemId).toBeNull()
  })
})
