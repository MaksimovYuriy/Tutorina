import { useState, type FormEvent } from 'react'
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { ApiError, login } from '../../services/api'
import './LoginPage.css'

export function LoginPage() {
  const [key, setKey] = useState('')
  const [showKey, setShowKey] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login(key.trim())
      window.location.assign('/admin')
    } catch (caughtError: unknown) {
      if (caughtError instanceof ApiError && caughtError.status === 429) {
        setError(
          'Слишком много попыток входа. Подождите минуту и попробуйте снова.',
        )
      } else if (
        caughtError instanceof ApiError &&
        caughtError.status === 401
      ) {
        setError('Ключ доступа не подошёл. Проверьте его и попробуйте снова.')
      } else {
        setError('Не удалось войти. Пожалуйста, попробуйте ещё раз.')
      }
      setSubmitting(false)
    }
  }

  return (
    <main className="login-page">
      <section className="login-story" aria-label="О сервисе Tutorina">
        <div className="story-orb story-orb-one" />
        <div className="story-orb story-orb-two" />

        <a className="brand" href="/" aria-label="Tutorina">
          <span className="brand-mark" aria-hidden="true">
            T
          </span>
          <span>Tutorina</span>
        </a>

        <div className="story-copy">
          <p className="eyebrow">Управление расписанием</p>
          <Typography component="h1" variant="h1" className="story-title">
            Занятия в одном
            <br />
            спокойном ритме
          </Typography>
          <Typography className="story-description">
            Публикуйте свободные слоты и управляйте занятостью в одном месте.
          </Typography>
        </div>

        <div className="schedule-preview" aria-hidden="true">
          <div className="preview-heading">
            <div>
              <span className="preview-caption">Сегодня</span>
              <strong>4 занятия</strong>
            </div>
            <span className="preview-date">30 сен</span>
          </div>
          <div className="lesson-row">
            <span className="lesson-time">14:00</span>
            <span className="lesson-dot lesson-dot-violet" />
            <span>
              <strong>Английский язык</strong>
              <small>Индивидуально · 60 минут</small>
            </span>
          </div>
          <div className="lesson-row">
            <span className="lesson-time">16:30</span>
            <span className="lesson-dot lesson-dot-yellow" />
            <span>
              <strong>Подготовка к экзамену</strong>
              <small>Группа · 3 свободных места</small>
            </span>
          </div>
        </div>
      </section>

      <section className="login-panel">
        <Box component="form" className="login-form" onSubmit={handleSubmit}>
          <Stack spacing={3.5}>
            <div>
              <a
                className="mobile-brand"
                href="/"
                aria-label="Tutorina — на главную"
              >
                Tutorina
              </a>
              <Typography component="h2" variant="h2" className="login-title">
                С возвращением
              </Typography>
              <Typography color="text.secondary" className="login-subtitle">
                Войдите в админку расписания
              </Typography>
            </div>

            {error && <Alert severity="error">{error}</Alert>}

            <Stack spacing={2.25}>
              <TextField
                label="Ключ доступа"
                type={showKey ? 'text' : 'password'}
                value={key}
                onChange={(event) => setKey(event.target.value)}
                autoComplete="off"
                required
                autoFocus
                fullWidth
                slotProps={{
                  input: {
                    endAdornment: (
                      <InputAdornment position="end">
                        <Button
                          type="button"
                          className="access-key-toggle"
                          onClick={() => setShowKey((visible) => !visible)}
                          aria-label={showKey ? 'Скрыть ключ' : 'Показать ключ'}
                        >
                          {showKey ? 'Скрыть' : 'Показать'}
                        </Button>
                      </InputAdornment>
                    ),
                  },
                }}
              />
            </Stack>

            <Button
              type="submit"
              variant="contained"
              size="large"
              disabled={submitting}
              fullWidth
            >
              {submitting ? (
                <CircularProgress
                  size={22}
                  color="inherit"
                  aria-label="Выполняется вход"
                />
              ) : (
                'Войти в кабинет'
              )}
            </Button>

            <Typography className="login-help" color="text.secondary">
              Введите ключ доступа к админке расписания.
            </Typography>
          </Stack>
        </Box>

        <p className="login-footer">Tutorina · админка расписания</p>
      </section>
    </main>
  )
}
