import { useEffect, useState, type FormEvent } from 'react'
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Container,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material'
import { AdminHeader } from '../../components/AdminHeader'
import {
  ApiError,
  deleteDirection,
  getDirections,
  saveDirection,
  type Direction,
} from '../../services/api'
export function DirectionsPage() {
  const [directions, setDirections] = useState<Direction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState(false)
  const [id, setId] = useState<number>()
  const [name, setName] = useState('')
  const [removing, setRemoving] = useState<Direction>()
  const [busy, setBusy] = useState(false)
  function handleError(error: unknown) {
    if (error instanceof ApiError && error.status === 401) {
      window.location.assign('/login')
      return
    }
    setError(
      error instanceof Error ? error.message : 'Не удалось выполнить действие.',
    )
  }
  useEffect(() => {
    const controller = new AbortController()
    getDirections(controller.signal)
      .then(setDirections)
      .catch((error) => {
        if (!controller.signal.aborted) handleError(error)
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [])
  function edit(direction?: Direction) {
    setError('')
    setId(direction?.id)
    setName(direction?.name ?? '')
    setEditing(true)
  }
  async function submit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      await saveDirection(name, id)
      setDirections(await getDirections())
      setEditing(false)
    } catch (error) {
      handleError(error)
    } finally {
      setBusy(false)
    }
  }
  async function remove() {
    if (!removing) return
    setBusy(true)
    setError('')
    try {
      await deleteDirection(removing.id)
      setDirections((prev) =>
        prev.filter((direction) => direction.id !== removing.id),
      )
      setRemoving(undefined)
    } catch (error) {
      handleError(error)
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
          my: 4,
          justifyContent: 'space-between',
          alignItems: { sm: 'center' },
        }}
      >
        <Box>
          <Typography variant="h1">Направления</Typography>
          <Typography color="text.secondary">
            Справочник занятий для расписания
          </Typography>
        </Box>
        <Button variant="contained" disabled={loading} onClick={() => edit()}>
          Добавить направление
        </Button>
      </Stack>
      {error && !editing && !removing && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      {loading ? (
        <CircularProgress aria-label="Загрузка направлений" />
      ) : !directions.length ? (
        <Alert severity="info">
          Добавьте первое направление, чтобы выбирать его при создании слотов.
        </Alert>
      ) : (
        <TableContainer>
          <Table aria-label="Направления">
            <TableHead>
              <TableRow>
                <TableCell>Название</TableCell>
                <TableCell align="right">Действия</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {directions.map((direction) => (
                <TableRow key={direction.id}>
                  <TableCell sx={{ overflowWrap: 'anywhere' }}>
                    {direction.name}
                  </TableCell>
                  <TableCell align="right">
                    <Stack
                      direction={{ xs: 'column', sm: 'row' }}
                      sx={{ justifyContent: 'flex-end' }}
                    >
                      <Button onClick={() => edit(direction)}>Изменить</Button>
                      <Button
                        color="error"
                        onClick={() => {
                          setError('')
                          setRemoving(direction)
                        }}
                      >
                        Удалить
                      </Button>
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
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
          <DialogTitle>
            {id ? 'Изменить направление' : 'Новое направление'}
          </DialogTitle>
          <DialogContent>
            <Stack spacing={2} sx={{ pt: 1 }}>
              {error && <Alert severity="error">{error}</Alert>}
              <TextField
                label="Название направления"
                autoFocus
                required
                fullWidth
                value={name}
                onChange={(event) => setName(event.target.value)}
                slotProps={{ htmlInput: { maxLength: 120 } }}
              />
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button disabled={busy} onClick={() => setEditing(false)}>
              Отмена
            </Button>
            <Button type="submit" variant="contained" disabled={busy}>
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
        <DialogTitle>Удалить направление «{removing?.name}»?</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            {error && <Alert severity="error">{error}</Alert>}
            <Typography>
              Удаление доступно, если направление не используется в расписании.
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button disabled={busy} onClick={() => setRemoving(undefined)}>
            Отмена
          </Button>
          <Button color="error" disabled={busy} onClick={remove}>
            Удалить
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  )
}
