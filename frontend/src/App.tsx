import { useAuth } from './auth/AuthContext'
import { LoginPage } from './pages/LoginPage'

export function App() {
  const { user, ready, logout } = useAuth()

  if (!ready) return <div className="loading">Ładowanie…</div>
  if (!user) return <LoginPage />

  // TODO: lista subskrypcji (#14)
  return (
    <main className="container">
      <div className="card">
        <h1>Cześć, {user.email}!</h1>
        <p className="muted">Logowanie działa, lista subskrypcji w drodze.</p>
        <button onClick={logout}>Wyloguj</button>
      </div>
    </main>
  )
}
