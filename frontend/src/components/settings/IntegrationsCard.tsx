import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import CardShell from '../common/CardShell'
import { useApi } from '../../hooks/useApi'

interface IntegrationField {
  key: string
  label: string
  type?: 'text' | 'password'
}

interface IntegrationDefinition {
  id: string
  name: string
  subtitle: string
  enabledKey?: string
  fields: IntegrationField[]
}

const INTEGRATIONS: IntegrationDefinition[] = [
  {
    id: 'prometheus',
    name: 'Prometheus Metrics',
    subtitle: 'Export metrics to Prometheus',
    enabledKey: 'prometheus_enabled',
    fields: [
      { key: 'prometheus_endpoint', label: 'Endpoint URL' },
      { key: 'prometheus_token', label: 'Auth Token', type: 'password' },
      { key: 'prometheus_path', label: 'Metrics Path' },
    ],
  },
  {
    id: 'cloudflare',
    name: 'Cloudflare Tunnel',
    subtitle: 'Expose your server via Cloudflare Tunnel',
    enabledKey: 'cloudflare_enabled',
    fields: [
      { key: 'cloudflare_email', label: 'Cloudflare Email' },
      { key: 'cloudflare_global_key', label: 'Global API Key', type: 'password' },
      { key: 'cloudflare_account_id', label: 'Account ID' },
      { key: 'cloudflare_tunnel_name', label: 'Tunnel Name' },
      { key: 'cloudflare_public_url', label: 'Public URL (optional)' },
    ],
  },
  {
    id: 'firebase',
    name: 'Firebase Auth',
    subtitle: 'External authentication provider',
    enabledKey: 'firebase_enabled',
    fields: [
      { key: 'firebase_project_id', label: 'Project ID' },
      { key: 'firebase_api_key', label: 'API Key', type: 'password' },
      { key: 'firebase_auth_domain', label: 'Auth Domain' },
    ],
  },
]

export default function IntegrationsCard() {
  const api = useApi()
  const queryClient = useQueryClient()
  const { data: settings } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.getSettings(),
  })

  const [drafts, setDrafts] = useState<Record<string, string>>({})

  useEffect(() => {
    if (settings) {
      setDrafts(settings)
    }
  }, [settings])

  const updateSettingsMutation = useMutation({
    mutationFn: (data: Record<string, string>) => api.updateSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })

  const setDraftValue = (key: string, value: string) => {
    setDrafts((prev) => ({ ...prev, [key]: value }))
  }

  const handleSave = (definition: IntegrationDefinition) => {
    const payload: Record<string, string> = {}
    definition.fields.forEach((field) => {
      if (drafts[field.key] !== undefined) {
        payload[field.key] = drafts[field.key]
      }
    })
    if (definition.enabledKey && drafts[definition.enabledKey] !== undefined) {
      payload[definition.enabledKey] = drafts[definition.enabledKey]
    }
    updateSettingsMutation.mutate(payload)
  }

  const integrationsState = useMemo(
    () =>
      INTEGRATIONS.map((definition) => {
        const enabledValue = definition.enabledKey ? drafts[definition.enabledKey] : undefined
        const isEnabled = enabledValue === 'true'
        const configuredFields = definition.fields.filter((field) => Boolean(drafts[field.key]))
        const status = isEnabled
          ? 'Active'
          : configuredFields.length > 0
            ? 'Configured'
            : 'Not configured'
        return { definition, isEnabled, configuredFieldsCount: configuredFields.length, status }
      }),
    [drafts]
  )

  return (
    <CardShell>
      <div className="p-2">
        <h3 className="settings-card-title">Integrations</h3>

        <div className="integration-card-list">
          {integrationsState.map(({ definition, isEnabled, status }) => (
            <div key={definition.id} className="integration-item">
              <div className="integration-header">
                <div className="flex-1">
                  <div className="flex-center gap-3">
                    <h4 className="integration-name">{definition.name}</h4>
                    {definition.enabledKey && (
                      <span className={`badge ${isEnabled ? 'badge-active' : 'badge-inactive'}`}>
                        {isEnabled ? 'ON' : 'OFF'}
                      </span>
                    )}
                  </div>
                  <p className="integration-subtitle">{definition.subtitle}</p>
                  <p className="integration-status">{status}</p>
                </div>
              </div>

              <div className="integration-drawer">
                <div className="drawer-grid">
                  {definition.fields.map((field) => (
                    <div key={field.key}>
                      <label className="settings-label">{field.label}</label>
                      <input
                        type={field.type || 'text'}
                        value={drafts[field.key] ?? ''}
                        onChange={(event) => setDraftValue(field.key, event.target.value)}
                        className="settings-input"
                        placeholder="Enter value"
                      />
                    </div>
                  ))}
                </div>

                <div className="flex-between border-top-separator">
                  <div className="flex-1">
                    {definition.enabledKey && (
                      <label className="flex items-center gap-2 text-sm text-gray-400">
                        <input
                          type="checkbox"
                          checked={isEnabled}
                          onChange={(event) =>
                            setDraftValue(definition.enabledKey as string, event.target.checked ? 'true' : 'false')
                          }
                        />
                        Enabled
                      </label>
                    )}
                  </div>
                  <button
                    onClick={() => handleSave(definition)}
                    disabled={updateSettingsMutation.isPending}
                    className="settings-btn-secondary"
                  >
                    {updateSettingsMutation.isPending ? 'Saving...' : 'Save'}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </CardShell>
  )
}
