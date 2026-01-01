import { Link, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  LayoutDashboard,
  Server,
  Network,
  Activity,
  Clock,
  Settings,
  Moon,
  User,
} from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import { useTimelineStore } from '../../hooks/useTimelineStore'
import { useSystemState } from '../../hooks/useSystemState'
import { POLLING_INTERVALS_MS } from '../../constants'
import { isApiConfigured } from '../../lib/config'

const tabs = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Devices', path: '/devices', icon: Server },
  { name: 'Tunnels', path: '/tunnels', icon: Network },
  { name: 'Services', path: '/services', icon: Activity },
  { name: 'Timeline', path: '/timeline', icon: Clock },
  { name: 'Settings', path: '/settings', icon: Settings },
  { name: 'Profile', path: '/profile', icon: User },
]

export default function HeaderDock() {
  const location = useLocation()
  const api = useApi()
  const { activeIncidents } = useTimelineStore()
  const { data: nodes = [] } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => api.getNodes(),
    refetchInterval: POLLING_INTERVALS_MS.nodes,
    enabled: isApiConfigured,
  })
  const systemState = useSystemState({
    totalAgents: nodes.length,
    activeIncidents: activeIncidents.length,
  })

  return (
    <header className="header-dock">
      <div className="header-content">
        {/* Logo */}
        <div className="logo-container">
          <div className="logo-icon">
            <span className="logo-symbol">N</span>
          </div>
          <span className="logo-text">Netly</span>
        </div>

        {/* Tabs */}
        <nav className="nav-container">
          {tabs.map((tab) => {
            const isActive =
              location.pathname === tab.path ||
              (tab.path === '/' && location.pathname === '/dashboard')
            const Icon = tab.icon
            return (
              <Link
                key={tab.path}
                to={tab.path}
                className={`nav-link ${
                  isActive
                    ? 'nav-link-active'
                    : 'nav-link-inactive'
                }`}
              >
                <Icon size={18} className={`${isActive ? "text-neon-a" : "text-muted"} icon-mr-2`} />
                {tab.name}
              </Link>
            )
          })}
        </nav>

        {/* Right side */}
        <div className="header-right-actions">
          <div className="header-status">
            <div className={`header-status-state ${systemState.statusColorClass}`}>
              {systemState.stateText}
            </div>
            <div className="header-status-meta">
              <span className="text-gray-400">{systemState.stateReason}</span>
              <span className="text-gray-500">· Updated {systemState.freshness.lastSuccessAtLabel}</span>
            </div>
          </div>
          <button className="header-icon-btn">
            <Moon size={20} />
          </button>
          <div className="user-avatar">
            <User size={20} className="text-muted" />
          </div>
        </div>
      </div>
    </header>
  )
}
