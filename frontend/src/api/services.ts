// Klienci zasobów REST API – odpowiadają kontrolerom (handlerom) backendu.
import type { HttpClient } from './httpClient'
import type { Currency, LoginResponse, User } from '../types'

export class AuthApi {
  private readonly http: HttpClient
  constructor(http: HttpClient) {
    this.http = http
  }
  register(email: string, password: string) {
    return this.http.post<User>('/auth/register', { email, password })
  }
  login(email: string, password: string) {
    return this.http.post<LoginResponse>('/auth/login', { email, password })
  }
  logout() {
    return this.http.post<void>('/auth/logout')
  }
  me() {
    return this.http.get<User>('/me')
  }
  setBaseCurrency(currency: Currency) {
    return this.http.put<User>('/me/currency', { currency })
  }
}
