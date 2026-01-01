import { ApiConfigError, ApiNetworkError, ApiResponseError, ApiValidationError } from './apiErrors'
import { logger } from './logger'

export interface ApiEnvelope<T> {
  data: T
  meta?: Record<string, unknown>
  error?: { message?: string; code?: string } | string
}

interface ApiClientOptions {
  baseUrl: string
  onSuccess?: () => void
  onError?: () => void
  onConnecting?: () => void
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const parseEnvelope = <T>(payload: unknown): ApiEnvelope<T> => {
  if (!isRecord(payload)) {
    throw new ApiValidationError('Response is not an object', payload)
  }
  if (!('data' in payload)) {
    throw new ApiValidationError('Missing data field in response', payload)
  }
  return payload as ApiEnvelope<T>
}

export const createApiClient = ({ baseUrl, onSuccess, onError, onConnecting }: ApiClientOptions) => {
  const sanitizedBaseUrl = baseUrl.replace(/\/$/, '')

  const request = async <T>(endpoint: string, options?: RequestInit) => {
    if (!sanitizedBaseUrl) {
      throw new ApiConfigError()
    }

    onConnecting?.()

    let response: Response
    try {
      response = await fetch(`${sanitizedBaseUrl}${endpoint}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...(options?.headers || {}),
        },
      })
    } catch (error) {
      onError?.()
      logger.error('API network error', { endpoint, error })
      throw new ApiNetworkError(undefined, error)
    }

    const payload = await response.json().catch((error) => {
      logger.warn('Failed to parse API response', { endpoint, error })
      return null
    })

    if (!response.ok) {
      onError?.()
      const message =
        (isRecord(payload) && typeof payload.error === 'string' && payload.error) ||
        response.statusText ||
        `HTTP ${response.status}`
      throw new ApiResponseError(message, { status: response.status, details: payload })
    }

    const envelope = parseEnvelope<T>(payload)

    if (envelope.error) {
      onError?.()
      const message =
        typeof envelope.error === 'string'
          ? envelope.error
          : envelope.error.message || 'Backend returned an error'
      throw new ApiResponseError(message, { details: envelope.error })
    }

    onSuccess?.()
    return envelope.data
  }

  return { request }
}
