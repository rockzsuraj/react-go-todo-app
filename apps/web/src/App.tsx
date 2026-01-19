import React, { Suspense } from 'react';
import './App.css';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import ErrorBoundary from './components/ErrorBoundary';
import LoadingSkeleton from './components/LoadingSkeleton';

// Lazy load pages
const Home = React.lazy(() => import('./pages/Home'));
const Login = React.lazy(() => import('./pages/Login'));
const AuthCallback = React.lazy(() => import('./pages/AuthCallback'));

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={
          <ErrorBoundary>
            <Layout />
          </ErrorBoundary>
        }>
          <Route path="/" element={
            <Suspense fallback={<LoadingSkeleton />}>
              <Home />
            </Suspense>
          } />
          <Route path="/login" element={
            <Suspense fallback={<LoadingSkeleton />}>
              <Login />
            </Suspense>
          } />
          <Route path="/oauth/callback" element={
            <Suspense fallback={<LoadingSkeleton />}>
              <AuthCallback />
            </Suspense>
          } />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;