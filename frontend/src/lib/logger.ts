type LogLevel = 'debug' | 'info' | 'warn' | 'error'

const levelOrder: Record<LogLevel, number> = {
  debug: 10,
  info: 20,
  warn: 30,
  error: 40,
}

const envLevel = (import.meta.env.VITE_LOG_LEVEL as LogLevel | undefined) || 'info'

const shouldLog = (level: LogLevel) => levelOrder[level] >= levelOrder[envLevel]

const formatMessage = (level: LogLevel, message: string, meta?: Record<string, unknown>) => ({
  level,
  message,
  meta,
  timestamp: new Date().toISOString(),
})

export const logger = {
  debug: (message: string, meta?: Record<string, unknown>) => {
    if (shouldLog('debug')) {
      console.debug(formatMessage('debug', message, meta))
    }
  },
  info: (message: string, meta?: Record<string, unknown>) => {
    if (shouldLog('info')) {
      console.info(formatMessage('info', message, meta))
    }
  },
  warn: (message: string, meta?: Record<string, unknown>) => {
    if (shouldLog('warn')) {
      console.warn(formatMessage('warn', message, meta))
    }
  },
  error: (message: string, meta?: Record<string, unknown>) => {
    if (shouldLog('error')) {
      console.error(formatMessage('error', message, meta))
    }
  },
}
