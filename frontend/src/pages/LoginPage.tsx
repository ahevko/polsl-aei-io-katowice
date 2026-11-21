// UC: Zaloguj się, Zarejestruj konto (aktor: Gość).
import { useState } from 'react'
import type { FormEvent } from 'react'
import { useAuth } from '../auth/AuthContext'
import { ApiError } from '../api/httpClient'

export function LoginPage() {
  const { login, register } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      if (mode === 'login') await login(email, password)
      else await register(email, password)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Nieznany błąd')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-screen">
      <form className="card auth-card" onSubmit={submit}>
        <h1>Monitor subskrypcji</h1>
        <p className="muted">
          {mode === 'login' ? 'Zaloguj się, aby zobaczyć swoje subskrypcje.' : 'Utwórz nowe konto.'}
        </p>
        <label>
          E-mail
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required autoFocus />
        </label>
        <label>
          Hasło
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={mode === 'register' ? 8 : undefined} />
        </label>
        {error && <div className="error">{error}</div>}
        <button className="primary" disabled={busy}>
          {mode === 'login' ? 'Zaloguj się' : 'Zarejestruj się'}
        </button>
        <button type="button" className="link" onClick={() => setMode(mode === 'login' ? 'register' : 'login')}>
          {mode === 'login' ? 'Nie masz konta? Zarejestruj się' : 'Masz już konto? Zaloguj się'}
        </button>
        {mode === 'login' && <p className="hint">Konto demonstracyjne: demo@example.com / demo1234</p>}
      </form>
    </div>
  )
}
