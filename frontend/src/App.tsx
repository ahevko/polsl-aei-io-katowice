import { useState } from 'react'
import { useAuth } from './auth/AuthContext'
import { NavBar } from './components/NavBar'
import type { View } from './components/NavBar'
import { LoginPage } from './pages/LoginPage'
import { SubscriptionsPage } from './pages/SubscriptionsPage'

export function App() {
  const { user, ready } = useAuth()
  const [view, setView] = useState<View>('subscriptions')

  if (!ready) return <div className="loading">Ładowanie…</div>
  if (!user) return <LoginPage />

  return (
    <>
      <NavBar view={view} onNavigate={setView} />
      <main className="container">
        {view === 'subscriptions' && <SubscriptionsPage />}
      </main>
    </>
  )
}
