import { QueryClient } from '@tanstack/react-query'
import { QUERY_STALE_TIME_MS } from '../constants'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: QUERY_STALE_TIME_MS,
    },
  },
})
