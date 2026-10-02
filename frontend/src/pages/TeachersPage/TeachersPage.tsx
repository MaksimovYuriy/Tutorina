import { useEffect, useState } from 'react'
import { Alert, Avatar, Box, Button, CircularProgress, Container, Stack, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import { getCurrentUser, getPublicTeacherProfiles, type CurrentUser, type TeacherProfile } from '../../services/api'

export function TeachersPage() {
  const [teachers, setTeachers] = useState<TeacherProfile[]>([])
  const [currentUser, setCurrentUser] = useState<CurrentUser | null | undefined>(undefined)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    const controller = new AbortController()
    getCurrentUser(controller.signal)
      .then(setCurrentUser)
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === 'AbortError') return
        setCurrentUser(null)
      })
    getPublicTeacherProfiles(controller.signal)
      .then(setTeachers)
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === 'AbortError') return
        setFailed(true)
      })
      .finally(() => setLoading(false))
    return () => controller.abort()
  }, [])

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 3, md: 6 } }}>
      <Container maxWidth="lg">
        <Stack spacing={{ xs: 4, md: 6 }}>
          <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
            <BrandLink />
            {currentUser === undefined ? (
              <Button variant="outlined" disabled>Проверяем вход…</Button>
            ) : currentUser ? (
              <Button href={currentUser.roles.includes('admin') ? '/admin' : '/teacher'} variant="outlined">
                {currentUser.roles.includes('admin') ? 'В админку' : 'В кабинет'}
              </Button>
            ) : (
              <Button href="/login" variant="outlined">Войти</Button>
            )}
          </Stack>

          <Box sx={{ maxWidth: 720 }}>
            <Typography component="h1" variant="h2">Преподаватели, с которыми легко учиться</Typography>
            <Typography color="text.secondary" sx={{ mt: 2, fontSize: 18 }}>
              Познакомьтесь с подходом и опытом преподавателей Tutorina.
            </Typography>
          </Box>

          {loading && <Box sx={{ display: 'grid', placeItems: 'center', py: 8 }}><CircularProgress /></Box>}
          {failed && <Alert severity="error">Не удалось загрузить преподавателей. Попробуйте обновить страницу.</Alert>}
          {!loading && !failed && teachers.length === 0 && <Alert severity="info">Скоро здесь появятся преподаватели.</Alert>}

          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, minmax(0, 1fr))' }, gap: 3 }}>
            {teachers.map((teacher) => (
              <Box key={teacher.id} component="article" sx={{ p: { xs: 3, md: 4 }, bgcolor: 'background.paper', borderRadius: 4, border: '1px solid', borderColor: 'divider' }}>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={3}>
                  <Avatar src={teacher.photoUrl || undefined} alt={teacher.displayName} sx={{ width: 112, height: 112, bgcolor: 'primary.light', color: 'primary.dark', fontSize: 36 }}>
                    {teacher.displayName.slice(0, 1).toUpperCase()}
                  </Avatar>
                  <Stack spacing={1.5}>
                    <div>
                      <Typography component="h2" variant="h5">{teacher.displayName}</Typography>
                      {teacher.education && <Typography color="text.secondary">{teacher.education}</Typography>}
                    </div>
                    {teacher.experience && <Typography><strong>Опыт:</strong> {teacher.experience}</Typography>}
                    {teacher.approach && <Typography>{teacher.approach}</Typography>}
                  </Stack>
                </Stack>
              </Box>
            ))}
          </Box>
        </Stack>
      </Container>
    </Box>
  )
}
