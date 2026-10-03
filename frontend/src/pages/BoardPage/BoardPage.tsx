import { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Container,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import { getPublicSlots, type PublicSlot } from '../../services/api'
import { SchoolSchedule } from '../../components/SchoolSchedule/SchoolSchedule'
import { useCurrentTime } from '../../services/time'
export function BoardPage() {
  const now = useCurrentTime()
  const [slots, setSlots] = useState<PublicSlot[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState('')
  const [version, setVersion] = useState(0)
  useEffect(() => {
    const controller = new AbortController()
    getPublicSlots(controller.signal)
      .then(setSlots)
      .catch((e) => {
        if (!controller.signal.aborted) setError(e.message)
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [version])
  const available = slots.filter(
    (s) =>
      `${s.title} ${s.level}`
        .toLocaleLowerCase('ru')
        .includes(filter.toLocaleLowerCase('ru')) &&
      new Date(s.startsAt).getTime() > now,
  )
  return (
    <Container maxWidth="lg" sx={{ py: { xs: 3, md: 6 } }}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        spacing={2}
        sx={{ justifyContent: 'space-between', alignItems: { sm: 'center' } }}
      >
        <BrandLink />
        <Button href="/login">Вход администратора</Button>
      </Stack>
      <Box sx={{ my: 6 }}>
        <Typography variant="h1">Расписание занятий</Typography>
        <Typography color="text.secondary" sx={{ mt: 2 }}>
          Как на школьной доске: выберите неделю и найдите свободное время для
          занятия.
        </Typography>
      </Box>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ mb: 3 }}>
        <TextField
          label="Поиск по направлению или уровню"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          fullWidth
        />
        <Button
          disabled={loading}
          onClick={() => {
            setLoading(true)
            setError('')
            setVersion((v) => v + 1)
          }}
        >
          Обновить
        </Button>
      </Stack>
      {!loading && !error && filter.trim() && !available.length && (
        <Alert severity="info" sx={{ mb: 2 }}>
          По вашему запросу занятий не найдено.
        </Alert>
      )}
      {loading && version === 0 ? (
        <CircularProgress aria-label="Загрузка расписания" />
      ) : error ? (
        <Alert severity="error">{error}</Alert>
      ) : (
        <SchoolSchedule slots={available} />
      )}
    </Container>
  )
}
