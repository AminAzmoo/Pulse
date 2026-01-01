export class ApiError extends Error {
  status?: number
  code?: string
  details?: unknown

  constructor(message: string, options?: { status?: number; code?: string; details?: unknown }) {
    super(message)
    this.name = 'ApiError'
    this.status = options?.status
    this.code = options?.code
    this.details = options?.details
  }
}

export class ApiConfigError extends ApiError {
  constructor(message = 'Backend API base URL is not configured') {
    super(message)
    this.name = 'ApiConfigError'
  }
}

export class ApiNetworkError extends ApiError {
  constructor(message = 'Network error while contacting backend', details?: unknown) {
    super(message, { details })
    this.name = 'ApiNetworkError'
  }
}

export class ApiValidationError extends ApiError {
  constructor(message = 'Invalid response envelope from backend', details?: unknown) {
    super(message, { details })
    this.name = 'ApiValidationError'
  }
}

export class ApiResponseError extends ApiError {
  constructor(message: string, options?: { status?: number; code?: string; details?: unknown }) {
    super(message, options)
    this.name = 'ApiResponseError'
  }
}
