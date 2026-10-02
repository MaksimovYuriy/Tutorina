import { Box, Typography } from '@mui/material'

export function BrandLink() {
  return (
    <Box
      component="a"
      href="/"
      aria-label="Tutorina — на главную"
      sx={{
        width: 'fit-content',
        display: 'inline-flex',
        alignItems: 'center',
        gap: 1.25,
        color: 'primary.main',
        textDecoration: 'none',
        '&:hover': { color: 'primary.dark' },
        '&:focus-visible': { outline: '2px solid', outlineColor: 'primary.main', outlineOffset: 4, borderRadius: 1 },
      }}
    >
      <Box
        component="span"
        aria-hidden="true"
        sx={{
          width: 34,
          height: 34,
          display: 'grid',
          placeItems: 'center',
          borderRadius: '11px 11px 11px 4px',
          bgcolor: 'primary.main',
          color: 'primary.contrastText',
          fontFamily: 'Georgia, serif',
          fontSize: 21,
          boxShadow: '0 8px 22px rgba(94, 73, 111, 0.18)',
        }}
      >
        T
      </Box>
      <Typography component="span" sx={{ color: 'inherit', fontWeight: 750, fontSize: 20, letterSpacing: '-0.02em' }}>
        Tutorina
      </Typography>
    </Box>
  )
}