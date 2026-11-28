// UC: Przeglądaj listę subskrypcji, Usuń / Wstrzymaj / Wznów (dodawanie i edycja – #17).
import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { ApiError } from '../api/httpClient'
import { SubscriptionTable } from '../components/SubscriptionTable'
import type { Category, Subscription } from '../types'

export function SubscriptionsPage() {
  const { api } = useAuth()
  const [items, setItems] = useState<Subscription[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [error, setError] = useState<string | null>(null)

  const reload = useCallback(async () => {
    try {
      const [subs, cats] = await Promise.all([api.subscriptions.list(), api.subscriptions.categories()])
      setItems(subs)
      setCategories(cats)
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Nie udało się pobrać danych')
    }
  }, [api])

  useEffect(() => {
    reload()
  }, [reload])

  async function run(action: () => Promise<unknown>) {
    setError(null)
    try {
      await action()
      await reload()
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Operacja nie powiodła się')
    }
  }

  return (
    <section>
      <header className="page-header">
        <h1>Subskrypcje</h1>
      </header>
      {error && <div className="error">{error}</div>}
      <SubscriptionTable
        items={items}
        categories={categories}
        onEdit={() => window.alert('Edycja będzie razem z formularzem (#17)')}
        onPause={(s) => run(() => api.subscriptions.pause(s.id))}
        onResume={(s, date) => run(() => api.subscriptions.resume(s.id, date))}
        onDelete={(s) => {
          if (window.confirm(`Usunąć „${s.name}” wraz z historią płatności?`)) {
            run(() => api.subscriptions.remove(s.id))
          }
        }}
      />
    </section>
  )
}
