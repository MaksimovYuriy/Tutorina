import { useEffect, useState } from 'react'
import { Alert, Box, CircularProgress, Container, Stack, Typography } from '@mui/material'
import { getApiStatus, type ApiStatus } from './services/api'

export function App() {
  const [status, setStatus] = useState<ApiStatus | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    getApiStatus(controller.signal)
      .then(setStatus)
      .catch((caughtError: unknown) => {
        if (caughtError instanceof DOMException && caughtError.name === 'AbortError') return
        setError(caughtError instanceof Error ? caughtError.message : 'Не удалось связаться с API')
      })
    return () => controller.abort()
  }, [])

  return (
    <Box component="main" sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center', py: 6 }}>
      <Container maxWidth="sm">
        <Stack spacing={3}>
          <div>
            <Typography component="h1" variant="h3" gutterBottom>
              Tutorina
            </Typography>
            <Typography color="text.secondary">
              Стартовый каркас приложения готов. Здесь можно начинать собирать первый экран.
            </Typography>
          </div>

          {!status && !error && (
            <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
              <CircularProgress size={20} />
              <Typography>Проверяем frontend → Caddy → backend → PostgreSQL…</Typography>
            </Stack>
          )}
          {status && (
            <Alert severity="success">
              API отвечает, PostgreSQL: {status.database}. Время сервера: {status.time}
            </Alert>
          )}
          {error && <Alert severity="warning">{error}</Alert>}
        </Stack>
      </Container>
    </Box>
  )
}
