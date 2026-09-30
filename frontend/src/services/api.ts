export interface ApiStatus {
  status: string
  database: string
  time: string
}

export type UserRole = 'teacher' | 'admin'

export interface CurrentUser {
  id: string
  email: string
  roles: UserRole[]
}

interface UserDocument {
  data: {
    type: 'users'
    id: string
    attributes: {
      email: string
      roles: UserRole[]
    }
  }
}

interface ErrorDocument {
  errors?: Array<{
    title?: string
    detail?: string
  }>
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

export async function getApiStatus(signal?: AbortSignal): Promise<ApiStatus> {
  return request<ApiStatus>('/api/v1/ping', { signal })
}

export async function login(email: string, password: string): Promise<void> {
  await request<void>('/api/v1/auth/sessions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
}

export async function getCurrentUser(signal?: AbortSignal): Promise<CurrentUser> {
  const document = await request<UserDocument>('/api/v1/auth/me', { signal })
  return {
    id: document.data.id,
    email: document.data.attributes.email,
    roles: document.data.attributes.roles,
  }
}

export async function logout(): Promise<void> {
  await request<void>('/api/v1/auth/session', { method: 'DELETE' })
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  if (!response.ok) {
    let message = 'Не удалось выполнить запрос. Попробуйте ещё раз.'
    try {
      const document = (await response.json()) as ErrorDocument
      const error = document.errors?.[0]
      message = error?.detail || error?.title || message
    } catch {
      // The fallback message is more useful than a response parsing error.
    }
    throw new ApiError(message, response.status)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

