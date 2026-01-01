import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import CardShell from '../common/CardShell'
import { useApi } from '../../hooks/useApi'
import { PROCESS_TIMING_MS } from '../../constants'

export default function SystemCard() {
  const [isClearing, setIsClearing] = useState(false)
  const [statusMessage, setStatusMessage] = useState<string | null>(null)
  const [statusTone, setStatusTone] = useState<'success' | 'error' | null>(null)
  const api = useApi()

  const clearLogsMutation = useMutation({
    mutationFn: () => api.clearLogs(),
    onSuccess: () => {
      setStatusMessage('Logs cleared successfully')
      setStatusTone('success')
    },
    onError: () => {
      setStatusMessage('Failed to clear logs')
      setStatusTone('error')
    },
  })

  const handleClearLogs = () => {
    if (confirm('Are you sure you want to clear all logs?')) {
      setIsClearing(true)
      setStatusMessage(null)
      setStatusTone(null)
      clearLogsMutation.mutate()
      setTimeout(() => setIsClearing(false), PROCESS_TIMING_MS.completionDelay)
    }
  }

  return (
    <CardShell className="system-card">
      <div className="flex-col-between-full">
        <div>
          <h3 className="settings-card-title">System</h3>
          <div className="settings-form-grid">
            <div>
              <label className="settings-label">Log Management</label>
              <p className="text-xs text-gray-500 mb-3">Clear all system logs and error files</p>
              <button 
                onClick={handleClearLogs}
                disabled={isClearing}
                className="px-4 py-2 bg-red-500/20 border border-red-500 text-red-500 rounded-lg hover:bg-red-500/30 disabled:opacity-50"
              >
                {isClearing ? 'Clearing...' : 'Clear Logs'}
              </button>
              {statusMessage && (
                <p className={`text-xs mt-2 ${statusTone === 'success' ? 'text-green-400' : 'text-red-400'}`}>
                  {statusMessage}
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </CardShell>
  )
}
