import { useQuery } from '@tanstack/react-query'
import CardShell from '../components/common/CardShell'
import StatusBadge from '../components/common/StatusBadge'
import NetworkGlobe from '../components/common/NetworkGlobe'
import { Server, Network, AlertTriangle, Bell, Activity } from 'lucide-react'
import { DEFAULT_STRINGS, POLLING_INTERVALS_MS } from '../constants'
import { useApi } from '../hooks/useApi'
import { useTimelineStore } from '../hooks/useTimelineStore'
import { useSystemState } from '../hooks/useSystemState'

export default function DashboardPage() {
  const api = useApi()
  const { activeIncidents, recentActivity } = useTimelineStore()
  const { data: nodes = [], isLoading: nodesLoading } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => api.getNodes(),
    refetchInterval: POLLING_INTERVALS_MS.nodes,
  })

  const { data: tunnels = [], isLoading: tunnelsLoading } = useQuery({
    queryKey: ['tunnels'],
    queryFn: () => api.getTunnels(),
    refetchInterval: POLLING_INTERVALS_MS.tunnels,
  })

  const onlineNodes = nodes.filter((n: any) => n.status === 'online').length
  const activeTunnels = tunnels.filter((t: any) => t.status === 'active').length
  const incidents = activeIncidents.length

  const globeNodes = nodes.map((node: any) => {
    const rawStatus = typeof node.status === 'string' ? node.status.toLowerCase() : ''
    const status =
      rawStatus === 'online'
        ? 'Online'
        : rawStatus === 'degraded'
          ? 'Degraded'
          : rawStatus === 'offline'
            ? 'Offline'
            : 'Unknown'
    return {
      id: node.id,
      name: node.name,
      lat: node.geo_data?.latitude || 0,
      lng: node.geo_data?.longitude || 0,
      status,
      role: node.role,
    }
  })

  const globeLinks = tunnels.map((tunnel: any) => {
    const rawStatus = typeof tunnel.status === 'string' ? tunnel.status.toLowerCase() : ''
    const status =
      rawStatus === 'active'
        ? 'Live'
        : rawStatus === 'configuring'
          ? 'Configuring'
          : rawStatus === 'error'
            ? 'Error'
            : 'Unknown'
    return {
      source: tunnel.source_node?.name || '',
      target: tunnel.dest_node?.name || '',
      status,
    }
  })

  const systemState = useSystemState({
    totalAgents: nodes.length,
    activeIncidents: activeIncidents.length,
  })
  const systemBadgeVariant =
    systemState.state === 'HEALTHY'
      ? 'neonA'
      : systemState.state === 'DEGRADED' || systemState.state === 'STALE'
        ? 'warn'
        : systemState.state === 'RECOVERING'
          ? 'default'
          : 'neonB'

  return (
    <div className="dashboard-container">
      {/* Fullscreen Globe Background */}
      <div className="dashboard-globe-wrapper">
        <NetworkGlobe nodes={globeNodes} links={globeLinks} />
      </div>

      {/* Floating Title - Top Center */}
      <div className="dashboard-title-wrapper">
        <h1 className="dashboard-title">
          Dashboard
        </h1>
        <p className="dashboard-subtitle">
          Monitor your entire Netly network at a glance
        </p>
      </div>

      {/* Floating Metrics - Left */}
      <div className="dashboard-metrics-wrapper">
        <CardShell className="dashboard-metric-card">
          <div className="dashboard-metric-icon-wrapper">
            <Server size={24} className="text-neon-a" />
          </div>
          <div className="dashboard-metric-value text-neon">{nodesLoading ? '...' : onlineNodes}</div>
          <div className="dashboard-metric-label">Online Nodes</div>
          <div className="dashboard-progress-track">
            <div className="dashboard-progress-bar dashboard-bar-online"></div>
          </div>
        </CardShell>

        <CardShell className="dashboard-metric-card">
          <div className="dashboard-metric-icon-wrapper">
            <Network size={24} className="text-neon-a" />
          </div>
          <div className="dashboard-metric-value text-neon">{tunnelsLoading ? '...' : activeTunnels}</div>
          <div className="dashboard-metric-label">Active Tunnels</div>
          <div className="dashboard-progress-track">
            <div className="dashboard-progress-bar dashboard-bar-tunnels"></div>
          </div>
        </CardShell>

        <CardShell className="dashboard-metric-card">
          <div className="dashboard-metric-icon-wrapper">
            <AlertTriangle size={24} className="text-neon-a" />
          </div>
          <div className="dashboard-metric-value text-neon">{incidents}</div>
          <div className="dashboard-metric-label">Current Incidents</div>
          <div className="dashboard-progress-track">
            <div className="dashboard-progress-bar dashboard-bar-incidents"></div>
          </div>
        </CardShell>
      </div>

      {/* Floating Recent Incidents - Right */}
      <div className="dashboard-incidents-wrapper">
        <CardShell className="dashboard-metric-card h-auto justify-start">
          <div className="flex items-center gap-3 mb-4">
            <div className="dashboard-metric-icon-wrapper mb-0">
              <Bell size={24} className="text-neon-a" />
            </div>
            <h3 className="text-white font-bold text-lg">Recent Incidents</h3>
          </div>

          <div className="dashboard-incidents-list w-full">
            {activeIncidents.length === 0 && (
              <div className="dashboard-incident-item">
                <p className="dashboard-incident-desc text-gray-400">{DEFAULT_STRINGS.notAvailable}</p>
              </div>
            )}
            {activeIncidents.map((event) => (
              <div key={event.id} className="dashboard-incident-item">
                <div className="dashboard-incident-header">
                  <StatusBadge
                    status={event.status || event.severity || 'Incident'}
                    variant={event.severity === 'ERROR' || event.status === 'failed' ? 'error' : event.status === 'pending' ? 'warn' : 'default'}
                  />
                  <span className="dashboard-incident-time">
                    {event.created_at ? new Date(event.created_at).toLocaleString() : DEFAULT_STRINGS.unknown}
                  </span>
                </div>
                <p className="dashboard-incident-desc">{event.message || DEFAULT_STRINGS.unknown}</p>
              </div>
            ))}
          </div>

          <div className="border-t border-white/5 mt-4 pt-4 w-full">
            <h4 className="text-sm font-semibold text-gray-200 mb-3">Recent Activity</h4>
            <div className="dashboard-incidents-list w-full">
              {recentActivity.length === 0 && (
                <div className="dashboard-incident-item">
                  <p className="dashboard-incident-desc text-gray-400">{DEFAULT_STRINGS.notAvailable}</p>
                </div>
              )}
              {recentActivity.map((event) => (
                <div key={`${event.id}-activity`} className="dashboard-incident-item">
                  <div className="dashboard-incident-header">
                    <StatusBadge
                      status={event.type || event.status || 'Event'}
                      variant={event.severity === 'ERROR' ? 'error' : event.severity === 'WARN' ? 'warn' : 'default'}
                      size="sm"
                    />
                    <span className="dashboard-incident-time">
                      {event.created_at ? new Date(event.created_at).toLocaleString() : DEFAULT_STRINGS.unknown}
                    </span>
                  </div>
                  <p className="dashboard-incident-desc">{event.message || DEFAULT_STRINGS.unknown}</p>
                </div>
              ))}
            </div>
          </div>
        </CardShell>
      </div>

      {/* Bottom - Status Bar */}
      <div className="dashboard-status-wrapper">
        <CardShell className="dashboard-status-card">
          <div className="dashboard-status-content">
            <div className="dashboard-status-row">
              <StatusBadge status={systemState.stateText} variant={systemBadgeVariant} />
              <span className="dashboard-label-sm">
                <Activity size={16} className="text-neon-a" />
                {systemState.stateReason}
              </span>
            </div>
            <div className="dashboard-label-xs">
              Last updated: {systemState.freshness.lastSuccessAtLabel}
            </div>
          </div>
        </CardShell>
      </div>
    </div>
  )
}
