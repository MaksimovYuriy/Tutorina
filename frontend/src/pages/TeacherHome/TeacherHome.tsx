import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'
import { Alert, Avatar, Box, Button, Checkbox, CircularProgress, Container, FormControlLabel, Stack, TextField, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import { ApiError, changePassword, getCurrentUser, getMyTeacherProfile, logout, removeMyTeacherPhoto, updateMyTeacherProfile, uploadMyTeacherPhoto, type CurrentUser, type TeacherProfile, type TeacherProfileInput } from '../../services/api'

export function TeacherHome() {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [teacher, setTeacher] = useState<TeacherProfile | null>(null)
  const [profile, setProfile] = useState<TeacherProfileInput | null>(null)
  const [photo, setPhoto] = useState<File | null>(null)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([getCurrentUser(controller.signal), getMyTeacherProfile(controller.signal)])
      .then(([currentUser, currentProfile]) => {
        if (!currentUser.roles.includes('teacher')) {
          window.location.replace('/admin')
          return
        }
        setUser(currentUser)
        setTeacher(currentProfile)
        setProfile(toInput(currentProfile))
      })
      .catch((caught: unknown) => {
        if (caught instanceof DOMException && caught.name === 'AbortError') return
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.replace('/login')
          return
        }
        setError('Не удалось загрузить кабинет преподавателя.')
      })
    return () => controller.abort()
  }, [])

  async function saveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!profile) return
    setSaving(true)
    setError(null)
    try {
      let updated = await updateMyTeacherProfile(profile)
      if (photo) updated = await uploadMyTeacherPhoto(photo)
      setTeacher(updated)
      setProfile(toInput(updated))
      setPhoto(null)
    } catch (caught) {
      setError(caught instanceof ApiError ? caught.message : 'Не удалось сохранить профиль.')
    } finally {
      setSaving(false)
    }
  }

  async function deletePhoto() {
    setError(null)
    try {
      const updated = await removeMyTeacherPhoto()
      setTeacher(updated)
    } catch {
      setError('Не удалось удалить фотографию.')
    }
  }

  async function savePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSaving(true)
    setError(null)
    try {
      await changePassword(currentPassword, newPassword)
      window.alert('Пароль изменён. Войдите ещё раз с новым паролем.')
      window.location.replace('/login')
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 401) setError('Текущий пароль указан неверно.')
      else setError(caught instanceof ApiError ? caught.message : 'Не удалось изменить пароль.')
      setSaving(false)
    }
  }

  async function handleLogout() {
    await logout()
    window.location.assign('/login')
  }

  function selectPhoto(event: ChangeEvent<HTMLInputElement>) {
    setPhoto(event.target.files?.[0] ?? null)
  }

  if (!user || !teacher || !profile) return <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}>{error ? <Alert severity="error">{error}</Alert> : <CircularProgress />}</Box>

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 3, md: 6 } }}>
      <Container maxWidth="lg">
        <Stack spacing={4}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between', alignItems: { sm: 'center' } }}>
            <div>
              <BrandLink />
              <Typography component="h1" variant="h3">Кабинет преподавателя</Typography>
              <Typography color="text.secondary">{user.email}</Typography>
            </div>
            <Stack direction="row" spacing={1}>
              <Button href="/teacher/schedule" variant="outlined">Расписание</Button>
              <Button variant="outlined" onClick={handleLogout}>Выйти</Button>
            </Stack>
          </Stack>

          {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}

          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'minmax(0, 1fr) 360px' }, gap: 3, alignItems: 'start' }}>
            <Box component="form" onSubmit={saveProfile} sx={{ p: { xs: 3, md: 4 }, bgcolor: 'background.paper', borderRadius: 3 }}>
              <Stack spacing={2}>
                <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                  <Avatar src={teacher.photoUrl || undefined} alt={teacher.displayName} sx={{ width: 88, height: 88, bgcolor: 'primary.light', color: 'primary.dark', fontSize: 30 }}>
                    {teacher.displayName.slice(0, 1).toUpperCase()}
                  </Avatar>
                  <Stack spacing={1}>
                    <Button component="label" variant="outlined">
                      {photo ? 'Выбрать другое фото' : 'Заменить фото'}
                      <input hidden type="file" accept="image/jpeg,image/png,image/webp" onChange={selectPhoto} />
                    </Button>
                    {teacher.photoUrl && <Button size="small" color="error" onClick={deletePhoto}>Удалить фото</Button>}
                    {photo && <Typography variant="body2">{photo.name}</Typography>}
                  </Stack>
                </Stack>
                <TextField label="Имя" required value={profile.displayName} onChange={(event) => setProfile({ ...profile, displayName: event.target.value })} />
                <TextField label="Образование" value={profile.education} onChange={(event) => setProfile({ ...profile, education: event.target.value })} />
                <TextField label="Опыт" value={profile.experience} onChange={(event) => setProfile({ ...profile, experience: event.target.value })} />
                <TextField label="Подход к занятиям" multiline minRows={4} value={profile.approach} onChange={(event) => setProfile({ ...profile, approach: event.target.value })} />
                <FormControlLabel control={<Checkbox checked={profile.isPublished} onChange={(event) => setProfile({ ...profile, isPublished: event.target.checked })} />} label="Показывать профиль на публичной странице" />
                <Button type="submit" variant="contained" disabled={saving}>{saving ? 'Сохраняем…' : 'Сохранить профиль'}</Button>
              </Stack>
            </Box>

            <Stack spacing={3}>
              <Box component="form" onSubmit={savePassword} sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3 }}>
                <Stack spacing={2}>
                  <Typography variant="h5">Сменить пароль</Typography>
                  <TextField label="Текущий пароль" type="password" required value={currentPassword} onChange={(event) => setCurrentPassword(event.target.value)} />
                  <TextField label="Новый пароль" type="password" required helperText="Не менее 12 символов" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} />
                  <Button type="submit" variant="outlined" disabled={saving}>Изменить пароль</Button>
                </Stack>
              </Box>
              <Box sx={{ p: 3, bgcolor: 'primary.main', color: 'primary.contrastText', borderRadius: 3 }}>
                <Typography variant="h5">Расписание</Typography>
                <Typography sx={{ mt: 1, opacity: 0.85 }}>Создавайте занятия и управляйте открытой записью.</Typography>
                <Button href="/teacher/schedule" variant="contained" sx={{ mt: 2, bgcolor: 'background.paper', color: 'primary.dark' }}>Открыть расписание</Button>
              </Box>
            </Stack>
          </Box>
        </Stack>
      </Container>
    </Box>
  )
}

function toInput(profile: TeacherProfile): TeacherProfileInput {
  return {
    displayName: profile.displayName,
    education: profile.education,
    experience: profile.experience,
    approach: profile.approach,
    isPublished: profile.isPublished,
  }
}