export interface CurrentUser {
  id: string
  username: string
  roles: 'admin'[]
}
export interface SlotInput {
  title: string
  startsAt: string
  endsAt: string
  format: 'online' | 'offline'
  kind: 'individual' | 'group'
  capacity: number
  occupied: number
  status: 'planned' | 'completed' | 'cancelled'
  published: boolean
}
export interface Slot extends SlotInput {
  id: number
}
export interface PublicSlot extends Pick<
  Slot,
  'id' | 'title' | 'startsAt' | 'endsAt' | 'format' | 'kind'
> {
  freePlaces: number
}
export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}
async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/v1${url}`, init)
  if (!response.ok) {
    let message = 'Не удалось выполнить запрос. Попробуйте ещё раз.'
    try {
      const body = await response.json()
      message = body.errors?.[0]?.detail || body.errors?.[0]?.title || message
    } catch {
      /* Use fallback for non-JSON responses. */
    }
    throw new ApiError(message, response.status)
  }
  return response.status === 204
    ? (undefined as T)
    : (response.json() as Promise<T>)
}
export async function login(username: string, password: string) {
  await request('/auth/sessions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
}
export async function getCurrentUser(
  signal?: AbortSignal,
): Promise<CurrentUser> {
  const doc = await request<{
    data: { id: string; attributes: Omit<CurrentUser, 'id'> }
  }>('/auth/me', { signal })
  return { id: doc.data.id, ...doc.data.attributes }
}
export async function logout() {
  await request('/auth/session', { method: 'DELETE' })
}
export async function changePassword(
  currentPassword: string,
  newPassword: string,
) {
  await request('/auth/password', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ currentPassword, newPassword }),
  })
}
export async function getPublicSlots(signal?: AbortSignal) {
  return (await request<{ data: PublicSlot[] }>('/slots', { signal })).data
}
export async function getAdminSlots(signal?: AbortSignal) {
  return (await request<{ data: Slot[] }>('/admin/slots/', { signal })).data
}
export async function saveSlot(input: SlotInput, id?: number) {
  return (
    await request<{ data: Slot }>(`/admin/slots/${id ?? ''}`, {
      method: id ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
  ).data
}
export async function deleteSlot(id: number) {
  await request(`/admin/slots/${id}`, { method: 'DELETE' })
}
