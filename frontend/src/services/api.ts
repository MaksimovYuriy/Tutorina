export interface ApiStatus {
  status: string
  database: string
  time: string
}

export async function getApiStatus(signal?: AbortSignal): Promise<ApiStatus> {
  const response = await fetch('/api/v1/ping', { signal })
  if (!response.ok) {
    throw new Error(`API вернул статус ${response.status}`)
  }
  return response.json() as Promise<ApiStatus>
}

