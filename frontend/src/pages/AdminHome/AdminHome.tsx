import { useEffect, useState } from 'react'
import { Alert, Box, Button, CircularProgress, Container, Stack, Typography } from '@mui/material'
import { ApiError, getCurrentUser, logout, type CurrentUser } from '../../services/api'

const roleLabels = {
  teacher: 'Преподаватель',
  admin: 'Администратор',
} as const

export function AdminHome() {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    getCurrentUser(controller.signal)
      .then(setUser)
      .catch((caughtError: unknown) => {
        if (caughtError instanceof DOMException && caughtError.name === 'AbortError') return
        if (caughtError instanceof ApiError && caughtError.status === 401) {
          window.location.replace('/login')
          return
        }
        setError('Не удалось загрузить кабинет.')
      })
    return () => controller.abort()
  }, [])

  async function handleLogout() {
    await logout()
    window.location.assign('/login')
  }

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 4, md: 8 } }}>
      <Container maxWidth="md">
        {!user && !error && <CircularProgress aria-label="Загрузка кабинета" />}
        {error && <Alert severity="error">{error}</Alert>}
        {user && (
          <Stack spacing={4}>
            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
              <div>
                <Typography color="primary" sx={{ fontWeight: 700, mb: 1 }}>Tutorina</Typography>
                <Typography component="h1" variant="h3">Личный кабинет</Typography>
              </div>
              <Button variant="outlined" onClick={handleLogout}>Выйти</Button>
            </Stack>
            <Box sx={{ p: { xs: 3, md: 5 }, bgcolor: 'background.paper', borderRadius: 4 }}>
              <Typography variant="h5" gutterBottom>{user.email}</Typography>
              <Typography color="text.secondary">
                Роли: {user.roles.map((role) => roleLabels[role]).join(', ')}
              </Typography>
              <Alert severity="info" sx={{ mt: 3 }}>
                Основа кабинета готова. Следующие разделы появятся после согласования ТЗ.
              </Alert>
            </Box>
          </Stack>
        )}
      </Container>
    </Box>
  )
}

