import { AdminHome } from './pages/AdminHome/AdminHome'
import { AdminOffersPage } from './pages/AdminOffersPage/AdminOffersPage'
import { ApplicationsPage } from './pages/ApplicationsPage/ApplicationsPage'
import { LoginPage } from './pages/LoginPage/LoginPage'
import { SchedulePage } from './pages/SchedulePage/SchedulePage'
import { TeachersPage } from './pages/TeachersPage/TeachersPage'
import { TeacherHome } from './pages/TeacherHome/TeacherHome'

export function App() {
  if (window.location.pathname.startsWith('/admin/applications') || window.location.pathname.startsWith('/teacher/applications')) {
    return <ApplicationsPage />
  }

  if (window.location.pathname.startsWith('/admin/schedule') || window.location.pathname.startsWith('/teacher/schedule')) {
    return <SchedulePage />
  }

  if (window.location.pathname.startsWith('/admin/offers')) {
    return <AdminOffersPage />
  }

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
