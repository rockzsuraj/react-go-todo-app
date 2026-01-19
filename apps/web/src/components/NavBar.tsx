import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth, useLogout } from '../hooks/useAuth';

const NavBar = React.memo(function NavBar() {
  const { data, isLoading } = useAuth();
  const logout = useLogout();
  const user = data?.user ?? null;

  return (
    <nav className="navbar navbar-expand-lg navbar-light bg-light">
      <div className="container-fluid">
        <Link className="navbar-brand" to="/">
          Todo Manager
        </Link>

        <div className="d-flex align-items-center">
          {isLoading ? (
            <output className="spinner-border spinner-border-sm text-secondary">
              <span className="visually-hidden">Loading...</span>
            </output>
          ) : user ? (
            <div className="d-flex align-items-center gap-3">
              <img
                src={user.picture || `https://ui-avatars.com/api/?name=${user.name}`}
                alt={user.name}
                width={36}
                height={36}
                className="rounded-circle border"
              />

              <div className="me-3 text-end">
                <div className="fw-semibold">{user.name}</div>
                <small className="text-muted">{user.email}</small>
              </div>

              <button type="button" className="btn btn-outline-danger btn-sm" onClick={logout}>
                Logout
              </button>
            </div>
          ) : (
            <Link className="btn btn-danger btn-sm" to="/login">
              Login
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
});

export default NavBar;
