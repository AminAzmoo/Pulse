import { useMemo } from 'react'
import { createApiClient } from '../lib/apiClient'
import { runtimeConfig } from '../lib/config'
import { useFreshness } from './useFreshness'

export interface TimelineQueryParams {
  status?: string
  severity?: string
  type?: string
  entity_type?: string
  entity_id?: string
  from?: string
  to?: string
  resource_type?: string
  resource_id?: string
}

const buildQuery = (params?: TimelineQueryParams) => {
  if (!params) return ''
  const entries = Object.entries(params).filter(([, value]) => value)
  if (entries.length === 0) return ''
  return `?${new URLSearchParams(entries as Array<[string, string]>).toString()}`
}

export const useApi = () => {
  const { recordSuccess, recordError, setConnecting } = useFreshness()

  const client = useMemo(
    () =>
      createApiClient({
        baseUrl: runtimeConfig.apiBaseUrl,
        onSuccess: recordSuccess,
        onError: recordError,
        onConnecting: setConnecting,
      }),
    [recordError, recordSuccess, setConnecting]
  )

  return useMemo(
    () => ({
      getNodes: () => client.request<any[]>('/nodes'),
      createNode: (data: {
        name: string
        ip: string
        ssh_port?: number
        role: string
        username: string
        password?: string
        private_key?: string
      }) =>
        client.request('/nodes', {
          method: 'POST',
          body: JSON.stringify(data),
        }),
      deleteNode: (id: string) => client.request(`/nodes/${id}`, { method: 'DELETE' }),
      updateNode: (id: string, data: any) =>
        client.request(`/nodes/${id}`, {
          method: 'PUT',
          body: JSON.stringify(data),
        }),
      installAgent: (id: string) => client.request(`/nodes/${id}/install-agent`, { method: 'POST' }),
      getTaskStatus: (taskId: string) => client.request(`/tasks/${taskId}`),
      getInstallCommand: (nodeId: string) => client.request<{ command: string; api_url: string; token: string }>(`/nodes/${nodeId}/command`),
      getTunnels: () => client.request<any[]>('/tunnels'),
      createTunnel: (data: {
        name: string
        protocol: string
        source_node_id: number
        dest_node_id: number
        source_port?: number
        dest_port?: number
      }) =>
        client.request('/tunnels', {
          method: 'POST',
          body: JSON.stringify(data),
        }),
      deleteTunnel: (id: string) => client.request(`/tunnels/${id}`, { method: 'DELETE' }),
      getServices: () => client.request<any[]>('/services'),
      createService: (data: {
        name: string
        protocol: string
        node_id: number
        listen_port: number
        routing_mode: string
        config?: any
      }) =>
        client.request('/services', {
          method: 'POST',
          body: JSON.stringify(data),
        }),
      deleteService: (id: string) => client.request(`/services/${id}`, { method: 'DELETE' }),
      getTimeline: (params?: TimelineQueryParams) => client.request<any[]>(`/timeline${buildQuery(params)}`),
      cleanupNode: (data: { node_id: number; mode: 'soft' | 'hard'; force?: boolean; confirm_text?: string }) =>
        client.request('/cleanup', {
          method: 'POST',
          body: JSON.stringify(data),
        }),
      getSettings: () => client.request<Record<string, string>>('/settings'),
      updateSettings: (settings: Record<string, string>) =>
        client.request('/settings', {
          method: 'POST',
          body: JSON.stringify(settings),
        }),
      getGeneralSettings: () =>
        client.request<{
          systemName?: string
          adminEmail?: string
          publicUrl?: string
          environment?: string
        }>('/settings/general'),
      updateGeneralSettings: (data: {
        systemName: string
        adminEmail: string
        publicUrl: string
        environment: string
      }) =>
        client.request('/settings/general', {
          method: 'PUT',
          body: JSON.stringify(data),
        }),
      getNetworkStats: () =>
        client.request<{
          ipam?: {
            ipv4_cidr?: string
            ipv6_cidr?: string
            allocated_count?: number
            allocations?: Array<{
              ip?: string
              ipv6?: string
              type?: string
              resource_id?: number
              resource_name?: string
              allocated_at?: string
            }>
          }
          portam?: {
            min_port?: number
            max_port?: number
            total_range?: number
            used_count?: number
            available_count?: number
            allocations?: Array<{
              port?: number
              node_id?: number
              node_name?: string
              protocol?: string
              type?: string
              resource_id?: number
              resource_name?: string
            }>
          }
          fqdnam?: {
            base_domain?: string
            allocated_count?: number
            allocations?: Array<{
              fqdn?: string
              service_id?: number
              service_name?: string
              node_id?: number
              protocol?: string
              port?: number
              created_at?: string
            }>
          }
          summary?: {
            total_nodes?: number
            total_tunnels?: number
            total_services?: number
          }
        }>('/network/stats'),
      clearLogs: () => client.request('/settings/logs', { method: 'DELETE' }),
      getProfile: () => client.request<Record<string, string>>('/profile'),
    }),
    [client]
  )
}
