import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth, useLogout } from '../hooks/useAuth';

const NavBar = React.memo(function NavBar() {
  const { data: user, isLoading } = useAuth();
  const logout = useLogout();
  
  // Extract user from the response. 
  // If your Go backend returns { "user": { "Name": "..." } }, use data?.user
  // If it returns the user object directly, use data.

  console.log('User in NavBar:', user);

  return (
    <nav className="navbar navbar-expand-lg navbar-light bg-light border-bottom">
      <div className="container-fluid">
        <Link className="navbar-brand fw-semibold" to="/">
          Todo Manager
        </Link>

        <button
          className="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#navbarContent"
          aria-controls="navbarContent"
          aria-expanded="false"
          aria-label="Toggle navigation"
        >
          <span className="navbar-toggler-icon" />
        </button>

        <div className="collapse navbar-collapse" id="navbarContent">
          <div className="navbar-nav ms-auto">
            {isLoading ? (
              <div className="spinner-border spinner-border-sm text-secondary" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            ) : user?.name ? (
              <>
                {/* Mobile: Show user info directly */}
                <div className="d-flex flex-column align-items-center gap-3 d-lg-none">
                  <img
                    src={
                      user.picture || 
                      `https://ui-avatars.com/api/?name=${encodeURIComponent(user.name)}`
                    }
                    alt={user.name}
                    width={32}
                    height={32}
                    className="rounded-circle border flex-shrink-0"
                  />
                  <div className="text-center">
                    <div className="fw-semibold">{user.name}</div>
                    <div className="small text-muted">{user.email}</div>
                  </div>
                  
                  <div className="d-flex flex-column gap-2 mt-3">
                    {user.role === 'admin' && (
                      <Link className="btn btn-outline-primary btn-sm" to="/admin">
                        <i className="bi bi-gear me-1"></i>
                        Admin
                      </Link>
                    )}
                    <button
                      className="btn btn-outline-danger btn-sm"
                      onClick={() => logout.mutate()}
                    >
                      <i className="bi bi-box-arrow-right me-1"></i>
                      Logout
                    </button>
                  </div>
                </div>

                {/* Desktop: Show dropdown */}
                <div className="dropdown d-none d-lg-block">
                  <button
                    className="btn btn-light d-flex align-items-center gap-2 dropdown-toggle"
                    type="button"
                    data-bs-toggle="dropdown"
                    aria-expanded="false"
                  >
                    <img
                      src={
                        user.picture || 
                        `https://ui-avatars.com/api/?name=${encodeURIComponent(user.name)}`
                      }
                      alt={user.name}
                      width={32}
                      height={32}
                      className="rounded-circle border flex-shrink-0"
                    />
                    <span className="d-none d-lg-inline fw-semibold">
                      {user.name}
                    </span>
                  </button>

                  <ul className="dropdown-menu dropdown-menu-end">
                    <li className="px-3 py-2">
                      <div className="fw-semibold">{user.name}</div>
                      <div className="small text-muted text-break">{user.email}</div>
                    </li>

                    <li>
                      <hr className="dropdown-divider" />
                    </li>

                    {user.role === 'admin' && (
                      <li>
                        <Link className="dropdown-item" to="/admin">
                          <i className="bi bi-gear me-2"></i>
                          Admin Panel
                        </Link>
                      </li>
                    )}

                    {user.role === 'admin' && (
                      <li>
                        <hr className="dropdown-divider" />
                      </li>
                    )}

                    <li>
                      <button
                        className="dropdown-item text-danger"
                        onClick={() => logout.mutate()}
                      >
                        <i className="bi bi-box-arrow-right me-2"></i>
                        Logout
                      </button>
                    </li>
                  </ul>
                </div>
              </>
            ) : (
              <Link className="btn btn-primary" to="/login">
                <i className="bi bi-box-arrow-in-right me-2"></i>
                Login
              </Link>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
});

export default NavBar;