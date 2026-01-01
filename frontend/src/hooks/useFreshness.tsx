import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import { FRESHNESS_THRESHOLDS_MS } from '../constants'

type FreshnessStatus = 'fresh' | 'warning' | 'stale' | 'connecting'

interface FreshnessState {
  lastSuccessAt: number | null
  consecutiveErrors: number
  connecting: boolean
}

interface FreshnessContextValue extends FreshnessState {
  status: FreshnessStatus
  lastSuccessAtLabel: string
  recordSuccess: () => void
  recordError: () => void
  setConnecting: () => void
}

const FreshnessContext = createContext<FreshnessContextValue | null>(null)

const formatRelativeTime = (timestamp: number | null) => {
  if (!timestamp) return 'Never'
  const diffSeconds = Math.floor((Date.now() - timestamp) / 1000)
  if (diffSeconds < 60) return `${diffSeconds}s ago`
  const diffMinutes = Math.floor(diffSeconds / 60)
  if (diffMinutes < 60) return `${diffMinutes}m ago`
  const diffHours = Math.floor(diffMinutes / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  return `${diffDays}d ago`
}

export const FreshnessProvider = ({ children }: { children: ReactNode }) => {
  const [lastSuccessAt, setLastSuccessAt] = useState<number | null>(null)
  const [consecutiveErrors, setConsecutiveErrors] = useState(0)
  const [connecting, setConnectingState] = useState(false)

  const recordSuccess = () => {
    setLastSuccessAt(Date.now())
    setConsecutiveErrors(0)
    setConnectingState(false)
  }

  const recordError = () => {
    setConsecutiveErrors((prev) => prev + 1)
    setConnectingState(false)
  }

  const setConnecting = () => setConnectingState(true)

  const status = useMemo<FreshnessStatus>(() => {
    if (connecting) return 'connecting'
    if (!lastSuccessAt) return 'stale'
    const age = Date.now() - lastSuccessAt
    if (age >= FRESHNESS_THRESHOLDS_MS.stale) return 'stale'
    if (age >= FRESHNESS_THRESHOLDS_MS.warning) return 'warning'
    return 'fresh'
  }, [connecting, lastSuccessAt])

  const value = useMemo(
    () => ({
      lastSuccessAt,
      consecutiveErrors,
      connecting,
      status,
      lastSuccessAtLabel: formatRelativeTime(lastSuccessAt),
      recordSuccess,
      recordError,
      setConnecting,
    }),
    [connecting, consecutiveErrors, lastSuccessAt, status]
  )

  return <FreshnessContext.Provider value={value}>{children}</FreshnessContext.Provider>
}

export const useFreshness = () => {
  const context = useContext(FreshnessContext)
  if (!context) {
    throw new Error('useFreshness must be used within a FreshnessProvider')
  }
  return context
}
