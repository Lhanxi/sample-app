import { createSlice, type PayloadAction } from '@reduxjs/toolkit'

interface UIState {
  editingItemId: string | null
}

const initialState: UIState = {
  editingItemId: null,
}

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    startEditing(state, action: PayloadAction<string>) {
      state.editingItemId = action.payload
    },
    stopEditing(state) {
      state.editingItemId = null
    },
  },
})

export const { startEditing, stopEditing } = uiSlice.actions
export default uiSlice.reducer
