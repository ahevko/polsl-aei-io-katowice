// AuthContext - stan sesji użytkownika (token + profil) dostępny w całej
// aplikacji; tworzy też klientów API (wstrzykiwanie zależności przez Context).
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { HttpClient } from '../api/httpClient'
import { AuthApi } from '../api/services'
import type { Currency, User } from '../types'

const TOKEN_KEY = 'subskrypcje.token'

function loadToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

function saveToken(token: string | null) {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    /* przeglądarka bez localStorage - sesja tylko w pamięci */
  }
}

export interface Apis {
  auth: AuthApi
}

interface AuthState {
  user: User | null
  ready: boolean
  api: Apis
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  setBaseCurrency: (c: Currency) => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(loadToken)
  const [user, setUser] = useState<User | null>(null)
  const [ready, setReady] = useState(false)

  const clearSession = useCallback(() => {
    saveToken(null)
    setToken(null)
    setUser(null)
  }, [])

  const api = useMemo<Apis>(() => {
    const http = new HttpClient('/api', () => token, clearSession)
    return {
      auth: new AuthApi(http),
    }
  }, [token, clearSession])

  // Odtworzenie sesji po odświeżeniu strony.
  useEffect(() => {
    if (!token) {
      setReady(true)
      return
    }
    api.auth
      .me()
      .then(setUser)
      .catch(clearSession)
      .finally(() => setReady(true))
  }, [token, api, clearSession])

  const value: AuthState = {
    user,
    ready,
    api,
    async login(email, password) {
      const res = await api.auth.login(email, password)
      saveToken(res.token)
      setToken(res.token)
      setUser(res.user)
    },
    async register(email, password) {
      await api.auth.register(email, password)
      await value.login(email, password)
    },
    async logout() {
      await api.auth.logout().catch(() => undefined)
      clearSession()
    },
    async setBaseCurrency(c) {
      setUser(await api.auth.setBaseCurrency(c))
    },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth musi być użyte wewnątrz AuthProvider')
  return ctx
}
