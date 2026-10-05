import { Link, NavLink, Outlet } from 'react-router-dom'

export function AppShell() {
  return (
    <div className="app-shell">
      <header className="topbar">
        <Link to="/" className="brand" aria-label="返回媒体库">
          <span className="brand-mark">E</span>
          <span>
            <strong>English Media</strong>
            <small>Learner · Local</small>
          </span>
        </Link>
        <nav>
          <NavLink to="/" end>媒体库</NavLink>
          <span className="nav-disabled" title="将在后续阶段开放">生词本</span>
          <span className="nav-disabled" title="将在后续阶段开放">复习</span>
        </nav>
      </header>
      <main><Outlet /></main>
    </div>
  )
}

