import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import PageShell from '../components/layout/PageShell'
import CardShell from '../components/common/CardShell'
import StatusBadge from '../components/common/StatusBadge'
import { DEFAULT_STRINGS, TIMELINE_SEVERITIES, TIMELINE_STATUSES } from '../constants'
import { useTimelineStore } from '../hooks/useTimelineStore'

const getEntityRoute = (resourceType?: string, entityType?: string, id?: string | number) => {
  const type = (resourceType || entityType || '').toLowerCase()
  const identifier = id ? `?entityId=${id}` : ''
  if (type.includes('device') || type.includes('node')) return `/devices${identifier}`
  if (type.includes('tunnel')) return `/tunnels${identifier}`
  if (type.includes('service')) return `/services${identifier}`
  return null
}

export default function TimelinePage() {
  const { filteredEvents, events, loading, error, filters, setFilters } = useTimelineStore()
  const resourceTypes = useMemo(
    () => Array.from(new Set(events.map((event) => event.resource_type).filter(Boolean))),
    [events]
  )

  return (
    <PageShell title="Timeline" subtitle="View all system events and activity logs">
      <div className="timeline-grid">
        {/* Left - Filters */}
        <div className="timeline-filters-col">
          <CardShell>
            <h3 className="timeline-filter-title">Filters</h3>
            <div className="timeline-filter-group">
              <div>
                <label className="timeline-filter-label">Severity</label>
                <select
                  className="timeline-filter-select"
                  value={filters.severity || ''}
                  onChange={(event) => setFilters({ severity: event.target.value || undefined })}
                >
                  <option value="">All</option>
                  {TIMELINE_SEVERITIES.map((severity) => (
                    <option key={severity} value={severity}>
                      {severity}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="timeline-filter-label">Status</label>
                <select
                  className="timeline-filter-select"
                  value={filters.status || ''}
                  onChange={(event) => setFilters({ status: event.target.value || undefined })}
                >
                  <option value="">All</option>
                  {TIMELINE_STATUSES.map((status) => (
                    <option key={status} value={status}>
                      {status}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="timeline-filter-label">Resource Type</label>
                <select
                  className="timeline-filter-select"
                  value={filters.resource_type || ''}
                  onChange={(event) => setFilters({ resource_type: event.target.value || undefined })}
                >
                  <option value="">All</option>
                  {resourceTypes.map((resourceType) => (
                    <option key={resourceType} value={resourceType}>
                      {resourceType}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="timeline-filter-label">Date Range</label>
                <div className="timeline-date-range">
                  <input
                    type="datetime-local"
                    className="timeline-filter-select"
                    value={filters.from || ''}
                    onChange={(event) => setFilters({ from: event.target.value || undefined })}
                  />
                  <input
                    type="datetime-local"
                    className="timeline-filter-select"
                    value={filters.to || ''}
                    onChange={(event) => setFilters({ to: event.target.value || undefined })}
                  />
                </div>
              </div>
              <div>
                <label className="timeline-filter-label">Search</label>
                <input
                  type="text"
                  className="timeline-filter-select"
                  placeholder="Search message, ID, entity, correlation"
                  value={filters.search || ''}
                  onChange={(event) => setFilters({ search: event.target.value })}
                />
              </div>
            </div>
          </CardShell>
        </div>

        {/* Right - Events */}
        <div className="timeline-events-col">
          <div className="timeline-events-list">
            {loading ? (
              <CardShell><p className="text-white">Loading events...</p></CardShell>
            ) : error ? (
              <CardShell><p className="text-red-400 text-sm">{error}</p></CardShell>
            ) : filteredEvents.length === 0 ? (
              <CardShell><p className="text-white">No events found</p></CardShell>
            ) : filteredEvents.map((event) => {
              const entityRoute = getEntityRoute(event.resource_type, event.entity_type, event.entity_id || event.resource_id)
              return (
              <CardShell key={event.id} hover>
                <div className="timeline-event-content">
                  <div className="timeline-event-main">
                    <div className="timeline-event-header">
                      <StatusBadge
                        status={event.status || 'Unknown'}
                        variant={event.status === 'failed' ? 'error' : event.status === 'pending' ? 'warn' : 'default'}
                      />
                      <StatusBadge
                        status={event.severity || 'INFO'}
                        variant={event.severity === 'ERROR' ? 'error' : event.severity === 'WARN' ? 'warn' : 'default'}
                        size="sm"
                      />
                      <span className="timeline-event-time">
                        {event.created_at ? new Date(event.created_at).toLocaleString() : DEFAULT_STRINGS.unknown}
                      </span>
                    </div>
                    <h3 className="timeline-event-title">{event.type || 'Event'}</h3>
                    <p className="timeline-event-desc">{event.message || DEFAULT_STRINGS.unknown}</p>
                    {(event.resource_type || event.entity_type) && (
                      <div className="timeline-event-meta">
                        Related:{' '}
                        {entityRoute ? (
                          <Link className="text-neon-a hover:underline" to={entityRoute}>
                            {(event.resource_type || event.entity_type) ?? 'Resource'} #{event.resource_id || event.entity_id}
                          </Link>
                        ) : (
                          <span>
                            {(event.resource_type || event.entity_type) ?? 'Resource'} #{event.resource_id || event.entity_id}
                          </span>
                        )}
                        {event.correlation_id && (
                          <span className="ml-2 text-gray-500">Correlation: {event.correlation_id}</span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </CardShell>
            )})}
          </div>
        </div>
      </div>
    </PageShell>
  )
}
