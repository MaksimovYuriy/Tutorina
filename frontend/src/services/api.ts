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

export interface TeacherProfile {
  id: string
  userId: number | null
  displayName: string
  education: string
  experience: string
  approach: string
  photoUrl: string
  isPublished: boolean
}

export type TeacherProfileInput = Pick<TeacherProfile, 'displayName' | 'education' | 'experience' | 'approach' | 'isPublished'>

export type OfferFormat = 'online' | 'offline' | 'both'

export interface TeacherOffer {
  id: string
  offerId: string
  teacherProfileId: string
  teacherDisplayName: string
  durationMinutes: number | null
  priceRubles: number | null
  isPublished: boolean
}

export interface Offer {
  id: string
  title: string
  description: string
  goal: string
  defaultDurationMinutes: number
  format: OfferFormat
  priceRubles: number | null
  isPublished: boolean
  teachers: TeacherOffer[]
}

export type OfferInput = Omit<Offer, 'id' | 'teachers'>

export interface TeacherOfferInput {
  teacherProfileId: number
  durationMinutes: number | null
  priceRubles: number | null
  isPublished: boolean
}

interface TeacherProfilesDocument {
  data: Array<{ id: string; attributes: Omit<TeacherProfile, 'id'> }>
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

export async function getTeacherProfiles(signal?: AbortSignal): Promise<TeacherProfile[]> {
  return getProfiles('/api/v1/admin/teachers/', signal)
}

export async function getPublicTeacherProfiles(signal?: AbortSignal): Promise<TeacherProfile[]> {
  const document = await request<{ data: Array<{ id: string; attributes: Pick<TeacherProfile, 'displayName' | 'education' | 'experience' | 'approach' | 'photoUrl'> }> }>('/api/v1/teachers', { signal })
  return document.data.map(({ id, attributes }) => ({ id, userId: null, isPublished: true, ...attributes }))
}

export async function createTeacherProfile(profile: TeacherProfileInput): Promise<TeacherProfile> {
  return sendTeacherProfile('/api/v1/admin/teachers/', 'POST', profile)
}

export async function updateTeacherProfile(profileId: string, profile: TeacherProfileInput): Promise<TeacherProfile> {
  return sendTeacherProfile(`/api/v1/admin/teachers/${profileId}`, 'PUT', profile)
}

export async function uploadTeacherPhoto(profileId: string, photo: File): Promise<TeacherProfile> {
  const body = new FormData()
  body.append('photo', photo)
  const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>(`/api/v1/admin/teachers/${profileId}/photo`, {
    method: 'PUT',
    body,
  })
  return { id: document.data.id, ...document.data.attributes }
}

export async function removeTeacherPhoto(profileId: string): Promise<TeacherProfile> {
  const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>(`/api/v1/admin/teachers/${profileId}/photo`, {
    method: 'DELETE',
  })
  return { id: document.data.id, ...document.data.attributes }
}

export async function getMyTeacherProfile(signal?: AbortSignal): Promise<TeacherProfile> {
	const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>('/api/v1/teacher/profile', { signal })
	return { id: document.data.id, ...document.data.attributes }
}

export async function updateMyTeacherProfile(profile: TeacherProfileInput): Promise<TeacherProfile> {
	return sendTeacherProfile('/api/v1/teacher/profile', 'PUT', profile)
}

export async function uploadMyTeacherPhoto(photo: File): Promise<TeacherProfile> {
	const body = new FormData()
	body.append('photo', photo)
	const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>('/api/v1/teacher/profile/photo', { method: 'PUT', body })
	return { id: document.data.id, ...document.data.attributes }
}

export async function removeMyTeacherPhoto(): Promise<TeacherProfile> {
	const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>('/api/v1/teacher/profile/photo', { method: 'DELETE' })
	return { id: document.data.id, ...document.data.attributes }
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
	await request<void>('/api/v1/auth/password', {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ currentPassword, newPassword }),
	})
}

export async function resetTeacherPassword(profileId: string, password: string): Promise<void> {
	await request<void>(`/api/v1/admin/teachers/${profileId}/password`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ password }),
	})
}
export async function createTeacherAccount(profileId: string, email: string, password: string): Promise<TeacherProfile> {
  const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>(`/api/v1/admin/teachers/${profileId}/account`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, password }),
  })
  return { id: document.data.id, ...document.data.attributes }
}

export async function archiveTeacherProfile(profileId: string): Promise<void> {
  await request<void>(`/api/v1/admin/teachers/${profileId}`, { method: 'DELETE' })
}

export async function getOffers(signal?: AbortSignal): Promise<Offer[]> {
  const document = await request<{ data: Array<{ id: string; attributes: Omit<Offer, 'id'> }> }>('/api/v1/admin/offers/', { signal })
  return document.data.map(({ id, attributes }) => ({ id, ...attributes }))
}

export async function getPublicOffers(signal?: AbortSignal): Promise<Offer[]> {
  const document = await request<{ data: Array<{ id: string; attributes: Omit<Offer, 'id'> }> }>('/api/v1/offers', { signal })
  return document.data.map(({ id, attributes }) => ({ id, ...attributes }))
}

export async function createOffer(offer: OfferInput): Promise<Offer> {
  return sendOffer('/api/v1/admin/offers/', 'POST', offer)
}

export async function updateOffer(offerId: string, offer: OfferInput): Promise<Offer> {
  return sendOffer(`/api/v1/admin/offers/${offerId}`, 'PUT', offer)
}

export async function archiveOffer(offerId: string): Promise<void> {
  await request<void>(`/api/v1/admin/offers/${offerId}`, { method: 'DELETE' })
}

export async function assignTeacherToOffer(offerId: string, assignment: TeacherOfferInput): Promise<TeacherOffer> {
  const document = await request<{ data: { id: string; attributes: Omit<TeacherOffer, 'id'> } }>(`/api/v1/admin/offers/${offerId}/teachers`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(assignment),
  })
  return { id: document.data.id, ...document.data.attributes }
}

export async function updateTeacherOffer(assignmentId: string, assignment: TeacherOfferInput): Promise<TeacherOffer> {
  const document = await request<{ data: { id: string; attributes: Omit<TeacherOffer, 'id'> } }>(`/api/v1/admin/teacher-offers/${assignmentId}`, {
    method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(assignment),
  })
  return { id: document.data.id, ...document.data.attributes }
}

export async function removeTeacherFromOffer(assignmentId: string): Promise<void> {
  await request<void>(`/api/v1/admin/teacher-offers/${assignmentId}`, { method: 'DELETE' })
}

async function getProfiles(url: string, signal?: AbortSignal): Promise<TeacherProfile[]> {
  const document = await request<TeacherProfilesDocument>(url, { signal })
  return document.data.map(({ id, attributes }) => ({ id, ...attributes }))
}

async function sendTeacherProfile(url: string, method: 'POST' | 'PUT', profile: TeacherProfileInput): Promise<TeacherProfile> {
  const document = await request<{ data: { id: string; attributes: Omit<TeacherProfile, 'id'> } }>(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(profile),
  })
  return { id: document.data.id, ...document.data.attributes }
}

async function sendOffer(url: string, method: 'POST' | 'PUT', offer: OfferInput): Promise<Offer> {
  const document = await request<{ data: { id: string; attributes: Omit<Offer, 'id'> } }>(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(offer),
  })
  return { id: document.data.id, ...document.data.attributes }
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
