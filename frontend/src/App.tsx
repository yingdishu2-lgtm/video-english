import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from './components/AppShell'
import { LibraryPage } from './features/library/LibraryPage'
import { PlayerPage } from './features/player/PlayerPage'

export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route path="/" element={<LibraryPage />} />
        <Route path="/watch/:mediaId" element={<PlayerPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}

