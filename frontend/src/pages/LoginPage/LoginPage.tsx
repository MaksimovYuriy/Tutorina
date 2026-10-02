import { useState, type FormEvent } from 'react'
import { Alert, Box, Button, CircularProgress, InputAdornment, Stack, TextField, Typography } from '@mui/material'
import { ApiError, getCurrentUser, login } from '../../services/api'
import './LoginPage.css'

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login(email, password)
      const user = await getCurrentUser()
      window.location.assign(user.roles.includes('admin') ? '/admin' : '/teacher')
    } catch (caughtError: unknown) {
      if (caughtError instanceof ApiError && caughtError.status === 401) {
        setError('Проверьте почту и пароль — они не подошли.')
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
          <span className="brand-mark" aria-hidden="true">T</span>
          <span>Tutorina</span>
        </a>

        <div className="story-copy">
          <p className="eyebrow">Пространство преподавателей</p>
          <Typography component="h1" variant="h1" className="story-title">
            Занятия в одном<br />спокойном ритме
          </Typography>
          <Typography className="story-description">
            Расписание, заявки и ученики собраны рядом — чтобы оставалось больше времени на преподавание.
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
            <span><strong>Английский язык</strong><small>Индивидуально · 60 минут</small></span>
          </div>
          <div className="lesson-row">
            <span className="lesson-time">16:30</span>
            <span className="lesson-dot lesson-dot-yellow" />
            <span><strong>Подготовка к экзамену</strong><small>Группа · 3 ученика</small></span>
          </div>
        </div>
      </section>

      <section className="login-panel">
        <Box component="form" className="login-form" onSubmit={handleSubmit} noValidate>
          <Stack spacing={3.5}>
            <div>
              <a className="mobile-brand" href="/" aria-label="Tutorina — на главную">Tutorina</a>
              <Typography component="h2" variant="h2" className="login-title">
                С возвращением
              </Typography>
              <Typography color="text.secondary" className="login-subtitle">
                Войдите в личный кабинет преподавателя
              </Typography>
            </div>

            {error && <Alert severity="error">{error}</Alert>}

            <Stack spacing={2.25}>
              <TextField
                label="Электронная почта"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                autoComplete="email"
                autoFocus
                required
                fullWidth
              />
              <TextField
                label="Пароль"
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                autoComplete="current-password"
                required
                fullWidth
                slotProps={{
                  input: {
                    endAdornment: (
                      <InputAdornment position="end">
                        <Button
                          type="button"
                          className="password-toggle"
                          onClick={() => setShowPassword((visible) => !visible)}
                          aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                        >
                          {showPassword ? 'Скрыть' : 'Показать'}
                        </Button>
                      </InputAdornment>
                    ),
                  },
                }}
              />
            </Stack>

            <Button type="submit" variant="contained" size="large" disabled={submitting} fullWidth>
              {submitting ? <CircularProgress size={22} color="inherit" aria-label="Выполняется вход" /> : 'Войти в кабинет'}
            </Button>

            <Typography className="login-help" color="text.secondary">
              Нет доступа или забыли пароль? Обратитесь к администратору проекта.
            </Typography>
          </Stack>
        </Box>

        <p className="login-footer">Tutorina · личный кабинет</p>
      </section>
    </main>
  )
}
