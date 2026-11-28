import { useAuth } from '../auth/AuthContext'

export type View = 'subscriptions' | 'summary'

const LINKS: { view: View; label: string }[] = [
  { view: 'subscriptions', label: 'Subskrypcje' },
  { view: 'summary', label: 'Podsumowanie' },
]

export function NavBar({ view, onNavigate }: { view: View; onNavigate: (v: View) => void }) {
  const { user, logout } = useAuth()
  return (
    <nav className="navbar">
      <span className="brand">💳 Subskrypcje</span>
      <div className="nav-links">
        {LINKS.map((l) => (
          <button key={l.view} className={view === l.view ? 'active' : undefined} onClick={() => onNavigate(l.view)}>
            {l.label}
          </button>
        ))}
      </div>
      <div className="nav-user">
        <span className="muted">{user?.email}</span>
        <button onClick={logout}>Wyloguj</button>
      </div>
    </nav>
  )
}
