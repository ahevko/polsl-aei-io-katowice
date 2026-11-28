// Klienci zasobów REST API – odpowiadają kontrolerom (handlerom) backendu.
import type { HttpClient } from './httpClient'
import type { Category, CostSummary, Currency, LoginResponse, Subscription, SubscriptionInput, User } from '../types'

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

export class SubscriptionApi {
  private readonly http: HttpClient
  constructor(http: HttpClient) {
    this.http = http
  }
  list() {
    return this.http.get<Subscription[]>('/subscriptions')
  }
  categories() {
    return this.http.get<Category[]>('/categories')
  }
  create(input: SubscriptionInput) {
    return this.http.post<Subscription>('/subscriptions', input)
  }
  update(id: string, input: SubscriptionInput) {
    return this.http.put<Subscription>(`/subscriptions/${id}`, input)
  }
  remove(id: string) {
    return this.http.delete(`/subscriptions/${id}`)
  }
  pause(id: string) {
    return this.http.post<Subscription>(`/subscriptions/${id}/pause`)
  }
  resume(id: string, nextPaymentDate: string) {
    return this.http.post<Subscription>(`/subscriptions/${id}/resume`, { nextPaymentDate })
  }
}

export class AnalyticsApi {
  private readonly http: HttpClient
  constructor(http: HttpClient) {
    this.http = http
  }
  summary() {
    return this.http.get<CostSummary>('/analytics/summary')
  }
}
