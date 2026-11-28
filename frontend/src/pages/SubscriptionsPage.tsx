// UC: Przeglądaj listę subskrypcji, Dodaj / Edytuj / Usuń / Wstrzymaj / Wznów.
import { useCallback, useEffect, useState } from 'react'
import { useAuth } from '../auth/AuthContext'
import { ApiError } from '../api/httpClient'
import { SubscriptionForm } from '../components/SubscriptionForm'
import { SubscriptionTable } from '../components/SubscriptionTable'
import type { Category, Subscription, SubscriptionInput } from '../types'

type FormState = { mode: 'closed' } | { mode: 'create' } | { mode: 'edit'; sub: Subscription }

export function SubscriptionsPage() {
  const { api } = useAuth()
  const [items, setItems] = useState<Subscription[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [form, setForm] = useState<FormState>({ mode: 'closed' })
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

  async function save(input: SubscriptionInput) {
    if (form.mode === 'edit') await api.subscriptions.update(form.sub.id, input)
    else await api.subscriptions.create(input)
    setForm({ mode: 'closed' })
    await reload()
  }

  return (
    <section>
      <header className="page-header">
        <h1>Subskrypcje</h1>
        {form.mode === 'closed' && (
          <button className="primary" onClick={() => setForm({ mode: 'create' })}>
            + Dodaj subskrypcję
          </button>
        )}
      </header>
      {error && <div className="error">{error}</div>}
      {form.mode !== 'closed' && (
        <SubscriptionForm
          key={form.mode === 'edit' ? form.sub.id : 'new'}
          categories={categories}
          initial={form.mode === 'edit' ? form.sub : undefined}
          onSubmit={save}
          onCancel={() => setForm({ mode: 'closed' })}
        />
      )}
      <SubscriptionTable
        items={items}
        categories={categories}
        onEdit={(s) => setForm({ mode: 'edit', sub: s })}
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
