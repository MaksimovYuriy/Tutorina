import { useEffect, useState, type FormEvent } from 'react'
import { Alert, Box, Button, Checkbox, CircularProgress, Container, FormControlLabel, MenuItem, Stack, TextField, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import {
  ApiError,
  archiveOffer,
  assignTeacherToOffer,
  createOffer,
  getCurrentUser,
  getOffers,
  getTeacherProfiles,
  logout,
  removeTeacherFromOffer,
  updateOffer,
  updateTeacherOffer,
  type Offer,
  type OfferInput,
  type TeacherProfile,
} from '../../services/api'

const emptyOffer: OfferInput = {
  title: '',
  description: '',
  goal: '',
  defaultDurationMinutes: 60,
  format: 'online',
  priceRubles: null,
  isPublished: false,
}

const formatLabels = {
  online: 'Онлайн',
  offline: 'Очно',
  both: 'Онлайн и очно',
}

export function AdminOffersPage() {
  const [offers, setOffers] = useState<Offer[]>([])
  const [teachers, setTeachers] = useState<TeacherProfile[]>([])
  const [form, setForm] = useState<OfferInput>(emptyOffer)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [selectedTeachers, setSelectedTeachers] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      getCurrentUser(controller.signal),
      getOffers(controller.signal),
      getTeacherProfiles(controller.signal),
    ]).then(([user, offerItems, teacherItems]) => {
      if (!user.roles.includes('admin')) throw new ApiError('Недостаточно прав', 403)
      setOffers(offerItems)
      setTeachers(teacherItems)
    }).catch((caught: unknown) => {
      if (caught instanceof DOMException && caught.name === 'AbortError') return
      if (caught instanceof ApiError && caught.status === 401) {
        window.location.replace('/login')
        return
      }
      setError(caught instanceof ApiError && caught.status === 403 ? 'Для этого раздела нужна роль администратора.' : 'Не удалось загрузить направления.')
    }).finally(() => setLoading(false))
    return () => controller.abort()
  }, [])

  function resetForm() {
    setForm(emptyOffer)
    setEditingId(null)
  }

  function startEditing(offer: Offer) {
    setEditingId(offer.id)
    setForm({
      title: offer.title,
      description: offer.description,
      goal: offer.goal,
      defaultDurationMinutes: offer.defaultDurationMinutes,
      format: offer.format,
      priceRubles: offer.priceRubles,
      isPublished: offer.isPublished,
    })
    setError(null)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const saved = editingId ? await updateOffer(editingId, form) : await createOffer(form)
      setOffers((items) => {
        const next = items.some((item) => item.id === saved.id)
          ? items.map((item) => item.id === saved.id ? saved : item)
          : [...items, saved]
        return next.sort((left, right) => left.title.localeCompare(right.title, 'ru'))
      })
      resetForm()
    } catch (caught) {
      setError(caught instanceof ApiError ? caught.message : 'Не удалось сохранить направление.')
    } finally {
      setSaving(false)
    }
  }

  async function handleAssign(offer: Offer) {
    const teacherProfileId = selectedTeachers[offer.id]
    if (!teacherProfileId) return
    setError(null)
    try {
      const assignment = await assignTeacherToOffer(offer.id, {
        teacherProfileId: Number(teacherProfileId),
        durationMinutes: null,
        priceRubles: null,
        isPublished: true,
      })
      setOffers((items) => items.map((item) => item.id === offer.id
        ? { ...item, teachers: [...item.teachers, assignment].sort((a, b) => a.teacherDisplayName.localeCompare(b.teacherDisplayName, 'ru')) }
        : item))
      setSelectedTeachers((current) => ({ ...current, [offer.id]: '' }))
    } catch (caught) {
      setError(caught instanceof ApiError ? caught.message : 'Не удалось назначить преподавателя.')
    }
  }

  async function handleAssignmentVisibility(offer: Offer, assignmentId: string) {
    const assignment = offer.teachers.find((item) => item.id === assignmentId)
    if (!assignment) return
    setError(null)
    try {
      const updated = await updateTeacherOffer(assignment.id, {
        teacherProfileId: Number(assignment.teacherProfileId),
        durationMinutes: assignment.durationMinutes,
        priceRubles: assignment.priceRubles,
        isPublished: !assignment.isPublished,
      })
      setOffers((items) => items.map((item) => item.id === offer.id
        ? { ...item, teachers: item.teachers.map((entry) => entry.id === updated.id ? updated : entry) }
        : item))
    } catch {
      setError('Не удалось изменить видимость назначения.')
    }
  }

  async function handleRemoveAssignment(offer: Offer, assignmentId: string) {
    if (!window.confirm('Убрать преподавателя из этого направления?')) return
    setError(null)
    try {
      await removeTeacherFromOffer(assignmentId)
      setOffers((items) => items.map((item) => item.id === offer.id
        ? { ...item, teachers: item.teachers.filter((entry) => entry.id !== assignmentId) }
        : item))
    } catch {
      setError('Не удалось убрать преподавателя.')
    }
  }

  async function handleArchive(offer: Offer) {
    if (!window.confirm(`Архивировать направление «${offer.title}»?`)) return
    setError(null)
    try {
      await archiveOffer(offer.id)
      setOffers((items) => items.filter((item) => item.id !== offer.id))
      if (editingId === offer.id) resetForm()
    } catch {
      setError('Не удалось архивировать направление.')
    }
  }

  async function handleLogout() {
    await logout()
    window.location.assign('/login')
  }

  if (loading) return <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}><CircularProgress /></Box>

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 3, md: 6 } }}>
      <Container maxWidth="lg">
        <Stack spacing={4}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
            <div>
              <BrandLink />
              <Typography component="h1" variant="h3">Направления</Typography>
            </div>
            <Stack direction="row" spacing={1}>
              <Button href="/admin" variant="outlined">Преподаватели</Button>
              <Button href="/admin/schedule" variant="outlined">Расписание</Button>
              <Button href="/admin/applications" variant="outlined">Заявки</Button>
              <Button variant="outlined" onClick={handleLogout}>Выйти</Button>
            </Stack>
          </Stack>

          {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}

          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'minmax(0, 1fr) 420px' }, gap: 3, alignItems: 'start' }}>
            <Stack spacing={2}>
              {offers.length === 0 && <Alert severity="info">Направлений пока нет. Создайте первое.</Alert>}
              {offers.map((offer) => {
                const assigned = new Set(offer.teachers.map((item) => item.teacherProfileId))
                const availableTeachers = teachers.filter((teacher) => !assigned.has(teacher.id))
                return (
                  <Box key={offer.id} sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3, border: '1px solid', borderColor: editingId === offer.id ? 'primary.main' : 'divider' }}>
                    <Stack direction="row" spacing={2} sx={{ justifyContent: 'space-between', alignItems: 'start' }}>
                      <div>
                        <Typography variant="h6">{offer.title}</Typography>
                        <Typography color="text.secondary">
                          {formatLabels[offer.format]} · {offer.defaultDurationMinutes} мин.
                          {offer.priceRubles !== null ? ` · ${offer.priceRubles.toLocaleString('ru-RU')} ₽` : ''}
                        </Typography>
                      </div>
                      <Typography color={offer.isPublished ? 'primary' : 'text.secondary'}>{offer.isPublished ? 'Опубликовано' : 'Черновик'}</Typography>
                    </Stack>
                    {offer.description && <Typography sx={{ mt: 2 }}>{offer.description}</Typography>}
                    {offer.goal && <Typography color="text.secondary" sx={{ mt: 1 }}>Цель: {offer.goal}</Typography>}

                    <Box sx={{ mt: 2, p: 2, bgcolor: 'background.default', borderRadius: 2 }}>
                      <Typography sx={{ fontWeight: 650, mb: 1 }}>Преподаватели</Typography>
                      {offer.teachers.length === 0 && <Typography variant="body2" color="text.secondary">Никто не назначен</Typography>}
                      <Stack spacing={1}>
                        {offer.teachers.map((assignment) => (
                          <Stack key={assignment.id} direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ justifyContent: 'space-between', alignItems: { sm: 'center' } }}>
                            <Typography variant="body2">
                              {assignment.teacherDisplayName} · {assignment.isPublished ? 'виден на сайте' : 'скрыт'}
                            </Typography>
                            <Stack direction="row" spacing={1}>
                              <Button size="small" onClick={() => handleAssignmentVisibility(offer, assignment.id)}>{assignment.isPublished ? 'Скрыть' : 'Показать'}</Button>
                              <Button size="small" color="error" onClick={() => handleRemoveAssignment(offer, assignment.id)}>Убрать</Button>
                            </Stack>
                          </Stack>
                        ))}
                      </Stack>
                      {availableTeachers.length > 0 && (
                        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} sx={{ mt: 2 }}>
                          <TextField select size="small" label="Добавить преподавателя" value={selectedTeachers[offer.id] ?? ''} onChange={(event) => setSelectedTeachers((current) => ({ ...current, [offer.id]: event.target.value }))} sx={{ minWidth: 240 }}>
                            {availableTeachers.map((teacher) => <MenuItem key={teacher.id} value={teacher.id}>{teacher.displayName}</MenuItem>)}
                          </TextField>
                          <Button variant="outlined" disabled={!selectedTeachers[offer.id]} onClick={() => handleAssign(offer)}>Назначить</Button>
                        </Stack>
                      )}
                    </Box>

                    <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
                      <Button size="small" onClick={() => startEditing(offer)}>Редактировать</Button>
                      <Button size="small" color="error" onClick={() => handleArchive(offer)}>Архивировать</Button>
                    </Stack>
                  </Box>
                )
              })}
            </Stack>

            <Box component="form" onSubmit={handleSubmit} sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3, position: { md: 'sticky' }, top: 24 }}>
              <Stack spacing={2}>
                <Typography variant="h5">{editingId ? 'Редактирование' : 'Новое направление'}</Typography>
                <TextField label="Название" required value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} />
                <TextField label="Описание" multiline minRows={3} value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
                <TextField label="Цель занятий" value={form.goal} onChange={(event) => setForm({ ...form, goal: event.target.value })} />
                <TextField label="Длительность, минут" type="number" required slotProps={{ htmlInput: { min: 15, max: 480 } }} value={form.defaultDurationMinutes} onChange={(event) => setForm({ ...form, defaultDurationMinutes: Number(event.target.value) })} />
                <TextField select label="Формат" value={form.format} onChange={(event) => setForm({ ...form, format: event.target.value as OfferInput['format'] })}>
                  <MenuItem value="online">Онлайн</MenuItem>
                  <MenuItem value="offline">Очно</MenuItem>
                  <MenuItem value="both">Онлайн и очно</MenuItem>
                </TextField>
                <TextField label="Цена, ₽" type="number" slotProps={{ htmlInput: { min: 0 } }} value={form.priceRubles ?? ''} onChange={(event) => setForm({ ...form, priceRubles: event.target.value === '' ? null : Number(event.target.value) })} helperText="Можно оставить пустой" />
                <FormControlLabel control={<Checkbox checked={form.isPublished} onChange={(event) => setForm({ ...form, isPublished: event.target.checked })} />} label="Показывать на сайте" />
                <Stack direction="row" spacing={1}>
                  <Button type="submit" variant="contained" disabled={saving}>{saving ? 'Сохраняем…' : editingId ? 'Сохранить' : 'Создать'}</Button>
                  {editingId && <Button type="button" onClick={resetForm}>Отмена</Button>}
                </Stack>
              </Stack>
            </Box>
          </Box>
        </Stack>
      </Container>
    </Box>
  )
}
