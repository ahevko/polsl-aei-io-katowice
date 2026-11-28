// Prosty wykres słupkowy (CSS) – w prototypie #2 zastąpiony biblioteką wykresów.
import { formatMoney } from '../types'
import type { CategoryCost } from '../types'

export function CategoryBars({ items }: { items: CategoryCost[] }) {
  if (items.length === 0) return <p className="muted">Brak aktywnych subskrypcji.</p>
  const max = Math.max(...items.map((i) => i.monthly.amount))
  const total = items.reduce((s, i) => s + i.monthly.amount, 0)
  return (
    <ul className="bars">
      {items.map((i) => (
        <li key={i.categoryId}>
          <span className="bar-label">{i.categoryName}</span>
          <span className="bar-track">
            <span className="bar-fill" style={{ width: `${(i.monthly.amount / max) * 100}%` }} />
          </span>
          <span className="bar-value">
            {formatMoney(i.monthly)} <small className="muted">({Math.round((i.monthly.amount / total) * 100)}%)</small>
          </span>
        </li>
      ))}
    </ul>
  )
}
