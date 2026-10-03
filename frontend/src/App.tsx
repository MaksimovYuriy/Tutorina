import { AdminPage } from './pages/AdminPage/AdminPage'
import { BoardPage } from './pages/BoardPage/BoardPage'
import { LoginPage } from './pages/LoginPage/LoginPage'
export function App() {
  const path = window.location.pathname
  if (path === '/admin' || path === '/admin/schedule') return <AdminPage />
  if (path === '/login') return <LoginPage />
  if (path !== '/')
    return (
      <main>
        <h1>Страница не найдена</h1>
        <a href="/">К расписанию</a>
      </main>
    )
  return <BoardPage />
}
