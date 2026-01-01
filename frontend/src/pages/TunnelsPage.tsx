import { useState, useCallback, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import PageShell from '../components/layout/PageShell'
import TunnelCard from '../components/entities/TunnelCard'
import TunnelsTable from '../components/entities/TunnelsTable'
import ViewToggle from '../components/common/ViewToggle'
import AddTunnelModal from '../components/tunnels/AddTunnelModal'
import { Tunnel, ProcessStep } from '../types'
import { DEFAULT_STRINGS, POLLING_INTERVALS_MS, PROCESS_TIMING_MS } from '../constants'
import { useApi } from '../hooks/useApi'

const EDIT_STEPS: ProcessStep[] = [
  { id: '1', label: 'Queued', state: 'pending', icon: 'ListStart' },
  { id: '2', label: 'Validating config', state: 'pending', icon: 'FileCode' },
  { id: '3', label: 'Applying routes', state: 'pending', icon: 'Route' },
  { id: '4', label: 'Warm-up checks', state: 'pending', icon: 'Thermometer' },
  { id: '5', label: 'Live', state: 'pending', icon: 'Zap' },
]

const DELETE_STEPS: ProcessStep[] = [
  { id: '1', label: 'Queued', state: 'pending', icon: 'ListStart' },
  { id: '2', label: 'Draining tunnel', state: 'pending', icon: 'Activity' },
  { id: '3', label: 'Removing routes', state: 'pending', icon: 'Route' },
  { id: '4', label: 'Deregistering', state: 'pending', icon: 'ServerCog' },
  { id: '5', label: 'Done', state: 'pending', icon: 'CheckCircle' },
]

type ProcessType = 'edit' | 'delete'

interface ActiveProcess {
  type: ProcessType
  stepIndex: number
}

export default function TunnelsPage() {
  const queryClient = useQueryClient()
  const api = useApi()
  const [view, setView] = useState<'card' | 'table'>('card')
  const [processes, setProcesses] = useState<Record<string, ActiveProcess>>({})
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editDraft, setEditDraft] = useState<Tunnel | null>(null)
  const [isAddTunnelOpen, setIsAddTunnelOpen] = useState(false)
  const intervalsRef = useRef<Record<string, number>>({})

  const { data: rawTunnels = [], refetch: refetchTunnels } = useQuery({
    queryKey: ['tunnels'],
    queryFn: () => api.getTunnels(),
    refetchInterval: POLLING_INTERVALS_MS.tunnels,
  })

  const { data: rawNodes = [] } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => api.getNodes(),
    refetchInterval: POLLING_INTERVALS_MS.nodes,
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

  const tunnels: Tunnel[] = rawTunnels.map((t: any) => ({
    id: String(t.id),
    name: t.name || DEFAULT_STRINGS.unknown,
    path: `${t.source_node?.name || DEFAULT_STRINGS.unknown} → ${t.dest_node?.name || DEFAULT_STRINGS.unknown}`,
    type: t.type ? (t.type === 'chain' ? 'Multi-hop' : 'Single-hop') : 'Unknown',
    status: t.status
      ? t.status === 'active'
        ? 'Live'
        : t.status === 'error'
          ? 'Error'
          : 'Configuring'
      : 'Unknown',
    latency: t.latency ?? DEFAULT_STRINGS.notAvailable,
    lastAction: t.last_action || DEFAULT_STRINGS.notAvailable,
    lastActionTime: t.updated_at ? new Date(t.updated_at).toLocaleString() : DEFAULT_STRINGS.unknown,
  }))

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteTunnel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tunnels'] })
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

  const runProcess = useCallback((tunnelId: string, type: ProcessType, stepsTemplate: ProcessStep[], onComplete?: () => void) => {
    if (intervalsRef.current[tunnelId]) {
      clearInterval(intervalsRef.current[tunnelId])
      delete intervalsRef.current[tunnelId]
    }

    setProcesses(prev => ({ ...prev, [tunnelId]: { type, stepIndex: 0 } }))

    let currentStep = 0
    const totalSteps = stepsTemplate.length

    const interval = setInterval(() => {
      currentStep++
      
      setProcesses(prev => {
        if (!prev[tunnelId]) return prev
        return { ...prev, [tunnelId]: { type, stepIndex: currentStep } }
      })

      if (currentStep >= totalSteps - 1) {
        setTimeout(() => {
          clearInterval(interval)
          delete intervalsRef.current[tunnelId]
          onComplete?.()
          setProcesses(prev => {
            const next = { ...prev }
            delete next[tunnelId]
            return next
          })
        }, PROCESS_TIMING_MS.completionDelay)
      }
    }, PROCESS_TIMING_MS.stepInterval)

    intervalsRef.current[tunnelId] = interval
  }, [])

  const handleEdit = (tunnel: Tunnel) => {
    if (processes[tunnel.id]) return
    setEditingId(tunnel.id)
    setEditDraft({ ...tunnel })
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

  const handleSaveEdit = (tunnelId: string) => {
    if (!editDraft) return
    runProcess(tunnelId, 'edit', EDIT_STEPS, () => {
      setEditingId(null)
      setEditDraft(null)
    })
  }

  const handleDelete = (tunnelId: string) => {
    if (processes[tunnelId]) return
    if (window.confirm('Are you sure you want to delete this tunnel?')) {
      runProcess(tunnelId, 'delete', DELETE_STEPS, () => {
        deleteMutation.mutate(tunnelId)
      })
    }
  }

  return (
    <>
    <PageShell
      title="Tunnels"
      subtitle="Manage network tunnels and routing paths"
      headerRight={
        <button className="btn-primary-glow" onClick={() => setIsAddTunnelOpen(true)}>
          Create Tunnel
        </button>
      }
    >
      {/* Filters */}
      <div className="filter-bar">
        <div className="filter-group">
          <select className="filter-select">
            <option>All Status</option>
            <option>Live</option>
            <option>Configuring</option>
            <option>Error</option>
          </select>
          <select className="filter-select">
            <option>All Types</option>
            <option>Single-hop</option>
            <option>Multi-hop</option>
          </select>
          <input
            type="text"
            placeholder="Search tunnels"
            className="filter-input"
          />
        </div>
        <ViewToggle value={view} onChange={setView} />
      </div>

      {/* Content */}
      {view === 'card' ? (
        <div className="content-list">
          {tunnels.map((tunnel) => {
            const process = processes[tunnel.id]
            const isEditing = editingId === tunnel.id
            const steps = process 
              ? getStepsWithState(process.type === 'edit' ? EDIT_STEPS : DELETE_STEPS, process.stepIndex) 
              : undefined

            return (
              <TunnelCard 
                key={tunnel.id} 
                tunnel={isEditing && editDraft ? editDraft : tunnel}
                onEdit={() => handleEdit(tunnel)}
                onDelete={() => handleDelete(tunnel.id)}
                isEditing={isEditing}
                isProcessing={!!process}
                processSteps={steps}
                processType={process?.type}
                onSave={() => handleSaveEdit(tunnel.id)}
                onCancel={handleCancelEdit}
                onDraftChange={setEditDraft}
              />
            )
          })}
        </div>
      ) : (
        <TunnelsTable 
            tunnels={tunnels}
            processes={processes}
            editingId={editingId}
            editDraft={editDraft}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onSave={handleSaveEdit}
            onCancel={handleCancelEdit}
            onDraftChange={setEditDraft}
            getStepsWithState={getStepsWithState}
            EDIT_STEPS={EDIT_STEPS}
            DELETE_STEPS={DELETE_STEPS}
        />
      )}
    </PageShell>
    
    <AddTunnelModal
      isOpen={isAddTunnelOpen}
      onClose={() => setIsAddTunnelOpen(false)}
      onSuccess={refetchTunnels}
      nodes={nodes}
    />
    </>
  )
}
