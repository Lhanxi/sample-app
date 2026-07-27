export interface Item {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface ItemInput {
  name: string
  description: string
}
