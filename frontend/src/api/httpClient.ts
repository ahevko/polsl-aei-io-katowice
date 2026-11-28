// HttpClient – jedyne miejsce, w którym frontend komunikuje się z backendem
// (fetch + nagłówek Authorization + jednolita obsługa błędów).

export class ApiError extends Error {
  readonly status: number
  readonly field?: string

  constructor(status: number, message: string, field?: string) {
    super(message)
    this.status = status
    this.field = field
  }
}

export class HttpClient {
  private readonly baseUrl: string
  private readonly getToken: () => string | null
  private readonly onUnauthorized: () => void

  constructor(baseUrl: string, getToken: () => string | null, onUnauthorized: () => void) {
    this.baseUrl = baseUrl
    this.getToken = getToken
    this.onUnauthorized = onUnauthorized
  }

  async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const headers: Record<string, string> = {}
    const token = this.getToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
    if (body !== undefined) headers['Content-Type'] = 'application/json'

    let res: Response
    try {
      res = await fetch(this.baseUrl + path, {
        method,
        headers,
        body: body === undefined ? undefined : JSON.stringify(body),
      })
    } catch {
      throw new ApiError(0, 'Brak połączenia z serwerem')
    }

    if (res.status === 401 && token) this.onUnauthorized()
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new ApiError(res.status, err.error ?? 'Błąd serwera', err.field)
    }
    if (res.status === 204) return undefined as T
    return (await res.json()) as T
  }

  get<T>(path: string) {
    return this.request<T>('GET', path)
  }
  post<T>(path: string, body?: unknown) {
    return this.request<T>('POST', path, body)
  }
  put<T>(path: string, body?: unknown) {
    return this.request<T>('PUT', path, body)
  }
  delete(path: string) {
    return this.request<void>('DELETE', path)
  }

  // Pobranie pliku (eksport CSV) z nagłówkiem autoryzacji.
  async download(path: string, filename: string): Promise<void> {
    const token = this.getToken()
    const res = await fetch(this.baseUrl + path, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    if (!res.ok) throw new ApiError(res.status, 'Nie udało się pobrać pliku')
    const url = URL.createObjectURL(await res.blob())
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }
}
