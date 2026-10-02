import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'
import { Alert, Avatar, Box, Button, Checkbox, CircularProgress, Container, FormControlLabel, Stack, TextField, Typography } from '@mui/material'
import { BrandLink } from '../../components/BrandLink'
import {
  ApiError,
  archiveTeacherProfile,
  createTeacherAccount,
  createTeacherProfile,
  getCurrentUser,
  getTeacherProfiles,
  logout,
  removeTeacherPhoto,
  resetTeacherPassword,
  updateTeacherProfile,
  uploadTeacherPhoto,
  type CurrentUser,
  type TeacherProfile,
  type TeacherProfileInput,
} from '../../services/api'

const emptyProfile: TeacherProfileInput = {
  displayName: '',
  education: '',
  experience: '',
  approach: '',
  isPublished: false,
}

export function AdminHome() {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [teachers, setTeachers] = useState<TeacherProfile[]>([])
  const [profile, setProfile] = useState<TeacherProfileInput>(emptyProfile)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [photo, setPhoto] = useState<File | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [accountProfileId, setAccountProfileId] = useState<string | null>(null)
  const [accountEmail, setAccountEmail] = useState('')
  const [accountPassword, setAccountPassword] = useState('')
  const [passwordProfileId, setPasswordProfileId] = useState<string | null>(null)
  const [temporaryPassword, setTemporaryPassword] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([getCurrentUser(controller.signal), getTeacherProfiles(controller.signal)])
      .then(([currentUser, profiles]) => {
        setUser(currentUser)
        setTeachers(profiles)
      })
      .catch((caught: unknown) => {
        if (caught instanceof DOMException && caught.name === 'AbortError') return
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.replace('/login')
          return
        }
        setError(caught instanceof ApiError && caught.status === 403 ? 'Для этого раздела нужна роль администратора.' : 'Не удалось загрузить кабинет.')
      })
    return () => controller.abort()
  }, [])

  function resetForm() {
    setProfile(emptyProfile)
    setEditingId(null)
    setPhoto(null)
  }

  function startEditing(teacher: TeacherProfile) {
    setEditingId(teacher.id)
    setProfile({
      displayName: teacher.displayName,
      education: teacher.education,
      experience: teacher.experience,
      approach: teacher.approach,
      isPublished: teacher.isPublished,
    })
    setPhoto(null)
    setError(null)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const baseSaved = editingId
        ? await updateTeacherProfile(editingId, profile)
        : await createTeacherProfile(profile)

      setTeachers((items) => {
        const exists = items.some((item) => item.id === baseSaved.id)
        return (exists ? items.map((item) => item.id === baseSaved.id ? baseSaved : item) : [...items, baseSaved])
          .sort((left, right) => left.displayName.localeCompare(right.displayName, 'ru'))
      })

      if (photo) {
        try {
          const updated = await uploadTeacherPhoto(baseSaved.id, photo)
          setTeachers((items) => items.map((item) => item.id === updated.id ? updated : item))
        } catch (caught) {
          setEditingId(baseSaved.id)
          setError(caught instanceof ApiError ? 'Профиль сохранён, но фото не загружено: ' + caught.message : 'Профиль сохранён, но фото не загружено.')
          return
        }
      }
      resetForm()
    } catch (caught) {
      setError(caught instanceof ApiError ? caught.message : 'Не удалось сохранить профиль.')
    } finally {
      setSaving(false)
    }
  }

  function handlePhotoSelection(event: ChangeEvent<HTMLInputElement>) {
    setPhoto(event.target.files?.[0] ?? null)
  }

  async function handleRemovePhoto(teacher: TeacherProfile) {
    setError(null)
    try {
      const updated = await removeTeacherPhoto(teacher.id)
      setTeachers((items) => items.map((item) => item.id === updated.id ? updated : item))
    } catch {
      setError('Не удалось удалить фотографию.')
    }
  }

  async function handleLogout() {
    await logout()
    window.location.assign('/login')
  }

  async function handleArchive(teacher: TeacherProfile) {
    if (!window.confirm(`Архивировать профиль «${teacher.displayName}»? Доступ в кабинет будет отключён, фотография удалена.`)) return
    setError(null)
    try {
      await archiveTeacherProfile(teacher.id)
      setTeachers((items) => items.filter((item) => item.id !== teacher.id))
      if (editingId === teacher.id) resetForm()
    } catch {
      setError('Не удалось архивировать преподавателя.')
    }
  }

  async function handleAccountSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!accountProfileId) return
    setSaving(true)
    setError(null)
    try {
      const updated = await createTeacherAccount(accountProfileId, accountEmail, accountPassword)
      setTeachers((items) => items.map((item) => item.id === updated.id ? updated : item))
      setAccountProfileId(null)
      setAccountEmail('')
      setAccountPassword('')
    } catch (caught) {
      setError(caught instanceof ApiError && caught.status === 409 ? 'Этот email уже используется или доступ уже выдан.' : 'Не удалось создать аккаунт.')
    } finally {
      setSaving(false)
    }
  }

  async function handlePasswordReset(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!passwordProfileId) return
    setSaving(true)
    setError(null)
    try {
      await resetTeacherPassword(passwordProfileId, temporaryPassword)
      setPasswordProfileId(null)
      setTemporaryPassword('')
    } catch (caught) {
      setError(caught instanceof ApiError ? caught.message : 'Не удалось сбросить пароль.')
    } finally {
      setSaving(false)
    }
  }
  if (!user && !error) return <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center' }}><CircularProgress /></Box>

  return (
    <Box component="main" sx={{ minHeight: '100vh', py: { xs: 3, md: 6 } }}>
      <Container maxWidth="lg">
        <Stack spacing={4}>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} sx={{ justifyContent: 'space-between' }}>
            <div>
              <BrandLink />
              <Typography component="h1" variant="h3">Преподаватели</Typography>
            </div>
            <Button variant="outlined" onClick={handleLogout}>Выйти</Button>
          </Stack>

          {error && <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>}

          {user?.roles.includes('admin') && (
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: 'minmax(0, 1fr) 420px' }, gap: 3, alignItems: 'start' }}>
              <Stack spacing={2}>
                {teachers.length === 0 && <Alert severity="info">Профилей пока нет. Создайте первого преподавателя.</Alert>}
                {teachers.map((teacher) => (
                  <Box key={teacher.id} sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3, border: '1px solid', borderColor: editingId === teacher.id ? 'primary.main' : 'divider' }}>
                    <Stack direction="row" sx={{ justifyContent: 'space-between', gap: 2 }}>
                      <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
                        <Avatar src={teacher.photoUrl || undefined} alt={teacher.displayName} sx={{ width: 64, height: 64, bgcolor: 'primary.light', color: 'primary.dark' }}>
                          {teacher.displayName.slice(0, 1).toUpperCase()}
                        </Avatar>
                        <div>
                          <Typography variant="h6">{teacher.displayName}</Typography>
                          <Typography color="text.secondary">{teacher.education || 'Образование не указано'}</Typography>
                        </div>
                      </Stack>
                      <Typography color={teacher.isPublished ? 'primary' : 'text.secondary'}>{teacher.isPublished ? 'Опубликован' : 'Черновик'}</Typography>
                    </Stack>

                    {teacher.experience && <Typography sx={{ mt: 2 }}>{teacher.experience}</Typography>}

                    <Box sx={{ mt: 2 }}>
                      {teacher.userId ? (
                        passwordProfileId === teacher.id ? (
                          <Box component="form" onSubmit={handlePasswordReset}>
                            <Stack spacing={1.5}>
                              <Typography variant="body2" color="text.secondary">Новый временный пароль</Typography>
                              <TextField type="password" required helperText="Не менее 12 символов. Все сессии будут завершены." value={temporaryPassword} onChange={(event) => setTemporaryPassword(event.target.value)} />
                              <Stack direction="row" spacing={1}>
                                <Button type="submit" size="small" variant="contained" disabled={saving}>Сбросить пароль</Button>
                                <Button size="small" onClick={() => setPasswordProfileId(null)}>Отмена</Button>
                              </Stack>
                            </Stack>
                          </Box>
                        ) : (
                          <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
                            <Typography variant="body2" color="text.secondary">Доступ в кабинет выдан</Typography>
                            <Button size="small" onClick={() => { setPasswordProfileId(teacher.id); setTemporaryPassword('') }}>Сбросить пароль</Button>
                          </Stack>
                        )
                      ) : accountProfileId === teacher.id ? (
                        <Box component="form" onSubmit={handleAccountSubmit} sx={{ mt: 2 }}>
                          <Stack spacing={2}>
                            <TextField label="Email для входа" type="email" required value={accountEmail} onChange={(event) => setAccountEmail(event.target.value)} />
                            <TextField label="Временный пароль" type="password" required helperText="Не менее 12 символов" value={accountPassword} onChange={(event) => setAccountPassword(event.target.value)} />
                            <Stack direction="row" spacing={1}>
                              <Button type="submit" variant="contained" disabled={saving}>Создать аккаунт</Button>
                              <Button type="button" onClick={() => setAccountProfileId(null)}>Отмена</Button>
                            </Stack>
                          </Stack>
                        </Box>
                      ) : (
                        <Button size="small" onClick={() => setAccountProfileId(teacher.id)}>Выдать доступ</Button>
                      )}
                    </Box>

                    <Stack direction="row" spacing={1} sx={{ mt: 2, flexWrap: 'wrap' }}>
                      <Button size="small" onClick={() => startEditing(teacher)}>Редактировать</Button>
                      {teacher.photoUrl && <Button size="small" onClick={() => handleRemovePhoto(teacher)}>Удалить фото</Button>}
                      <Button color="error" size="small" onClick={() => handleArchive(teacher)}>Архивировать</Button>
                    </Stack>
                  </Box>
                ))}
              </Stack>

              <Box component="form" onSubmit={handleSubmit} sx={{ p: 3, bgcolor: 'background.paper', borderRadius: 3, position: { md: 'sticky' }, top: 24 }}>
                <Stack spacing={2}>
                  <Typography variant="h5">{editingId ? 'Редактирование' : 'Новый преподаватель'}</Typography>
                  <TextField label="Имя" required value={profile.displayName} onChange={(event) => setProfile({ ...profile, displayName: event.target.value })} />
                  <TextField label="Образование" value={profile.education} onChange={(event) => setProfile({ ...profile, education: event.target.value })} />
                  <TextField label="Опыт" value={profile.experience} onChange={(event) => setProfile({ ...profile, experience: event.target.value })} />
                  <TextField label="Подход к занятиям" multiline minRows={3} value={profile.approach} onChange={(event) => setProfile({ ...profile, approach: event.target.value })} />

                  <Box sx={{ p: 2, border: '1px dashed', borderColor: 'divider', borderRadius: 2, bgcolor: 'background.default' }}>
                    <Stack spacing={1}>
                      <Typography sx={{ fontWeight: 650 }}>Фотография</Typography>
                      <Typography variant="body2" color="text.secondary">JPEG, PNG или WebP, не более 5 МБ. Новое фото заменит старое.</Typography>
                      <Button component="label" variant="outlined">
                        {photo ? 'Выбрать другой файл' : 'Выбрать файл'}
                        <input hidden type="file" accept="image/jpeg,image/png,image/webp" onChange={handlePhotoSelection} />
                      </Button>
                      {photo && <Typography variant="body2">{photo.name}</Typography>}
                    </Stack>
                  </Box>

                  <FormControlLabel control={<Checkbox checked={profile.isPublished} onChange={(event) => setProfile({ ...profile, isPublished: event.target.checked })} />} label="Показывать на публичной странице" />
                  <Stack direction="row" spacing={1}>
                    <Button type="submit" variant="contained" disabled={saving}>{saving ? 'Сохраняем…' : editingId ? 'Сохранить' : 'Создать профиль'}</Button>
                    {editingId && <Button type="button" onClick={resetForm} disabled={saving}>Отмена</Button>}
                  </Stack>
                </Stack>
              </Box>
            </Box>
          )}
        </Stack>
      </Container>
    </Box>
  )
}
