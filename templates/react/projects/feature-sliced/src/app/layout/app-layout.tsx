import { NavLink, Outlet } from 'react-router';
import { navItems } from './nav-items';

export function AppLayout() {
  return (
    <div className="shell">
      <header className="topbar">
        <span className="brand">Sample App</span>
        <nav aria-label="Principal">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} end>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </header>
      <Outlet />
    </div>
  );
}
