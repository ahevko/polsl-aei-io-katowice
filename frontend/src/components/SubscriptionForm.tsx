// Formularz dodawania/edycji subskrypcji – jeden ekran (wymaganie użyteczności:
// dodanie subskrypcji w maksymalnie dwóch ekranach formularza).
import { useState } from 'react'
import type { FormEvent } from 'react'
import { ApiError } from '../api/httpClient'
import { CURRENCIES, CYCLE_LABELS, todayISO } from '../types'
import type { BillingCycle, Category, Currency, Subscription, SubscriptionInput } from '../types'

interface Props {
  categories: Category[]
  initial?: Subscription
  onSubmit: (input: SubscriptionInput) => Promise<void>
  onCancel: () => void
}

export function SubscriptionForm({ categories, initial, onSubmit, onCancel }: Props) {
  const [name, setName] = useState(initial?.name ?? '')
  const [amount, setAmount] = useState(initial ? String(initial.price.amount) : '')
  const [currency, setCurrency] = useState<Currency>(initial?.price.currency ?? 'PLN')
  const [cycle, setCycle] = useState<BillingCycle>(initial?.cycle ?? 'MONTHLY')
  const [nextPaymentDate, setNext] = useState(initial?.nextPaymentDate ?? todayISO(1))
  const [categoryIds, setCategoryIds] = useState<string[]>(initial?.categoryIds ?? [])
  const [error, setError] = useState<{ field?: string; message: string } | null>(null)
  const [busy, setBusy] = useState(false)

  function toggleCategory(id: string) {
    setCategoryIds((ids) => (ids.includes(id) ? ids.filter((x) => x !== id) : [...ids, id]))
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    // walidacja wstępna w warstwie prezentacji (WSS)
    const value = Number(amount.replace(',', '.'))
    if (!Number.isFinite(value) || value <= 0) {
      setError({ field: 'amount', message: 'Podaj koszt większy od zera' })
      return
    }
    if (categoryIds.length === 0) {
      setError({ field: 'categoryIds', message: 'Wybierz co najmniej jedną kategorię' })
      return
    }
    setBusy(true)
    try {
      await onSubmit({ name, amount: value, currency, cycle, nextPaymentDate, categoryIds })
    } catch (err) {
      setError(err instanceof ApiError ? { field: err.field, message: err.message } : { message: 'Nieznany błąd' })
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="card form-grid" onSubmit={submit}>
      <h2>{initial ? `Edytuj: ${initial.name}` : 'Nowa subskrypcja'}</h2>
      <label className="span-2">
        Nazwa
        <input value={name} onChange={(e) => setName(e.target.value)} required maxLength={100} placeholder="np. Netflix" />
      </label>
      <label>
        Koszt
        <input inputMode="decimal" value={amount} onChange={(e) => setAmount(e.target.value)} required placeholder="49,99" />
      </label>
      <label>
        Waluta
        <select value={currency} onChange={(e) => setCurrency(e.target.value as Currency)}>
          {CURRENCIES.map((c) => (
            <option key={c}>{c}</option>
          ))}
        </select>
      </label>
      <label>
        Cykl rozliczeniowy
        <select value={cycle} onChange={(e) => setCycle(e.target.value as BillingCycle)}>
          {(Object.keys(CYCLE_LABELS) as BillingCycle[]).map((c) => (
            <option key={c} value={c}>
              {CYCLE_LABELS[c]}
            </option>
          ))}
        </select>
      </label>
      <label>
        Data następnej płatności
        <input type="date" value={nextPaymentDate} onChange={(e) => setNext(e.target.value)} required />
      </label>
      <fieldset className="span-2">
        <legend>Kategorie</legend>
        <div className="chips">
          {categories.map((c) => (
            <button
              type="button"
              key={c.id}
              className={categoryIds.includes(c.id) ? 'chip selected' : 'chip'}
              onClick={() => toggleCategory(c.id)}
            >
              {c.name}
            </button>
          ))}
        </div>
      </fieldset>
      {error && <div className="error span-2">{error.message}</div>}
      <div className="actions span-2">
        <button type="button" onClick={onCancel}>
          Anuluj
        </button>
        <button className="primary" disabled={busy}>
          {initial ? 'Zapisz zmiany' : 'Dodaj subskrypcję'}
        </button>
      </div>
    </form>
  )
}
