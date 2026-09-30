import { createTheme } from '@mui/material'
import { colors } from './designTokens'

export const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: colors.violet,
      dark: colors.violetDark,
      light: colors.violetLight,
      contrastText: colors.white,
    },
    background: {
      default: colors.canvas,
      paper: colors.surface,
    },
    text: {
      primary: colors.plum,
      secondary: colors.muted,
    },
  },
  typography: {
    fontFamily: 'Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    h1: {
      fontWeight: 650,
      letterSpacing: '-0.045em',
    },
    h2: {
      fontWeight: 650,
      letterSpacing: '-0.035em',
    },
    button: {
      fontWeight: 650,
      textTransform: 'none',
    },
  },
  shape: {
    borderRadius: 14,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          minHeight: 48,
          boxShadow: 'none',
        },
        contained: {
          '&:hover': {
            boxShadow: '0 10px 24px rgba(98, 76, 136, 0.18)',
          },
        },
      },
    },
    MuiOutlinedInput: {
      styleOverrides: {
        root: {
          backgroundColor: colors.white,
          '& fieldset': {
            borderColor: colors.border,
          },
          '&:hover fieldset': {
            borderColor: colors.violet,
          },
        },
      },
    },
  },
})
