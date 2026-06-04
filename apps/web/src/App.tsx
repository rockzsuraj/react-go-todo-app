import React, { Suspense } from 'react';
import './App.css';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import ErrorBoundary from './components/ErrorBoundary';
import GuestRoute from './components/GuestRoute';
import Layout from './components/Layout';
import LoadingSkeleton from './components/LoadingSkeleton';
import ProtectedRoute from './components/ProtectedRoute';

// Lazy load pages
const Home = React.lazy(() => import('./pages/Home'));
const Login = React.lazy(() => import('./pages/Login'));
const AuthCallback = React.lazy(() => import('./pages/AuthCallback'));
const Admin = React.lazy(() => import('./pages/Admin'));

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          element={
            <ErrorBoundary>
              <Layout />
            </ErrorBoundary>
          }
        >
          {/* Protected: requires authentication */}
          <Route element={<ProtectedRoute />}>
            <Route
              path="/"
              element={
                <Suspense fallback={<LoadingSkeleton />}>
                  <Home />
                </Suspense>
              }
            />
            <Route
              path="/admin"
              element={
                <Suspense fallback={<LoadingSkeleton />}>
                  <Admin />
                </Suspense>
              }
            />
          </Route>

          {/* Guest only: redirects logged-in users away */}
          <Route element={<GuestRoute />}>
            <Route
              path="/login"
              element={
                <Suspense fallback={<LoadingSkeleton />}>
                  <Login />
                </Suspense>
              }
            />
          </Route>

          {/* Public: handles its own auth flow */}
          <Route
            path="/oauth/callback"
            element={
              <Suspense fallback={<LoadingSkeleton />}>
                <AuthCallback />
              </Suspense>
            }
          />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
