export interface SlotInput {
  directionId: number
  level: string
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
  title: string
  id: number
}
export interface PublicSlot extends Pick<
  Slot,
  'id' | 'title' | 'level' | 'startsAt' | 'endsAt' | 'format' | 'kind'
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
const sessionTokenKey = 'tutorina.session'

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  const token = sessionStorage.getItem(sessionTokenKey)
  if (token && (url.startsWith('/admin/') || url === '/auth/session')) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  const response = await fetch(`/api/v1${url}`, {
    ...init,
    headers,
    credentials: 'omit',
  })
  if (response.status === 401) sessionStorage.removeItem(sessionTokenKey)
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
export async function login(key: string) {
  const session = await request<{ token: string }>('/auth/sessions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key }),
  })
  sessionStorage.setItem(sessionTokenKey, session.token)
}
export async function requireSession(signal?: AbortSignal): Promise<void> {
  await request('/auth/session', { signal })
}
export async function logout() {
  try {
    await request('/auth/session', { method: 'DELETE' })
  } finally {
    sessionStorage.removeItem(sessionTokenKey)
  }
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

export interface Direction {
  id: number
  name: string
}
export async function getDirections(signal?: AbortSignal) {
  return (
    await request<{ data: Direction[] }>('/admin/directions/', { signal })
  ).data
}
export async function saveDirection(name: string, id?: number) {
  return (
    await request<{ data: Direction }>(`/admin/directions/${id ?? ''}`, {
      method: id ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
  ).data
}
export async function deleteDirection(id: number) {
  await request(`/admin/directions/${id}`, { method: 'DELETE' })
}

export const slotStatusLabels = {
  planned: 'Запланирован',
  completed: 'Завершён',
  cancelled: 'Отменён',
}
