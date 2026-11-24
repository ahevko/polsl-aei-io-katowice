// Typy DTO - odpowiadają strukturom z backend/internal/api/dto.go.

export type Currency = 'PLN' | 'EUR' | 'USD'
export type BillingCycle = 'WEEKLY' | 'MONTHLY' | 'YEARLY'
export type SubscriptionStatus = 'ACTIVE' | 'PAUSED'

export const CURRENCIES: Currency[] = ['PLN', 'EUR', 'USD']

export const CYCLE_LABELS: Record<BillingCycle, string> = {
  WEEKLY: 'tygodniowy',
  MONTHLY: 'miesięczny',
  YEARLY: 'roczny',
}

export interface Money {
  amount: number
  currency: Currency
}

export interface User {
  id: string
  email: string
  baseCurrency: Currency
  role: 'USER' | 'ADMIN'
}

export interface LoginResponse {
  token: string
  user: User
}

export interface Category {
  id: string
  name: string
  isDefault: boolean
}

export interface Subscription {
  id: string
  name: string
  price: Money
  monthlyCost: Money
  cycle: BillingCycle
  nextPaymentDate: string // RRRR-MM-DD
  categoryIds: string[]
  status: SubscriptionStatus
}

export interface SubscriptionInput {
  name: string
  amount: number
  currency: Currency
  cycle: BillingCycle
  nextPaymentDate: string
  categoryIds: string[]
}

export interface CategoryCost {
  categoryId: string
  categoryName: string
  monthly: Money
}

export interface CostSummary {
  baseCurrency: Currency
  monthlyTotal: Money
  yearlyTotal: Money
  activeSubscriptions: number
  pausedSubscriptions: number
  byCategory: CategoryCost[]
}

export interface Payment {
  id: string
  subscriptionId: string
  subscriptionName: string
  date: string
  amount: Money
  converted: Money
}

export interface WorkerReport {
  day: string
  paymentsRecorded: number
  remindersSent: number
}

export interface EmailMessage {
  to: string
  subject: string
  body: string
  sentAt: string
}

export function formatMoney(m: Money): string {
  return new Intl.NumberFormat('pl-PL', { style: 'currency', currency: m.currency }).format(m.amount)
}

export function formatDate(iso: string): string {
  const [y, mo, d] = iso.split('-')
  return `${d}.${mo}.${y}`
}

export function todayISO(offsetDays = 0): string {
  const d = new Date()
  d.setDate(d.getDate() + offsetDays)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
