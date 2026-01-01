export const useApiStatus = (endpoint: string) => {
  const config = useRuntimeConfig();
  const url = endpoint.startsWith('http')
    ? endpoint
    : `${config.public.apiBase}${endpoint}`;

  const { data, error, pending } = useFetch(url, { method: 'GET' });

  const statusLabel = computed(() => {
    if (pending.value) {
      return 'Checking';
    }
    if (error.value) {
      return 'Offline';
    }
    if (!data.value) {
      return 'Unknown';
    }
    return 'Online';
  });

  return {
    data,
    error,
    pending,
    statusLabel,
  };
};
