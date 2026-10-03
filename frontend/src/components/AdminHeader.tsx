import { Button, Stack } from '@mui/material'
import { BrandLink } from './BrandLink'
import { logout } from '../services/api'
export function AdminHeader({
  onError,
}: {
  onError: (error: unknown) => void
}) {
  return (
    <Stack
      direction={{ xs: 'column', sm: 'row' }}
      spacing={2}
      sx={{ justifyContent: 'space-between' }}
    >
      <BrandLink />
      <Stack direction="row" spacing={1} useFlexGap sx={{ flexWrap: 'wrap' }}>
        <Button href="/admin">Расписание</Button>
        <Button href="/admin/directions">Направления</Button>
        <Button
          onClick={async () => {
            try {
              await logout()
              window.location.assign('/login')
            } catch (error) {
              onError(error)
            }
          }}
        >
          Выйти
        </Button>
      </Stack>
    </Stack>
  )
}
