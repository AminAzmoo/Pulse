export type ProcessStepState = 'pending' | 'running' | 'done' | 'error'

export interface ProcessStep {
  id: string
  label: string
  icon?: string
  state: ProcessStepState
}

export type DeviceRole = 'Entry' | 'Exit' | 'Hybrid' | 'Internal' | 'Unknown'
export type DeviceStatus = 'Pending' | 'Installing' | 'Online' | 'Degraded' | 'Offline' | 'Unknown'

export interface Device {
  id: string
  name: string
  role: DeviceRole
  ip: string
  location: string
  status: DeviceStatus
  cpu: number | string
  ram: number | string
  lastAction?: string
  lastActionTime?: string
  processSteps?: ProcessStep[]
  flagCode?: string
}

export type TunnelType = 'Single-hop' | 'Multi-hop' | 'Unknown'
export type TunnelStatus = 'Queued' | 'Planning' | 'Configuring' | 'Live' | 'Error' | 'Unknown'

export interface Tunnel {
  id: string
  name: string
  path: string
  type: TunnelType
  status: TunnelStatus
  latency: number | string
  lastAction?: string
  lastActionTime?: string
  processSteps?: ProcessStep[]
}

export type ServiceProtocol = 'HTTP' | 'HTTPS' | 'TCP' | 'UDP' | 'Unknown'
export type ServiceStatus = 'Queued' | 'Configuring' | 'Ready' | 'Error' | 'Unknown'

export interface Service {
  id: string
  name: string
  protocol: ServiceProtocol
  entryNode: string
  exitNode: string
  users: number | string
  traffic: string
  status: ServiceStatus
  lastAction?: string
  lastActionTime?: string
  processSteps?: ProcessStep[]
}

export type EventSeverity = 'INFO' | 'WARN' | 'ERROR'

export interface TimelineEvent {
  id: string
  time: string
  title: string
  description: string
  severity: EventSeverity
  entityType?: string
  entityId?: string
}
