// Lista subskrypcji z akcjami: edytuj, wstrzymaj / wznów, usuń.
import { useState } from 'react'
import { CYCLE_LABELS, formatDate, formatMoney, todayISO } from '../types'
import type { Category, Subscription } from '../types'

interface Props {
  items: Subscription[]
  categories: Category[]
  onEdit: (s: Subscription) => void
  onPause: (s: Subscription) => void
  onResume: (s: Subscription, nextPaymentDate: string) => void
  onDelete: (s: Subscription) => void
}

export function SubscriptionTable({ items, categories, onEdit, onPause, onResume, onDelete }: Props) {
  const [resumingId, setResumingId] = useState<string | null>(null)
  const [resumeDate, setResumeDate] = useState(todayISO(1))
  const catName = (id: string) => categories.find((c) => c.id === id)?.name ?? id

  if (items.length === 0) {
    return <div className="card empty">Nie masz jeszcze żadnych subskrypcji.</div>
  }

  return (
    <div className="card table-wrap">
      <table>
        <thead>
          <tr>
            <th>Nazwa</th>
            <th>Kategorie</th>
            <th className="num">Koszt</th>
            <th>Cykl</th>
            <th>Następna płatność</th>
            <th>Status</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {items.map((s) => (
            <tr key={s.id} className={s.status === 'PAUSED' ? 'paused' : undefined}>
              <td data-label="Nazwa">
                <strong>{s.name}</strong>
              </td>
              <td data-label="Kategorie">
                {s.categoryIds.map((id) => (
                  <span key={id} className="tag">
                    {catName(id)}
                  </span>
                ))}
              </td>
              <td data-label="Koszt" className="num">
                {formatMoney(s.price)}
              </td>
              <td data-label="Cykl">{CYCLE_LABELS[s.cycle]}</td>
              <td data-label="Następna płatność">{s.status === 'ACTIVE' ? formatDate(s.nextPaymentDate) : '—'}</td>
              <td data-label="Status">
                <span className={s.status === 'ACTIVE' ? 'badge ok' : 'badge warn'}>
                  {s.status === 'ACTIVE' ? 'aktywna' : 'wstrzymana'}
                </span>
              </td>
              <td className="row-actions">
                {resumingId === s.id ? (
                  <span className="resume">
                    <input type="date" value={resumeDate} onChange={(e) => setResumeDate(e.target.value)} />
                    <button
                      className="primary"
                      onClick={() => {
                        onResume(s, resumeDate)
                        setResumingId(null)
                      }}
                    >
                      Wznów
                    </button>
                    <button onClick={() => setResumingId(null)}>×</button>
                  </span>
                ) : (
                  <>
                    <button onClick={() => onEdit(s)}>Edytuj</button>
                    {s.status === 'ACTIVE' ? (
                      <button onClick={() => onPause(s)}>Wstrzymaj</button>
                    ) : (
                      <button
                        onClick={() => {
                          setResumeDate(todayISO(1))
                          setResumingId(s.id)
                        }}
                      >
                        Wznów…
                      </button>
                    )}
                    <button className="danger" onClick={() => onDelete(s)}>
                      Usuń
                    </button>
                  </>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
