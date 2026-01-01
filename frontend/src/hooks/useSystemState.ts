import { useMemo } from 'react'
import { SYSTEM_STATE_COLORS, SYSTEM_STATE_LABELS } from '../constants'
import { useFreshness } from './useFreshness'

export type SystemState = 'NO_AGENTS' | 'STALE' | 'HEALTHY' | 'DEGRADED' | 'RECOVERING'

interface SystemStateInput {
  totalAgents?: number
  activeIncidents?: number
}

export const useSystemState = ({ totalAgents, activeIncidents }: SystemStateInput) => {
  const freshness = useFreshness()

  return useMemo(() => {
    let state: SystemState = 'HEALTHY'
    let stateReason = 'All systems operational'

    if ((totalAgents ?? 0) === 0) {
      state = 'NO_AGENTS'
      stateReason = 'No agents reporting yet'
    } else if (freshness.status === 'stale') {
      state = 'STALE'
      stateReason = 'Backend data is stale'
    } else if (freshness.consecutiveErrors > 0) {
      state = 'RECOVERING'
      stateReason = 'Recent backend errors detected'
    } else if ((activeIncidents ?? 0) > 0 || freshness.status === 'warning') {
      state = 'DEGRADED'
      stateReason = 'Incidents or warnings detected'
    }

    return {
      state,
      stateText: SYSTEM_STATE_LABELS[state],
      stateReason,
      statusColorClass: SYSTEM_STATE_COLORS[state],
      freshness,
    }
  }, [activeIncidents, freshness, totalAgents])
}
