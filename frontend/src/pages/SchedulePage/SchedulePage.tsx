import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Alert, Box, Button, Checkbox, CircularProgress, Container, FormControlLabel, MenuItem, Stack, TextField, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import {
  ApiError,
  archiveAdminLesson,
  archiveMyLesson,
  createAdminLesson,
  createMyLesson,
  getAdminLessons,
  getCurrentUser,
  getMyLessons,
  getMyOffers,
  getOffers,
  logout,
  updateAdminLesson,
  updateMyLesson,
  type Lesson,
  type LessonInput,
  type Offer,
} from '../../services/api'

const MOSCOW_TIME_ZONE = 'Europe/Moscow'

interface ScheduleForm {
  teacherOfferId: string
  description: string
  startsAt: string
  endsAt: string
  deliveryFormat: LessonInput['deliveryFormat']
  lessonType: LessonInput['lessonType']
  capacity: number
  status: LessonInput['status']
  enrollmentOpen: boolean
  groupGoal: string
  groupLevel: string
}

const statusLabels = { planned: 'Запланировано', completed: 'Проведено', cancelled: 'Отменено' }
const typeLabels = { individual: 'Индивидуальное', group: 'Групповое' }
const deliveryLabels = { online: 'Онлайн', offline: 'Очно' }

export function SchedulePage() {
  const isAdmin = window.location.pathname.startsWith('/admin/')
  const [offers, setOffers] = useState<Offer[]>([])
  const [lessons, setLessons] = useState<Lesson[]>([])
  const [form, setForm] = useState<ScheduleForm>(() => emptyForm())
  const [editingId, setEditingId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const assignmentOptions = useMemo(() => offers.flatMap((offer) => offer.teachers.map((assignment) => ({
    id: assignment.id,
    label: isAdmin ? `${offer.title} — ${assignment.teacherDisplayName}` : offer.title,
    offerFormat: offer.format,
    duration: assignment.durationMinutes ?? offer.defaultDurationMinutes,
  }))), [offers, isAdmin])

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      getCurrentUser(controller.signal),
      isAdmin ? getOffers(controller.signal) : getMyOffers(controller.signal),
      isAdmin ? getAdminLessons(controller.signal) : getMyLessons(controller.signal),
    ]).then(([user, offerItems, lessonItems]) => {
      const requiredRole = isAdmin ? 'admin' : 'teacher'
      if (!user.roles.includes(requiredRole)) {
        window.location.replace(user.roles.includes('admin') ? '/admin' : '/teacher')
        return
      }
      setOffers(offerItems)
      setLessons(lessonItems)
    }).catch((caught: unknown) => {
      if (caught instanceof DOMException && caught.name === 'AbortError') return
      if (caught instanceof ApiError && caught.status === 401) {
        window.location.replace('/login')
        return
      }
      setError('Не удалось загрузить расписание.')
    }).finally(() => setLoading(false))
    return () => controller.abort()
  }, [isAdmin])

  function resetForm() {
    setForm(emptyForm())
    setEditingId(null)
  }

  function selectAssignment(id: string) {
    const option = assignmentOptions.find((item) => item.id === id)
    const start = fromMoscowLocal(form.startsAt)
    const end = new Date(start.getTime() + (option?.duration ?? 60) * 60_000)
    const deliveryFormat = option?.offerFormat === 'offline' ? 'offline' : option?.offerFormat === 'online' ? 'online' : form.deliveryFormat
    setForm({ ...form, teacherOfferId: id, endsAt: toMoscowLocal(end.toISOString()), deliveryFormat })
  }

  function startEditing(lesson: Lesson) {
    setEditingId(lesson.id)
    setForm({
      teacherOfferId: lesson.teacherOfferId,
      description: lesson.description,
      startsAt: toMoscowLocal(lesson.startsAt),
      endsAt: toMoscowLocal(lesson.endsAt),
      deliveryFormat: lesson.deliveryFormat,
      lessonType: lesson.lessonType,
      capacity: lesson.capacity,
      status: lesson.status,
      enrollmentOpen: lesson.enrollmentOpen,
      groupGoal: lesson.groupGoal,
      groupLevel: lesson.groupLevel,
    })
    setError(null)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!form.teacherOfferId) {
      setError('Выберите направление и преподавателя.')
      return
    }
    const input: LessonInput = {
      teacherOfferId: Number(form.teacherOfferId),
      description: form.description,
      startsAt: fromMoscowLocal(form.startsAt).toISOString(),
      endsAt: fromMoscowLocal(form.endsAt).toISOString(),
      deliveryFormat: form.deliveryFormat,
      lessonType: form.lessonType,
      capacity: form.lessonType === 'individual' ? 1 : form.capacity,
      status: form.status,
      enrollmentOpen: form.enrollmentOpen,
      groupGoal: form.lessonType === 'group' ? form.groupGoal : '',
      groupLevel: form.lessonType === 'group' ? form.groupLevel : '',
    }
    setSaving(true)
    setError(null)
    try {
      const saved = editingId
        ? await (isAdmin ? updateAdminLesson(editingId, input) : updateMyLesson(editingId, input))
        : await (isAdmin ? createAdminLesson(input) : createMyLesson(input))
      setLessons((items) => (items.some((item) => item.id === saved.id)
        ? items.map((item) => item.id === saved.id ? saved : item)
        : [...items, saved]).sort((a, b) => a.startsAt.localeCompare(b.startsAt)))
      resetForm()
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 409) {
        setError('В это время у преподавателя уже есть другое занятие.')
      } else {
        setError(caught instanceof ApiError ? caught.message : 'Не удалось сохранить занятие.')
      }
    } finally {
      setSaving(false)
    }
  }

  async function handleArchive(lesson: Lesson) {
    if (!window.confirm(`Удалить занятие «${lesson.offerTitle}» из расписания?`)) return
    setError(null)
    try {
      await (isAdmin ? archiveAdminLesson(lesson.id) : archiveMyLesson(lesson.id))
      setLessons((items) => items.filter((item) => item.id !== lesson.id))
      if (editingId === lesson.id) resetForm()
    } catch {
      setError('Не удалось удалить занятие.')
    }
  }

  async function handleLogout() {
    await logout()
    window.location.assign('/login')
  }

  const selectedOption = assignmentOptions.find((item) => item.id === form.teacherOfferId)
  const allowedDelivery = selectedOption?.offerFormat === 'both' || !selectedOption
    ? ['online', 'offline'] as const
    : [selectedOption.offerFormat] as const

  if (loading) return <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}><CircularProgress /></Box>

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 3, md: 6 } }}>
      <Container maxWidth="lg">
        <Stack spacing={4}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
            <div>
              <BrandLink />
              <Typography component="h1" variant="h3">Расписание</Typography>
              <Typography color="text.secondary">Время указано по Москве</Typography>
            </div>
            <Stack direction="row" spacing={1}>
              <Button href={isAdmin ? '/admin' : '/teacher'} variant="outlined">Назад в кабинет</Button>
              <Button href={isAdmin ? '/admin/applications' : '/teacher/applications'} variant="outlined">Заявки</Button>
              <Button variant="outlined" onClick={handleLogout}>Выйти</Button>
            </Stack>
          </Stack>

          {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}

          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'minmax(0, 1fr) 420px' }, gap: 3, alignItems: 'start' }}>
            <Stack spacing={2}>
              {lessons.length === 0 && <Alert severity="info">В выбранном периоде занятий пока нет.</Alert>}
              {lessons.map((lesson) => (
                <Box key={lesson.id} component="article" sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3, border: '1px solid', borderColor: editingId === lesson.id ? 'primary.main' : 'divider' }}>
                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
                    <div>
                      <Typography variant="h6">{lesson.offerTitle}</Typography>
                      <Typography color="text.secondary">{lesson.teacherDisplayName}</Typography>
                    </div>
                    <Typography sx={{ fontWeight: 650 }}>{formatScheduleTime(lesson.startsAt, lesson.endsAt)}</Typography>
                  </Stack>
                  <Typography sx={{ mt: 2 }}>
                    {typeLabels[lesson.lessonType]} · {deliveryLabels[lesson.deliveryFormat]} · мест: {lesson.capacity}
                    {lesson.priceRubles !== null ? ` · ${lesson.priceRubles.toLocaleString('ru-RU')} ₽` : ''}
                  </Typography>
                  <Typography color="text.secondary" sx={{ mt: 0.5 }}>
                    {statusLabels[lesson.status]} · запись {lesson.enrollmentOpen ? 'открыта' : 'закрыта'}
                  </Typography>
                  {lesson.description && <Typography sx={{ mt: 1 }}>{lesson.description}</Typography>}
                  {lesson.lessonType === 'group' && <Typography sx={{ mt: 1 }}>{lesson.groupLevel} · {lesson.groupGoal}</Typography>}
                  <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
                    <Button size="small" onClick={() => startEditing(lesson)}>Редактировать</Button>
                    <Button size="small" color="error" onClick={() => handleArchive(lesson)}>Удалить</Button>
                  </Stack>
                </Box>
              ))}
            </Stack>

            <Box component="form" onSubmit={handleSubmit} sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3, position: { md: 'sticky' }, top: 24 }}>
              <Stack spacing={2}>
                <Typography variant="h5">{editingId ? 'Редактирование занятия' : 'Новое занятие'}</Typography>
                {assignmentOptions.length === 0 && <Alert severity="warning">Сначала администратор должен назначить преподавателя на направление.</Alert>}
                <TextField select required label={isAdmin ? 'Направление и преподаватель' : 'Направление'} value={form.teacherOfferId} onChange={(event) => selectAssignment(event.target.value)} disabled={assignmentOptions.length === 0}>
                  {assignmentOptions.map((option) => <MenuItem key={option.id} value={option.id}>{option.label}</MenuItem>)}
                </TextField>
                <TextField label="Описание занятия" multiline minRows={3} value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} helperText="Напишите, как проходит урок и как вы свяжетесь с учеником" />
                <TextField label="Начало" type="datetime-local" required value={form.startsAt} onChange={(event) => setForm({ ...form, startsAt: event.target.value })} slotProps={{ inputLabel: { shrink: true } }} />
                <TextField label="Окончание" type="datetime-local" required value={form.endsAt} onChange={(event) => setForm({ ...form, endsAt: event.target.value })} slotProps={{ inputLabel: { shrink: true } }} />
                <TextField select label="Формат" value={form.deliveryFormat} onChange={(event) => setForm({ ...form, deliveryFormat: event.target.value as ScheduleForm['deliveryFormat'] })}>
                  {allowedDelivery.map((value) => <MenuItem key={value} value={value}>{deliveryLabels[value]}</MenuItem>)}
                </TextField>
                <TextField select label="Тип занятия" value={form.lessonType} onChange={(event) => {
                  const lessonType = event.target.value as ScheduleForm['lessonType']
                  setForm({ ...form, lessonType, capacity: lessonType === 'individual' ? 1 : Math.max(form.capacity, 2) })
                }}>
                  <MenuItem value="individual">Индивидуальное</MenuItem>
                  <MenuItem value="group">Групповое</MenuItem>
                </TextField>
                <TextField label="Количество мест" type="number" required disabled={form.lessonType === 'individual'} value={form.lessonType === 'individual' ? 1 : form.capacity} onChange={(event) => setForm({ ...form, capacity: Number(event.target.value) })} slotProps={{ htmlInput: { min: form.lessonType === 'group' ? 2 : 1 } }} />
                {form.lessonType === 'group' && (
                  <>
                    <TextField label="Уровень группы" required value={form.groupLevel} onChange={(event) => setForm({ ...form, groupLevel: event.target.value })} placeholder="Например, A2" />
                    <TextField label="Цель группы" required multiline minRows={2} value={form.groupGoal} onChange={(event) => setForm({ ...form, groupGoal: event.target.value })} />
                  </>
                )}
                <TextField select label="Статус" value={form.status} onChange={(event) => setForm({ ...form, status: event.target.value as ScheduleForm['status'] })}>
                  <MenuItem value="planned">Запланировано</MenuItem>
                  <MenuItem value="completed">Проведено</MenuItem>
                  <MenuItem value="cancelled">Отменено</MenuItem>
                </TextField>
                <FormControlLabel control={<Checkbox checked={form.enrollmentOpen} disabled={form.status !== 'planned'} onChange={(event) => setForm({ ...form, enrollmentOpen: event.target.checked })} />} label="Запись открыта" />
                <Stack direction="row" spacing={1}>
                  <Button type="submit" variant="contained" disabled={saving || assignmentOptions.length === 0}>{saving ? 'Сохраняем…' : editingId ? 'Сохранить' : 'Создать занятие'}</Button>
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

function emptyForm(): ScheduleForm {
  const start = new Date(Date.now() + 60 * 60_000)
  start.setUTCMinutes(0, 0, 0)
  const end = new Date(start.getTime() + 60 * 60_000)
  return {
    teacherOfferId: '', description: '', startsAt: toMoscowLocal(start.toISOString()), endsAt: toMoscowLocal(end.toISOString()),
    deliveryFormat: 'online', lessonType: 'individual', capacity: 1, status: 'planned',
    enrollmentOpen: true, groupGoal: '', groupLevel: '',
  }
}

function fromMoscowLocal(value: string): Date {
  return new Date(`${value}:00+03:00`)
}

function toMoscowLocal(value: string): string {
  const parts = new Intl.DateTimeFormat('sv-SE', {
    timeZone: MOSCOW_TIME_ZONE, year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', hourCycle: 'h23',
  }).formatToParts(new Date(value))
  const get = (type: Intl.DateTimeFormatPartTypes) => parts.find((part) => part.type === type)?.value ?? ''
  return `${get('year')}-${get('month')}-${get('day')}T${get('hour')}:${get('minute')}`
}

function formatScheduleTime(startsAt: string, endsAt: string): string {
  const formatter = new Intl.DateTimeFormat('ru-RU', {
    timeZone: MOSCOW_TIME_ZONE, day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit',
  })
  const endFormatter = new Intl.DateTimeFormat('ru-RU', { timeZone: MOSCOW_TIME_ZONE, hour: '2-digit', minute: '2-digit' })
  return `${formatter.format(new Date(startsAt))}–${endFormatter.format(new Date(endsAt))}`
}
