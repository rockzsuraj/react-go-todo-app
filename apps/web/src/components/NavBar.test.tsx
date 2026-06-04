import { fireEvent, render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { useAuth, useLogout } from '../hooks/useAuth';
import NavBar from './NavBar';

jest.mock('../hooks/useAuth', () => ({
  useAuth: jest.fn(),
  useLogout: jest.fn(),
}));

const mockUseAuth = useAuth as jest.Mock;
const mockUseLogout = useLogout as jest.Mock;

describe('NavBar Component', () => {
  const mockLogoutFn = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseLogout.mockReturnValue({ mutate: mockLogoutFn });
  });

  test('renders loading spinner when loading', () => {
    mockUseAuth.mockReturnValue({ data: null, isLoading: true });

    render(
      <BrowserRouter>
        <NavBar />
      </BrowserRouter>,
    );

    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  test('renders Login button when user is not authenticated', () => {
    mockUseAuth.mockReturnValue({ data: null, isLoading: false });

    render(
      <BrowserRouter>
        <NavBar />
      </BrowserRouter>,
    );

    expect(screen.getByText('Todo Manager')).toBeInTheDocument();
    expect(screen.getByText('Login')).toBeInTheDocument();
    expect(screen.queryByText('Logout')).not.toBeInTheDocument();
  });

  test('renders user info and Logout button when authenticated', () => {
    mockUseAuth.mockReturnValue({
      data: {
        name: 'Test User',
        email: 'test@example.com',
        picture: 'https://example.com/pic.jpg',
      },
      isLoading: false,
    });

    render(
      <BrowserRouter>
        <NavBar />
      </BrowserRouter>,
    );

    expect(screen.getAllByText('Test User')[0]).toBeInTheDocument();
    expect(screen.getAllByText('test@example.com')[0]).toBeInTheDocument();
    expect(screen.getAllByText('Logout')[0]).toBeInTheDocument();
    expect(screen.queryByText('Login')).not.toBeInTheDocument();
  });

  test('calls logout when Logout button is clicked', () => {
    mockUseAuth.mockReturnValue({
      data: { name: 'Test User', email: 'test@example.com' },
      isLoading: false,
    });

    render(
      <BrowserRouter>
        <NavBar />
      </BrowserRouter>,
    );

    fireEvent.click(screen.getAllByText('Logout')[0]);
    expect(mockLogoutFn).toHaveBeenCalledTimes(1);
  });
});
