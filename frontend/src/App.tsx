import { useEffect, useState } from 'react'

// Pierwsza wersja – sprawdzamy tylko, czy frontend dogada się z backendem
// (Vite przekierowuje /api na http://localhost:8080, patrz vite.config.ts).
export function App() {
  const [status, setStatus] = useState('Łączenie z serwerem…')

  useEffect(() => {
    fetch('/api/ping')
      .then((r) => r.json())
      .then((d: { message: string }) => setStatus(`Serwer odpowiada: ${d.message}`))
      .catch(() => setStatus('Brak połączenia z serwerem – czy backend działa na :8080?'))
  }, [])

  return (
    <main className="container">
      <h1>Monitor subskrypcji</h1>
      <p>{status}</p>
    </main>
  )
}
