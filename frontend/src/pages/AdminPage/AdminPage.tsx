import { useEffect, useState, type FormEvent } from 'react'
import {
  Alert,
  Box,
  Button,
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
import { AdminHeader } from '../../components/AdminHeader'
import {
  ApiError,
  deleteSlot,
  getAdminSlots,
  getDirections,
  type Direction,
  requireSession,
  saveSlot,
  slotStatusLabels,
  type Slot,
  type SlotInput,
} from '../../services/api'
import { SchoolSchedule } from '../../components/SchoolSchedule/SchoolSchedule'
const empty: SlotInput = {
  directionId: 0,
  level: '',
  startsAt: '',
  endsAt: '',
  format: 'online',
  kind: 'individual',
  capacity: 1,
  occupied: 0,
  status: 'planned',
  published: false,
}
function localTime(value: string) {
  return new Date(Date.parse(value) + 3 * 3600000).toISOString().slice(0, 16)
}
export function AdminPage() {
  const [slots, setSlots] = useState<Slot[]>([])
  const [directions, setDirections] = useState<Direction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState(false)
  const [id, setId] = useState<number>()
  const [form, setForm] = useState<SlotInput>(empty)
  const [busy, setBusy] = useState(false)
  const [removing, setRemoving] = useState<Slot>()
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
      requireSession(controller.signal),
      getAdminSlots(controller.signal),
      getDirections(controller.signal),
    ])
      .then(([, data, directions]) => {
        setSlots(data)
        setDirections(directions)
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
            directionId: slot.directionId,
            level: slot.level,
            startsAt: localTime(slot.startsAt),
            endsAt: localTime(slot.endsAt),
            format: slot.format,
            kind: slot.kind,
            capacity: slot.capacity,
            occupied: slot.occupied,
            status: slot.status,
            published: slot.published,
          }
        : { ...empty, directionId: directions[0]?.id ?? 0 },
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
  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <AdminHeader onError={handleError} />
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
            Управление слотами · Время по Москве
          </Typography>
        </Box>
        <Button
          variant="contained"
          onClick={() => edit()}
          disabled={loading || directions.length === 0}
        >
          Добавить слот
        </Button>
      </Stack>
      {!loading && directions.length === 0 && (
        <Alert severity="info" sx={{ mb: 2 }}>
          Сначала <Button href="/admin/directions">добавьте направление</Button>
          , затем создайте слот.
        </Alert>
      )}
      {error && !editing && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      {loading ? (
        <CircularProgress aria-label="Загрузка" />
      ) : (
        <SchoolSchedule
          slots={slots}
          onEdit={edit}
          onDelete={(slot) => {
            setError('')
            setRemoving(slot)
          }}
        />
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
                select
                label="Направление"
                required
                value={form.directionId || ''}
                onChange={(e) =>
                  setForm({ ...form, directionId: Number(e.target.value) })
                }
              >
                {directions.map((direction) => (
                  <MenuItem key={direction.id} value={direction.id}>
                    {direction.name}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                label="Уровень занятия"
                value={form.level}
                onChange={(e) => setForm({ ...form, level: e.target.value })}
                helperText="Например: Начальный, A2 или 7 класс"
                slotProps={{ htmlInput: { maxLength: 80 } }}
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
                {Object.entries(slotStatusLabels).map(([key, label]) => (
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
    </Container>
  )
}
