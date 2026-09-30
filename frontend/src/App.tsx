import { AdminHome } from './pages/AdminHome/AdminHome'
import { LoginPage } from './pages/LoginPage/LoginPage'

export function App() {
  if (window.location.pathname.startsWith('/admin')) {
    return <AdminHome />
  }

  return <LoginPage />
}
