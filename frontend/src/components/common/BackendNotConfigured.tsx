import CardShell from './CardShell'
import { runtimeConfig } from '../../lib/config'

export default function BackendNotConfigured() {
  return (
    <div className="backend-empty-state">
      <CardShell className="backend-empty-card">
        <h2 className="text-xl font-semibold text-white mb-2">Backend not configured</h2>
        <p className="text-sm text-gray-300 mb-4">
          Set <span className="text-neon-a font-semibold">VITE_API_BASE_URL</span> to point to your backend API.
        </p>
        <div className="text-xs text-gray-500">
          Current value: {runtimeConfig.apiBaseUrl || 'Not set'}
        </div>
      </CardShell>
    </div>
  )
}
