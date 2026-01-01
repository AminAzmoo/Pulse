export const DEFAULT_STRINGS = {
  unknown: 'Unknown / N/A',
  notAvailable: 'Not available',
}

export const POLLING_INTERVALS_MS = {
  timeline: 15000,
  nodes: 15000,
  tunnels: 15000,
  services: 15000,
  networkStats: 30000,
  taskStatus: 2000,
  profile: 60000,
}

export const FRESHNESS_THRESHOLDS_MS = {
  warning: 30000,
  stale: 60000,
}

export const TIMELINE_LIMITS = {
  recentActivity: 6,
  activeIncidents: 5,
}

export const QUERY_STALE_TIME_MS = 30000

export const PROCESS_TIMING_MS = {
  stepInterval: 1500,
  completionDelay: 1000,
  installTimeoutPolls: 120,
  settingsStepDelay: 800,
}

export const SYSTEM_STATE_LABELS = {
  NO_AGENTS: 'No agents connected',
  STALE: 'Data stale',
  HEALTHY: 'Healthy',
  DEGRADED: 'Degraded',
  RECOVERING: 'Recovering',
}

export const SYSTEM_STATE_COLORS = {
  NO_AGENTS: 'text-gray-400',
  STALE: 'text-yellow-400',
  HEALTHY: 'text-emerald-400',
  DEGRADED: 'text-red-400',
  RECOVERING: 'text-blue-400',
}

export const TIMELINE_SEVERITIES = ['INFO', 'WARN', 'ERROR'] as const
export const TIMELINE_STATUSES = ['open', 'resolved', 'failed', 'pending', 'success'] as const

export const INCIDENT_SEVERITIES = new Set(['ERROR'])
export const INCIDENT_STATUSES = new Set(['failed', 'open'])
