import { AdminHome } from './pages/AdminHome/AdminHome'
import { LoginPage } from './pages/LoginPage/LoginPage'
import { TeachersPage } from './pages/TeachersPage/TeachersPage'
import { TeacherHome } from './pages/TeacherHome/TeacherHome'

export function App() {
  if (window.location.pathname.startsWith('/admin')) {
    return <AdminHome />
  }

  if (window.location.pathname.startsWith('/teacher')) {
    return <TeacherHome />
  }

  if (window.location.pathname.startsWith('/login')) {
    return <LoginPage />
  }

  return <TeachersPage />
}
