import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { INCIDENT_SEVERITIES, INCIDENT_STATUSES, POLLING_INTERVALS_MS, TIMELINE_LIMITS } from '../constants'
import { isApiConfigured } from '../lib/config'
import { useApi, TimelineQueryParams } from './useApi'

export interface TimelineEventItem {
  id: string
  status?: string
  severity?: string
  message?: string
  type?: string
  created_at?: string
  resource_type?: string
  resource_id?: string | number
  entity_type?: string
  entity_id?: string | number
  correlation_id?: string
}

export interface TimelineFilters extends TimelineQueryParams {
  search?: string
}

interface TimelineStore {
  events: TimelineEventItem[]
  filteredEvents: TimelineEventItem[]
  activeIncidents: TimelineEventItem[]
  recentActivity: TimelineEventItem[]
  loading: boolean
  error: string | null
  filters: TimelineFilters
  setFilters: (filters: Partial<TimelineFilters>) => void
  refresh: () => Promise<void>
}

const TimelineContext = createContext<TimelineStore | null>(null)

const normalizeEvent = (event: any): TimelineEventItem => ({
  id: String(event.id ?? event.uuid ?? `${event.type || 'event'}-${event.created_at || Date.now()}`),
  status: event.status,
  severity: event.severity,
  message: event.message || event.description,
  type: event.type,
  created_at: event.created_at || event.timestamp,
  resource_type: event.resource_type,
  resource_id: event.resource_id,
  entity_type: event.entity_type,
  entity_id: event.entity_id,
  correlation_id: event.correlation_id,
})

const matchesSearch = (event: TimelineEventItem, search: string) => {
  if (!search) return true
  const needle = search.toLowerCase()
  return [
    event.message,
    event.id,
    event.entity_id?.toString(),
    event.correlation_id,
    event.resource_id?.toString(),
  ]
    .filter(Boolean)
    .some((value) => value?.toString().toLowerCase().includes(needle))
}

const filterEvents = (events: TimelineEventItem[], filters: TimelineFilters) => {
  return events.filter((event) => {
    if (filters.status && event.status !== filters.status) return false
    if (filters.severity && event.severity !== filters.severity) return false
    if (filters.type && event.type !== filters.type) return false
    if (filters.entity_type && event.entity_type !== filters.entity_type) return false
    if (filters.entity_id && String(event.entity_id) !== filters.entity_id) return false
    if (filters.resource_type && event.resource_type !== filters.resource_type) return false
    if (filters.resource_id && String(event.resource_id) !== filters.resource_id) return false
    if (filters.from && event.created_at && new Date(event.created_at) < new Date(filters.from)) return false
    if (filters.to && event.created_at && new Date(event.created_at) > new Date(filters.to)) return false
    if (!matchesSearch(event, filters.search || '')) return false
    return true
  })
}

const sortByTimestamp = (events: TimelineEventItem[]) =>
  [...events].sort((a, b) => {
    const aTime = a.created_at ? new Date(a.created_at).getTime() : 0
    const bTime = b.created_at ? new Date(b.created_at).getTime() : 0
    return bTime - aTime
  })

export const TimelineProvider = ({ children }: { children: ReactNode }) => {
  const api = useApi()
  const [events, setEvents] = useState<TimelineEventItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [filters, setFiltersState] = useState<TimelineFilters>({})

  const refresh = useCallback(async () => {
    if (!isApiConfigured) {
      setEvents([])
      setError('Backend not configured')
      setLoading(false)
      return
    }

    setLoading(true)
    setError(null)
    try {
      const data = await api.getTimeline()
      setEvents(data.map(normalizeEvent))
    } catch (err: any) {
      setError(err?.message || 'Failed to load timeline')
    } finally {
      setLoading(false)
    }
  }, [api])

  useEffect(() => {
    if (!isApiConfigured) {
      refresh()
      return
    }

    refresh()
    const interval = window.setInterval(refresh, POLLING_INTERVALS_MS.timeline)
    return () => window.clearInterval(interval)
  }, [refresh])

  const setFilters = useCallback((next: Partial<TimelineFilters>) => {
    setFiltersState((prev) => ({ ...prev, ...next }))
  }, [])

  const filteredEvents = useMemo(() => filterEvents(sortByTimestamp(events), filters), [events, filters])

  const activeIncidents = useMemo(
    () =>
      sortByTimestamp(events).filter((event) => {
        if (event.severity && INCIDENT_SEVERITIES.has(event.severity)) return true
        if (event.status && INCIDENT_STATUSES.has(event.status)) return true
        return false
      }),
    [events]
  )

  const recentActivity = useMemo(
    () => sortByTimestamp(events).slice(0, TIMELINE_LIMITS.recentActivity),
    [events]
  )

  const value = useMemo(
    () => ({
      events,
      filteredEvents,
      activeIncidents: activeIncidents.slice(0, TIMELINE_LIMITS.activeIncidents),
      recentActivity,
      loading,
      error,
      filters,
      setFilters,
      refresh,
    }),
    [events, filteredEvents, activeIncidents, recentActivity, loading, error, filters, setFilters, refresh]
  )

  return <TimelineContext.Provider value={value}>{children}</TimelineContext.Provider>
}

export const useTimelineStore = () => {
  const context = useContext(TimelineContext)
  if (!context) {
    throw new Error('useTimelineStore must be used within a TimelineProvider')
  }
  return context
}
