import { useEffect, useState, type FormEvent } from 'react'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  CircularProgress,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import {
  ApiError,
  changePassword,
  deleteSlot,
  getAdminSlots,
  getCurrentUser,
  logout,
  saveSlot,
  type Slot,
  type SlotInput,
} from '../../services/api'
import { slotTime, useCurrentTime } from '../../services/time'
const empty: SlotInput = {
  title: '',
  startsAt: '',
  endsAt: '',
  format: 'online',
  kind: 'individual',
  capacity: 1,
  occupied: 0,
  status: 'planned',
  published: false,
}
const statuses = {
  planned: 'Запланирован',
  completed: 'Завершён',
  cancelled: 'Отменён',
}
function localTime(value: string) {
  return new Date(Date.parse(value) + 3 * 3600000).toISOString().slice(0, 16)
}
export function AdminPage() {
  const now = useCurrentTime()
  const [slots, setSlots] = useState<Slot[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [username, setUsername] = useState('')
  const [editing, setEditing] = useState(false)
  const [id, setId] = useState<number>()
  const [form, setForm] = useState<SlotInput>(empty)
  const [busy, setBusy] = useState(false)
  const [removing, setRemoving] = useState<Slot>()
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  function handleError(e: unknown) {
    if (e instanceof ApiError && e.status === 401) {
      window.location.assign('/login')
      return
    }
    setError(e instanceof Error ? e.message : 'Не удалось выполнить действие.')
  }
  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      getCurrentUser(controller.signal),
      getAdminSlots(controller.signal),
    ])
      .then(([user, data]) => {
        setUsername(user.username)
        setSlots(data)
      })
      .catch((e) => {
        if (!controller.signal.aborted) handleError(e)
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [])
  function edit(slot?: Slot) {
    setError('')
    setId(slot?.id)
    setForm(
      slot
        ? {
            title: slot.title,
            startsAt: localTime(slot.startsAt),
            endsAt: localTime(slot.endsAt),
            format: slot.format,
            kind: slot.kind,
            capacity: slot.capacity,
            occupied: slot.occupied,
            status: slot.status,
            published: slot.published,
          }
        : { ...empty },
    )
    setEditing(true)
  }
  async function submit(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      await saveSlot(
        {
          ...form,
          startsAt: new Date(`${form.startsAt}:00+03:00`).toISOString(),
          endsAt: new Date(`${form.endsAt}:00+03:00`).toISOString(),
        },
        id,
      )
      setSlots(await getAdminSlots())
      setEditing(false)
    } catch (e) {
      handleError(e)
    } finally {
      setBusy(false)
    }
  }
  async function remove() {
    if (!removing) return
    setBusy(true)
    setError('')
    try {
      await deleteSlot(removing.id)
      setSlots((prev) => prev.filter((s) => s.id !== removing.id))
      setRemoving(undefined)
    } catch (e) {
      handleError(e)
    } finally {
      setBusy(false)
    }
  }
  async function updatePassword(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError('')
    try {
      await changePassword(currentPassword, newPassword)
      window.location.assign('/login')
    } catch (e) {
      handleError(e)
    } finally {
      setBusy(false)
    }
  }
  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        spacing={2}
        sx={{ justifyContent: 'space-between' }}
      >
        <BrandLink />
        <Stack direction="row" spacing={1}>
          <Button
            onClick={() => {
              setError('')
              setPasswordOpen(true)
            }}
          >
            Сменить пароль
          </Button>
          <Button
            onClick={async () => {
              try {
                await logout()
                window.location.assign('/login')
              } catch (e) {
                handleError(e)
              }
            }}
          >
            Выйти
          </Button>
        </Stack>
      </Stack>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        spacing={2}
        sx={{
          justifyContent: 'space-between',
          alignItems: { sm: 'center' },
          my: 4,
        }}
      >
        <Box>
          <Typography variant="h1">Расписание</Typography>
          <Typography color="text.secondary">
            Администратор: {username} · Время по Москве
          </Typography>
        </Box>
        <Button variant="contained" onClick={() => edit()} disabled={loading}>
          Добавить слот
        </Button>
      </Stack>
      {error && !editing && !passwordOpen && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      {loading ? (
        <CircularProgress aria-label="Загрузка" />
      ) : (
        <Stack spacing={2}>
          {!slots.length && (
            <Alert severity="info">
              Расписание пустое. Добавьте первый слот.
            </Alert>
          )}
          {slots.map((s) => (
            <Card key={s.id}>
              <CardContent>
                <Stack
                  direction={{ xs: 'column', md: 'row' }}
                  spacing={2}
                  sx={{ justifyContent: 'space-between' }}
                >
                  <Box>
                    <Typography variant="h3">{s.title}</Typography>
                    <Typography>
                      {slotTime(s.startsAt)} — {slotTime(s.endsAt)}
                    </Typography>
                    <Typography color="text.secondary">
                      {s.format === 'online' ? 'Онлайн' : 'Очно'} ·{' '}
                      {s.kind === 'individual' ? 'Индивидуально' : 'Группа'} ·
                      Занято {s.occupied} из {s.capacity}
                    </Typography>
                    <Typography>
                      {statuses[s.status]} ·{' '}
                      {s.published &&
                      s.status === 'planned' &&
                      s.occupied < s.capacity &&
                      Date.parse(s.startsAt) > now
                        ? 'На доске'
                        : 'Скрыт с доски'}
                    </Typography>
                  </Box>
                  <Stack direction="row" spacing={1}>
                    <Button onClick={() => edit(s)}>Изменить</Button>
                    <Button
                      color="error"
                      onClick={() => {
                        setError('')
                        setRemoving(s)
                      }}
                    >
                      Удалить
                    </Button>
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      )}
      <Dialog
        open={editing}
        onClose={() => {
          if (!busy) setEditing(false)
        }}
        fullWidth
        maxWidth="sm"
      >
        <Box component="form" onSubmit={submit}>
          <DialogTitle>{id ? 'Изменить слот' : 'Новый слот'}</DialogTitle>
          <DialogContent>
            <Stack spacing={2} sx={{ pt: 1 }}>
              {error && <Alert severity="error">{error}</Alert>}
              <TextField
                label="Направление"
                required
                value={form.title}
                onChange={(e) => setForm({ ...form, title: e.target.value })}
                slotProps={{ htmlInput: { maxLength: 120 } }}
              />
              <TextField
                label="Начало (Москва)"
                type="datetime-local"
                required
                value={form.startsAt}
                onChange={(e) => setForm({ ...form, startsAt: e.target.value })}
                slotProps={{ inputLabel: { shrink: true } }}
              />
              <TextField
                label="Окончание (Москва)"
                type="datetime-local"
                required
                value={form.endsAt}
                onChange={(e) => setForm({ ...form, endsAt: e.target.value })}
                slotProps={{
                  inputLabel: { shrink: true },
                  htmlInput: { min: form.startsAt },
                }}
              />
              <TextField
                select
                label="Формат"
                value={form.format}
                onChange={(e) =>
                  setForm({
                    ...form,
                    format: e.target.value as SlotInput['format'],
                  })
                }
              >
                <MenuItem value="online">Онлайн</MenuItem>
                <MenuItem value="offline">Очно</MenuItem>
              </TextField>
              <TextField
                select
                label="Тип"
                value={form.kind}
                onChange={(e) => {
                  const kind = e.target.value as SlotInput['kind']
                  setForm({
                    ...form,
                    kind,
                    capacity: kind === 'individual' ? 1 : form.capacity,
                    occupied:
                      kind === 'individual'
                        ? Math.min(form.occupied, 1)
                        : form.occupied,
                  })
                }}
              >
                <MenuItem value="individual">Индивидуально</MenuItem>
                <MenuItem value="group">Группа</MenuItem>
              </TextField>
              <TextField
                label="Всего мест"
                type="number"
                required
                disabled={form.kind === 'individual'}
                value={form.capacity}
                onChange={(e) =>
                  setForm({ ...form, capacity: Number(e.target.value) })
                }
                slotProps={{ htmlInput: { min: 1, max: 1000, step: 1 } }}
              />
              <TextField
                label="Занято мест"
                type="number"
                required
                value={form.occupied}
                onChange={(e) =>
                  setForm({ ...form, occupied: Number(e.target.value) })
                }
                slotProps={{
                  htmlInput: { min: 0, max: form.capacity, step: 1 },
                }}
              />
              <TextField
                select
                label="Статус"
                value={form.status}
                onChange={(e) =>
                  setForm({
                    ...form,
                    status: e.target.value as SlotInput['status'],
                  })
                }
              >
                {Object.entries(statuses).map(([key, label]) => (
                  <MenuItem key={key} value={key}>
                    {label}
                  </MenuItem>
                ))}
              </TextField>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={form.published}
                    onChange={(e) =>
                      setForm({ ...form, published: e.target.checked })
                    }
                  />
                }
                label="Показывать на доске при наличии свободных мест"
              />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={() => setEditing(false)}>
              Отмена
            </Button>
            <Button disabled={busy} type="submit" variant="contained">
              Сохранить
            </Button>
          </DialogActions>
        </Box>
      </Dialog>
      <Dialog
        open={!!removing}
        onClose={() => {
          if (!busy) setRemoving(undefined)
        }}
      >
        <DialogTitle>Удалить слот «{removing?.title}»?</DialogTitle>
        <DialogContent>Слот будет удалён из расписания.</DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={() => setRemoving(undefined)}>
            Отмена
          </Button>
          <Button disabled={busy} color="error" onClick={remove}>
            Удалить
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog
        open={passwordOpen}
        onClose={() => {
          if (!busy) setPasswordOpen(false)
        }}
        fullWidth
        maxWidth="sm"
      >
        <Box component="form" onSubmit={updatePassword}>
          <DialogTitle>Сменить пароль</DialogTitle>
          <DialogContent>
            <Stack spacing={2} sx={{ pt: 1 }}>
              {error && <Alert severity="error">{error}</Alert>}
              <TextField
                type="password"
                label="Текущий пароль"
                autoComplete="current-password"
                required
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
              />
              <TextField
                type="password"
                label="Новый пароль"
                autoComplete="new-password"
                required
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                slotProps={{ htmlInput: { minLength: 12, maxLength: 72 } }}
                helperText="Не менее 12 символов. После смены нужно войти заново."
              />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={() => setPasswordOpen(false)}>
              Отмена
            </Button>
            <Button type="submit" disabled={busy}>
              Сохранить
            </Button>
          </DialogActions>
        </Box>
      </Dialog>
    </Container>
  )
}
