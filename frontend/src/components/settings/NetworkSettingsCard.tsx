import { useQuery } from '@tanstack/react-query'
import CardShell from '../common/CardShell'
import { DEFAULT_STRINGS } from '../../constants'
import { useApi } from '../../hooks/useApi'

const parseList = (value?: string) => {
  if (!value) return []
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}

export default function NetworkSettingsCard() {
  const api = useApi()
  const { data: settings, isLoading, error } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.getSettings(),
  })

  const reservedBlocks = parseList(settings?.reserved_blocks)
  const servicePools = parseList(settings?.service_pools)

  return (
    <CardShell className="card-overflow-hidden">
      <div className="network-settings-content">
        <div>
          <h3 className="settings-card-title">IPAM / PortAM</h3>
          <p className="settings-card-subtitle">Network configuration values from the backend.</p>
        </div>

        {isLoading && <p className="text-gray-400">Loading network settings...</p>}
        {error && <p className="text-red-400 text-sm">Network settings unavailable.</p>}

        {!isLoading && !error && (
          <div className="settings-grid-2col">
            <div className="settings-column-space">
              <h4 className="settings-section-header settings-section-header-blue">Address Pools</h4>
              <div className="settings-readonly-row">
                <span className="settings-label">IPv4 Pool Range</span>
                <span className="text-white">{settings?.ipam_ipv4_pool || DEFAULT_STRINGS.unknown}</span>
              </div>
              <div className="settings-readonly-row">
                <span className="settings-label">IPv6 Pool Range</span>
                <span className="text-white">{settings?.ipam_ipv6_pool || DEFAULT_STRINGS.unknown}</span>
              </div>
              <div className="settings-readonly-row">
                <span className="settings-label">Reserved Blocks</span>
                <span className="text-white">
                  {reservedBlocks.length > 0 ? reservedBlocks.join(', ') : DEFAULT_STRINGS.notAvailable}
                </span>
              </div>
            </div>

            <div className="settings-column-space">
              <h4 className="settings-section-header settings-section-header-purple">Port Management</h4>
              <div className="settings-readonly-row">
                <span className="settings-label">Global Port Range</span>
                <span className="text-white">{settings?.portam_range || DEFAULT_STRINGS.unknown}</span>
              </div>
              <div className="settings-readonly-row">
                <span className="settings-label">Service Pools</span>
                <span className="text-white">
                  {servicePools.length > 0 ? servicePools.join(', ') : DEFAULT_STRINGS.notAvailable}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>
    </CardShell>
  )
}
