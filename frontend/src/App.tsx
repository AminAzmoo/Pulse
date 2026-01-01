import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './lib/queryClient'
import DotGridBackground from './components/layout/DotGridBackground'
import HeaderDock from './components/layout/HeaderDock'
import BackendNotConfigured from './components/common/BackendNotConfigured'
import DashboardPage from './pages/DashboardPage'
import DevicesPage from './pages/DevicesPage'
import TunnelsPage from './pages/TunnelsPage'
import ServicesPage from './pages/ServicesPage'
import TimelinePage from './pages/TimelinePage'
import SettingsPage from './pages/SettingsPage'
import ProfilePage from './pages/ProfilePage'
import { isApiConfigured } from './lib/config'
import { FreshnessProvider } from './hooks/useFreshness'
import { TimelineProvider } from './hooks/useTimelineStore'

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <FreshnessProvider>
        <TimelineProvider>
          <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
            <DotGridBackground>
              <HeaderDock />
              <main className="app-main-content">
                {!isApiConfigured ? (
                  <BackendNotConfigured />
                ) : (
                  <Routes>
                    <Route path="/" element={<DashboardPage />} />
                    <Route path="/dashboard" element={<DashboardPage />} />
                    <Route path="/devices" element={<DevicesPage />} />
                    <Route path="/tunnels" element={<TunnelsPage />} />
                    <Route path="/services" element={<ServicesPage />} />
                    <Route path="/timeline" element={<TimelinePage />} />
                    <Route path="/settings" element={<SettingsPage />} />
                    <Route path="/profile" element={<ProfilePage />} />
                  </Routes>
                )}
              </main>
            </DotGridBackground>
          </BrowserRouter>
        </TimelineProvider>
      </FreshnessProvider>
    </QueryClientProvider>
  )
}

export default App
