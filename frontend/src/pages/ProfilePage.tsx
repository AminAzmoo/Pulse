import { useQuery } from '@tanstack/react-query'
import PageShell from '../components/layout/PageShell'
import CardShell from '../components/common/CardShell'
import { DEFAULT_STRINGS, POLLING_INTERVALS_MS } from '../constants'
import { useApi } from '../hooks/useApi'

export default function ProfilePage() {
  const api = useApi()
  const { data, isLoading, error } = useQuery({
    queryKey: ['profile'],
    queryFn: () => api.getProfile(),
    refetchInterval: POLLING_INTERVALS_MS.profile,
  })

  const profileEntries = data ? Object.entries(data) : []

  return (
    <PageShell title="Profile" subtitle="Account and identity details">
      <CardShell>
        {isLoading && <p className="text-gray-300">Loading profile data...</p>}
        {error && <p className="text-red-400 text-sm">Profile data unavailable.</p>}
        {!isLoading && !error && profileEntries.length === 0 && (
          <p className="text-gray-400">{DEFAULT_STRINGS.notAvailable}</p>
        )}
        {!isLoading && !error && profileEntries.length > 0 && (
          <div className="grid gap-3">
            {profileEntries.map(([key, value]) => (
              <div key={key} className="flex justify-between text-sm border-b border-white/5 pb-2">
                <span className="text-gray-400">{key}</span>
                <span className="text-white">{value || DEFAULT_STRINGS.unknown}</span>
              </div>
            ))}
          </div>
        )}
      </CardShell>
    </PageShell>
  )
}
