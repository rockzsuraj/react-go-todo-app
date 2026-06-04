import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth, useLogout } from '../hooks/useAuth';

const NavBar = React.memo(function NavBar() {
  const { data: user, isLoading } = useAuth();
  const logout = useLogout();

  return (
    <nav className="navbar navbar-expand-lg premium-navbar">
      <div className="container-fluid">
        <Link className="navbar-brand navbar-brand-premium" to="/">
          <i className="bi bi-check2-square"></i>
          Todo Manager
        </Link>

        <button
          className="navbar-toggler border-0 shadow-none"
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
          <div className="navbar-nav ms-auto align-items-center gap-2">
            {isLoading ? (
              <output
                className="spinner-border spinner-border-sm text-secondary"
                style={{ borderWidth: '2px' }}
              >
                <span className="visually-hidden">Loading...</span>
              </output>
            ) : user?.name ? (
              <>
                {/* Mobile: Show user info directly */}
                <div className="d-flex flex-column align-items-center gap-3 d-lg-none w-100 py-3 mt-2 rounded-4 bg-light bg-opacity-50 border border-light-subtle">
                  <img
                    src={
                      user.picture ||
                      `https://ui-avatars.com/api/?name=${encodeURIComponent(user.name)}`
                    }
                    alt={user.name}
                    width={48}
                    height={48}
                    className="rounded-circle user-avatar-premium flex-shrink-0"
                  />
                  <div className="text-center">
                    <div className="fw-semibold text-dark">{user.name}</div>
                    <div className="small text-muted">{user.email}</div>
                  </div>

                  <div className="d-flex gap-2 mt-2 w-100 justify-content-center px-3">
                    {user.role === 'admin' && (
                      <Link
                        className="btn btn-outline-primary btn-sm rounded-pill px-3"
                        to="/admin"
                      >
                        <i className="bi bi-gear me-1"></i>
                        Admin
                      </Link>
                    )}
                    <button
                      type="button"
                      className="btn btn-outline-danger btn-sm rounded-pill px-3"
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
                    className="btn user-dropdown-btn d-flex align-items-center gap-2 dropdown-toggle shadow-none"
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
                      width={28}
                      height={28}
                      className="rounded-circle user-avatar-premium flex-shrink-0"
                    />
                    <span className="d-none d-lg-inline fw-semibold text-secondary">
                      {user.name}
                    </span>
                  </button>

                  <ul className="dropdown-menu dropdown-menu-end dropdown-menu-premium border-0">
                    <li className="dropdown-user-card">
                      <div className="fw-semibold text-dark">{user.name}</div>
                      <div className="small text-muted text-break">
                        {user.email}
                      </div>
                    </li>

                    {user.role === 'admin' && (
                      <li>
                        <Link
                          className="dropdown-item dropdown-item-premium"
                          to="/admin"
                        >
                          <i className="bi bi-gear me-2"></i>
                          Admin Panel
                        </Link>
                      </li>
                    )}

                    {user.role === 'admin' && (
                      <li>
                        <hr className="dropdown-divider my-1 opacity-50" />
                      </li>
                    )}

                    <li>
                      <button
                        type="button"
                        className="dropdown-item dropdown-item-premium text-danger"
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
              <Link
                className="btn btn-premium-primary rounded-pill px-4"
                to="/login"
              >
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
