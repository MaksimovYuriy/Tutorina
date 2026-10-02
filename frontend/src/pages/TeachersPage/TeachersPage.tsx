import { useEffect, useState } from 'react'
import { Alert, Avatar, Box, Button, Chip, CircularProgress, Container, Stack, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import { getCurrentUser, getPublicOffers, getPublicTeacherProfiles, type CurrentUser, type Offer, type TeacherProfile } from '../../services/api'

const formatLabels = {
  online: 'Онлайн',
  offline: 'Очно',
  both: 'Онлайн и очно',
}

export function TeachersPage() {
  const [teachers, setTeachers] = useState<TeacherProfile[]>([])
  const [offers, setOffers] = useState<Offer[]>([])
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
    Promise.all([getPublicOffers(controller.signal), getPublicTeacherProfiles(controller.signal)])
      .then(([offerItems, teacherItems]) => {
        setOffers(offerItems)
        setTeachers(teacherItems)
      })
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
              Выберите направление и познакомьтесь с преподавателями Tutorina.
            </Typography>
          </Box>

          {loading && <Box sx={{ display: 'grid', placeItems: 'center', py: 8 }}><CircularProgress /></Box>}
          {failed && <Alert severity="error">Не удалось загрузить направления и преподавателей. Попробуйте обновить страницу.</Alert>}

          {!loading && !failed && (
            <Stack component="section" spacing={3}>
              <div>
                <Typography component="h2" variant="h3">Направления занятий</Typography>
                <Typography color="text.secondary" sx={{ mt: 1 }}>Опубликованные программы и преподаватели, которые их ведут.</Typography>
              </div>
              {offers.length === 0 && <Alert severity="info">Скоро здесь появятся направления занятий.</Alert>}
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, minmax(0, 1fr))' }, gap: 3 }}>
                {offers.map((offer) => (
                  <Box key={offer.id} component="article" sx={{ p: { xs: 3, md: 4 }, bgcolor: 'background.paper', borderRadius: 4, border: '1px solid', borderColor: 'divider' }}>
                    <Stack spacing={2.5}>
                      <div>
                        <Typography component="h3" variant="h5">{offer.title}</Typography>
                        <Stack direction="row" spacing={1} sx={{ mt: 1.5, flexWrap: 'wrap', gap: 1 }}>
                          <Chip size="small" label={formatLabels[offer.format]} />
                          <Chip size="small" label={`${offer.defaultDurationMinutes} мин.`} />
                          {offer.priceRubles !== null && <Chip size="small" label={`${offer.priceRubles.toLocaleString('ru-RU')} ₽`} />}
                        </Stack>
                      </div>
                      {offer.description && <Typography>{offer.description}</Typography>}
                      {offer.goal && <Typography color="text.secondary"><strong>Цель:</strong> {offer.goal}</Typography>}
                      <Box>
                        <Typography sx={{ fontWeight: 650, mb: 1 }}>Преподаватели</Typography>
                        {offer.teachers.length === 0 ? (
                          <Typography variant="body2" color="text.secondary">Преподаватель будет назначен позже.</Typography>
                        ) : (
                          <Stack spacing={1}>
                            {offer.teachers.map((assignment) => {
                              const duration = assignment.durationMinutes ?? offer.defaultDurationMinutes
                              const price = assignment.priceRubles ?? offer.priceRubles
                              return (
                                <Stack key={assignment.id} direction="row" spacing={2} sx={{ justifyContent: 'space-between', alignItems: 'baseline' }}>
                                  <Typography>{assignment.teacherDisplayName}</Typography>
                                  <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
                                    {duration} мин.{price !== null ? ` · ${price.toLocaleString('ru-RU')} ₽` : ''}
                                  </Typography>
                                </Stack>
                              )
                            })}
                          </Stack>
                        )}
                      </Box>
                    </Stack>
                  </Box>
                ))}
              </Box>
            </Stack>
          )}

          {!loading && !failed && (
            <Stack component="section" spacing={3}>
              <Typography component="h2" variant="h3">Наши преподаватели</Typography>
              {teachers.length === 0 && <Alert severity="info">Скоро здесь появятся преподаватели.</Alert>}
              <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'repeat(2, minmax(0, 1fr))' }, gap: 3 }}>
                {teachers.map((teacher) => (
                  <Box key={teacher.id} component="article" sx={{ p: { xs: 3, md: 4 }, bgcolor: 'background.paper', borderRadius: 4, border: '1px solid', borderColor: 'divider' }}>
                    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={3}>
                      <Avatar src={teacher.photoUrl || undefined} alt={teacher.displayName} sx={{ width: 112, height: 112, bgcolor: 'primary.light', color: 'primary.dark', fontSize: 36 }}>
                        {teacher.displayName.slice(0, 1).toUpperCase()}
                      </Avatar>
                      <Stack spacing={1.5}>
                        <div>
                          <Typography component="h3" variant="h5">{teacher.displayName}</Typography>
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
          )}
        </Stack>
      </Container>
    </Box>
  )
}
