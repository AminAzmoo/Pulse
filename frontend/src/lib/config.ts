export const runtimeConfig = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '',
}

export const isApiConfigured = Boolean(runtimeConfig.apiBaseUrl)
