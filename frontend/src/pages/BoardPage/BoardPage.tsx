import { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Container,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import { getPublicSlots, type PublicSlot } from '../../services/api'
import { slotTime, useCurrentTime } from '../../services/time'
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
      s.title
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
        <Typography variant="h1">Свободное время для занятий</Typography>
        <Typography color="text.secondary" sx={{ mt: 2 }}>
          Доступные слоты и свободные места. Всё время указано по Москве.
        </Typography>
      </Box>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ mb: 3 }}>
        <TextField
          label="Поиск по направлению"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          fullWidth
        />
        <Button
          onClick={() => {
            setLoading(true)
            setError('')
            setVersion((v) => v + 1)
          }}
        >
          Обновить
        </Button>
      </Stack>
      {loading ? (
        <CircularProgress aria-label="Загрузка расписания" />
      ) : error ? (
        <Alert severity="error">{error}</Alert>
      ) : !available.length ? (
        <Alert severity="info">
          {slots.length
            ? 'Нет слотов по выбранному направлению.'
            : 'Сейчас нет свободных слотов. Загляните позже.'}
        </Alert>
      ) : (
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' },
            gap: 2,
          }}
        >
          {available.map((s) => (
            <Card key={s.id}>
              <CardContent>
                <Typography variant="h3">{s.title}</Typography>
                <Typography sx={{ my: 2 }}>{slotTime(s.startsAt)}</Typography>
                <Stack
                  direction="row"
                  spacing={1}
                  useFlexGap
                  sx={{ flexWrap: 'wrap' }}
                >
                  <Chip label={s.format === 'online' ? 'Онлайн' : 'Очно'} />
                  <Chip
                    label={s.kind === 'individual' ? 'Индивидуально' : 'Группа'}
                  />
                  <Chip
                    label={`${Math.round((Date.parse(s.endsAt) - Date.parse(s.startsAt)) / 60000)} мин`}
                  />
                  <Chip
                    color="primary"
                    label={`Свободных мест: ${s.freePlaces}`}
                  />
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Box>
      )}
    </Container>
  )
}
