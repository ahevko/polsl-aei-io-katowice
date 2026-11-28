// UC: Przeglądaj podsumowanie kosztów, Ustaw walutę domyślną,
// Grupuj wydatki według kategorii (wykres słupkowy – wersja uproszczona P1).
import { useEffect, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { CategoryBars } from '../components/CategoryBars'
import { CURRENCIES, formatMoney } from '../types'
import type { CostSummary, Currency } from '../types'

export function SummaryPage() {
  const { api, user, setBaseCurrency } = useAuth()
  const [summary, setSummary] = useState<CostSummary | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.analytics
      .summary()
      .then(setSummary)
      .catch(() => setError('Nie udało się pobrać podsumowania'))
  }, [api, user?.baseCurrency])

  return (
    <section>
      <header className="page-header">
        <h1>Podsumowanie kosztów</h1>
        <label className="inline">
          Waluta bazowa
          <select value={user?.baseCurrency} onChange={(e) => setBaseCurrency(e.target.value as Currency)}>
            {CURRENCIES.map((c) => (
              <option key={c}>{c}</option>
            ))}
          </select>
        </label>
      </header>
      {error && <div className="error">{error}</div>}
      {summary && (
        <>
          <div className="stats">
            <div className="card stat">
              <span className="muted">Miesięcznie</span>
              <strong>{formatMoney(summary.monthlyTotal)}</strong>
            </div>
            <div className="card stat">
              <span className="muted">Rocznie</span>
              <strong>{formatMoney(summary.yearlyTotal)}</strong>
            </div>
            <div className="card stat">
              <span className="muted">Aktywne / wstrzymane</span>
              <strong>
                {summary.activeSubscriptions} / {summary.pausedSubscriptions}
              </strong>
            </div>
          </div>
          <div className="card">
            <h2>Koszt miesięczny według kategorii</h2>
            <CategoryBars items={summary.byCategory} />
            <p className="hint">
              Kwoty w innych walutach przeliczono na {summary.baseCurrency} według kursów z (symulowanego) API, buforowanych po stronie serwera.
            </p>
          </div>
        </>
      )}
    </section>
  )
}
