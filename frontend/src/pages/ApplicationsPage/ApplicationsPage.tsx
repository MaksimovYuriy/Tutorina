import { useEffect, useState } from 'react'
import { Alert, Box, Button, CircularProgress, Container, MenuItem, Stack, TextField, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import {
  ApiError,
  getAdminApplications,
  getCurrentUser,
  getMyApplications,
  logout,
  updateAdminApplication,
  updateMyApplication,
  type ApplicationStatus,
  type LessonApplication,
} from '../../services/api'

const statusLabels: Record<ApplicationStatus, string> = {
  new: 'Новая',
  accepted: 'Принята',
  rejected: 'Отклонена',
  completed: 'Завершена',
}

export function ApplicationsPage() {
  const isAdmin = window.location.pathname.startsWith('/admin/')
  const [applications, setApplications] = useState<LessonApplication[]>([])
  const [selectedStatuses, setSelectedStatuses] = useState<Record<string, ApplicationStatus>>({})
  const [loading, setLoading] = useState(true)
  const [savingId, setSavingId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      getCurrentUser(controller.signal),
      isAdmin ? getAdminApplications(controller.signal) : getMyApplications(controller.signal),
    ]).then(([user, items]) => {
      const requiredRole = isAdmin ? 'admin' : 'teacher'
      if (!user.roles.includes(requiredRole)) {
        window.location.replace(user.roles.includes('admin') ? '/admin' : '/teacher')
        return
      }
      setApplications(items)
      setSelectedStatuses(Object.fromEntries(items.map((item) => [item.id, item.status])))
    }).catch((caught: unknown) => {
      if (caught instanceof DOMException && caught.name === 'AbortError') return
      if (caught instanceof ApiError && caught.status === 401) {
        window.location.replace('/login')
        return
      }
      setError('Не удалось загрузить заявки.')
    }).finally(() => setLoading(false))
    return () => controller.abort()
  }, [isAdmin])

  async function saveStatus(item: LessonApplication) {
    const status = selectedStatuses[item.id] ?? item.status
    setSavingId(item.id)
    setError(null)
    try {
      const updated = await (isAdmin ? updateAdminApplication(item.id, status) : updateMyApplication(item.id, status))
      setApplications((items) => items.map((entry) => entry.id === updated.id ? updated : entry))
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 409) {
        setError('Заявку нельзя принять: все места на занятие уже заняты.')
      } else {
        setError(caught instanceof ApiError ? caught.message : 'Не удалось изменить статус заявки.')
      }
    } finally {
      setSavingId(null)
    }
  }

  async function handleLogout() {
    await logout()
    window.location.assign('/login')
  }

  if (loading) return <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}><CircularProgress /></Box>

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 3, md: 6 } }}>
      <Container maxWidth="md">
        <Stack spacing={4}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
            <div>
              <BrandLink />
              <Typography component="h1" variant="h3">Заявки учеников</Typography>
            </div>
            <Stack direction="row" spacing={1}>
              <Button href={isAdmin ? '/admin' : '/teacher'} variant="outlined">Назад в кабинет</Button>
              <Button variant="outlined" onClick={handleLogout}>Выйти</Button>
            </Stack>
          </Stack>

          {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}
          {applications.length === 0 && <Alert severity="info">Новых заявок пока нет.</Alert>}

          <Stack spacing={2}>
            {applications.map((item) => (
              <Box key={item.id} component="article" sx={{ p: { xs: 3, md: 4 }, bgcolor: 'background.paper', borderRadius: 3, border: '1px solid', borderColor: 'divider' }}>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
                  <div>
                    <Typography variant="h6">{item.fullName}</Typography>
                    <Typography color="text.secondary">{item.offerTitle} · {formatLessonDate(item.lessonStartsAt)}</Typography>
                    {isAdmin && <Typography color="text.secondary">Преподаватель: {item.teacherDisplayName}</Typography>}
                  </div>
                  <Typography color="primary" sx={{ fontWeight: 650 }}>{statusLabels[item.status]}</Typography>
                </Stack>

                <Stack spacing={0.5} sx={{ mt: 2 }}>
                  <Typography>Телефон: <a href={`tel:${item.phone}`}>{item.phone}</a></Typography>
                  {item.email && <Typography>Email: <a href={`mailto:${item.email}`}>{item.email}</a></Typography>}
                  {item.comment && <Typography sx={{ mt: 1 }}>Комментарий: {item.comment}</Typography>}
                </Stack>

                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ mt: 3, alignItems: { sm: 'center' } }}>
                  <TextField select size="small" label="Статус" value={selectedStatuses[item.id] ?? item.status} onChange={(event) => setSelectedStatuses((current) => ({ ...current, [item.id]: event.target.value as ApplicationStatus }))} sx={{ minWidth: 180 }}>
                    <MenuItem value="new">Новая</MenuItem>
                    <MenuItem value="accepted">Принята</MenuItem>
                    <MenuItem value="rejected">Отклонена</MenuItem>
                    <MenuItem value="completed">Завершена</MenuItem>
                  </TextField>
                  <Button variant="contained" disabled={savingId === item.id || (selectedStatuses[item.id] ?? item.status) === item.status} onClick={() => saveStatus(item)}>
                    {savingId === item.id ? 'Сохраняем…' : 'Сохранить статус'}
                  </Button>
                </Stack>
              </Box>
            ))}
          </Stack>
        </Stack>
      </Container>
    </Box>
  )
}

function formatLessonDate(value: string): string {
  return new Intl.DateTimeFormat('ru-RU', {
    timeZone: 'Europe/Moscow', day: 'numeric', month: 'long', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  }).format(new Date(value))
}
