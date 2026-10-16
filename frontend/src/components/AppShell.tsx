import { Link, NavLink, Outlet, useMatch } from 'react-router-dom'
import { Icon } from './Icon'

export function AppShell() {
  const isPlayer = useMatch('/watch/:mediaId')
  return (
    <div className={`app-shell${isPlayer ? ' player-shell' : ''}`}>
      <header className="topbar">
        <Link to="/" className="brand" aria-label="返回媒体库">
          <span className="brand-mark"><Icon name="video" size={24} /></span>
          <span>
            <strong>English Media</strong>
            <small>English Media Learner</small>
          </span>
        </Link>
        <nav aria-label="主导航">
          <NavLink to="/" end><Icon name="library" /><span>媒体库</span></NavLink>
        </nav>
        <div className="nav-footer"><Icon name="folder" size={18} /><span>本地媒体库</span></div>
      </header>
      <main><Outlet /></main>
    </div>
  )
}

