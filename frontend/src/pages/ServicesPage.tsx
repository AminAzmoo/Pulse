import { useState, useCallback, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import PageShell from '../components/layout/PageShell'
import ServiceCard from '../components/entities/ServiceCard'
import ServicesTable from '../components/entities/ServicesTable'
import ViewToggle from '../components/common/ViewToggle'
import AddServiceModal from '../components/services/AddServiceModal'
import { Service, ProcessStep } from '../types'
import { DEFAULT_STRINGS, POLLING_INTERVALS_MS, PROCESS_TIMING_MS } from '../constants'
import { useApi } from '../hooks/useApi'

const EDIT_STEPS: ProcessStep[] = [
  { id: '1', label: 'Queued', state: 'pending', icon: 'ListStart' },
  { id: '2', label: 'Allocating IP/Port', state: 'pending', icon: 'Network' },
  { id: '3', label: 'Generating config', state: 'pending', icon: 'FileCode' },
  { id: '4', label: 'Pushing to nodes', state: 'pending', icon: 'Upload' },
  { id: '5', label: 'Testing connectivity', state: 'pending', icon: 'Wifi' },
  { id: '6', label: 'Ready', state: 'pending', icon: 'CheckCircle' },
]

const DELETE_STEPS: ProcessStep[] = [
  { id: '1', label: 'Queued', state: 'pending', icon: 'ListStart' },
  { id: '2', label: 'Draining connections', state: 'pending', icon: 'Activity' },
  { id: '3', label: 'Stopping service', state: 'pending', icon: 'Zap' },
  { id: '4', label: 'Releasing resources', state: 'pending', icon: 'ServerCog' },
  { id: '5', label: 'Done', state: 'pending', icon: 'CheckCircle' },
]

type ProcessType = 'edit' | 'delete'

interface ActiveProcess {
  type: ProcessType
  stepIndex: number
}

export default function ServicesPage() {
  const queryClient = useQueryClient()
  const api = useApi()
  const [view, setView] = useState<'card' | 'table'>('card')
  const [processes, setProcesses] = useState<Record<string, ActiveProcess>>({})
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editDraft, setEditDraft] = useState<Service | null>(null)
  const [isAddServiceOpen, setIsAddServiceOpen] = useState(false)
  const intervalsRef = useRef<Record<string, number>>({})

  const { data: rawServices = [], refetch: refetchServices } = useQuery({
    queryKey: ['services'],
    queryFn: () => api.getServices(),
    refetchInterval: POLLING_INTERVALS_MS.services,
  })

  const { data: rawNodes = [] } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => api.getNodes(),
    refetchInterval: POLLING_INTERVALS_MS.nodes,
  })

  const { data: rawTunnels = [] } = useQuery({
    queryKey: ['tunnels'],
    queryFn: () => api.getTunnels(),
    refetchInterval: POLLING_INTERVALS_MS.tunnels,
  })

  const nodes = rawNodes
    .map((node: any) => {
      const role = typeof node.role === 'string' ? node.role.toLowerCase() : ''
      if (role !== 'entry' && role !== 'exit') return null
      return {
        id: String(node.id),
        name: node.name || DEFAULT_STRINGS.unknown,
        role,
        ip: node.ip,
      }
    })
    .filter(Boolean) as Array<{ id: string; name: string; role: 'entry' | 'exit'; ip: string }>

  const tunnels = rawTunnels.map((t: any) => ({
    id: String(t.id),
    name: t.name || DEFAULT_STRINGS.unknown,
  }))

  const services: Service[] = rawServices.map((s: any) => {
    const rawStatus = typeof s.status === 'string' ? s.status.toLowerCase() : ''
    const status =
      rawStatus === 'ready'
        ? 'Ready'
        : rawStatus === 'error'
          ? 'Error'
          : rawStatus === 'configuring'
            ? 'Configuring'
            : rawStatus === 'queued'
              ? 'Queued'
              : 'Unknown'

    return {
      id: String(s.id),
      name: s.name || DEFAULT_STRINGS.unknown,
      protocol: s.protocol ? s.protocol.toUpperCase() : 'Unknown',
      entryNode: s.node?.name || DEFAULT_STRINGS.unknown,
      exitNode: s.exit_node?.name || DEFAULT_STRINGS.notAvailable,
      users: s.users ?? DEFAULT_STRINGS.notAvailable,
      traffic: s.traffic ?? DEFAULT_STRINGS.notAvailable,
      status,
      lastAction: s.last_action || DEFAULT_STRINGS.notAvailable,
      lastActionTime: s.updated_at ? new Date(s.updated_at).toLocaleString() : DEFAULT_STRINGS.unknown,
    }
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteService(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['services'] })
    },
  })

  useEffect(() => {
    return () => {
      Object.values(intervalsRef.current).forEach(clearInterval)
    }
  }, [])

  const getStepsWithState = (template: ProcessStep[], currentIndex: number): ProcessStep[] => {
    return template.map((step, index) => {
      let state: 'pending' | 'running' | 'done' = 'pending'
      if (index < currentIndex) state = 'done'
      else if (index === currentIndex) state = 'running'
      return { ...step, state }
    })
  }

  const runProcess = useCallback((serviceId: string, type: ProcessType, stepsTemplate: ProcessStep[], onComplete?: () => void) => {
    if (intervalsRef.current[serviceId]) {
      clearInterval(intervalsRef.current[serviceId])
      delete intervalsRef.current[serviceId]
    }

    setProcesses(prev => ({ ...prev, [serviceId]: { type, stepIndex: 0 } }))

    let currentStep = 0
    const totalSteps = stepsTemplate.length

    const interval = setInterval(() => {
      currentStep++
      
      setProcesses(prev => {
        if (!prev[serviceId]) return prev
        return { ...prev, [serviceId]: { type, stepIndex: currentStep } }
      })

      if (currentStep >= totalSteps - 1) {
        setTimeout(() => {
          clearInterval(interval)
          delete intervalsRef.current[serviceId]
          onComplete?.()
          setProcesses(prev => {
            const next = { ...prev }
            delete next[serviceId]
            return next
          })
        }, PROCESS_TIMING_MS.completionDelay)
      }
    }, PROCESS_TIMING_MS.stepInterval)

    intervalsRef.current[serviceId] = interval
  }, [])

  const handleEdit = (service: Service) => {
    if (processes[service.id]) return
    setEditingId(service.id)
    setEditDraft({ ...service })
  }

  const handleCancelEdit = () => {
    if (editingId && intervalsRef.current[editingId]) {
      clearInterval(intervalsRef.current[editingId])
      delete intervalsRef.current[editingId]
      setProcesses(prev => {
        const next = { ...prev }
        delete next[editingId]
        return next
      })
    }
    setEditingId(null)
    setEditDraft(null)
  }

  const handleSaveEdit = (serviceId: string) => {
    if (!editDraft) return
    runProcess(serviceId, 'edit', EDIT_STEPS, () => {
      setEditingId(null)
      setEditDraft(null)
    })
  }

  const handleDelete = (serviceId: string) => {
    if (processes[serviceId]) return
    if (window.confirm('Are you sure you want to delete this service?')) {
      runProcess(serviceId, 'delete', DELETE_STEPS, () => {
        deleteMutation.mutate(serviceId)
      })
    }
  }

  return (
    <>
    <PageShell
      title="Services"
      subtitle="Manage network services and configurations"
      headerRight={
        <button className="btn-primary-glow" onClick={() => setIsAddServiceOpen(true)}>
          Create Service
        </button>
      }
    >
      {/* Filters */}
      <div className="filter-bar">
        <div className="filter-group">
          <select className="filter-select">
            <option>All Status</option>
            <option>Ready</option>
            <option>Configuring</option>
            <option>Error</option>
          </select>
          <select className="filter-select">
            <option>All Protocols</option>
            <option>HTTP</option>
            <option>HTTPS</option>
            <option>TCP</option>
            <option>UDP</option>
          </select>
          <input
            type="text"
            placeholder="Search services"
            className="filter-input"
          />
        </div>
        <ViewToggle value={view} onChange={setView} />
      </div>

      {/* Content */}
      {view === 'card' ? (
        <div className="content-list">
          {services.map((service) => {
            const process = processes[service.id]
            const isEditing = editingId === service.id
            const steps = process 
              ? getStepsWithState(process.type === 'edit' ? EDIT_STEPS : DELETE_STEPS, process.stepIndex) 
              : undefined

            return (
              <ServiceCard 
                key={service.id} 
                service={isEditing && editDraft ? editDraft : service}
                onEdit={() => handleEdit(service)}
                onDelete={() => handleDelete(service.id)}
                isEditing={isEditing}
                isProcessing={!!process}
                processSteps={steps}
                processType={process?.type}
                onSave={() => handleSaveEdit(service.id)}
                onCancel={handleCancelEdit}
                onDraftChange={setEditDraft}
              />
            )
          })}
        </div>
      ) : (
        <ServicesTable 
            services={services}
            processes={processes}
            editingId={editingId}
            editDraft={editDraft}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onSave={handleSaveEdit}
            onCancel={handleCancelEdit}
        />
      )}
    </PageShell>
    
    <AddServiceModal
      isOpen={isAddServiceOpen}
      onClose={() => setIsAddServiceOpen(false)}
      onSuccess={refetchServices}
      nodes={nodes}
    />
    </>
  )
}
