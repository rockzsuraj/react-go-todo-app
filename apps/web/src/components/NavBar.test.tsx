import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import NavBar from './NavBar';
import { useAuth, useLogout } from '../hooks/useAuth';

// Define mock functions outside
// Define mock functions outside

// Mock the hook with a factory
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
        // Default mock implementations
        mockUseLogout.mockReturnValue(mockLogoutFn);
    });

    test('renders loading spinner when loading', () => {
        mockUseAuth.mockReturnValue({
            data: null,
            isLoading: true,
        });

        render(
            <BrowserRouter>
                <NavBar />
            </BrowserRouter>
        );

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    test('renders Login button when user is not authenticated', () => {
        mockUseAuth.mockReturnValue({
            data: null,
            isLoading: false,
        });

        render(
            <BrowserRouter>
                <NavBar />
            </BrowserRouter>
        );

        expect(screen.getByText('Todo Manager')).toBeInTheDocument();
        expect(screen.getByText('Login')).toBeInTheDocument();
        expect(screen.queryByText('Logout')).not.toBeInTheDocument();
    });

    test('renders user info and Logout button when authenticated', () => {
        mockUseAuth.mockReturnValue({
            data: {
                user: {
                    name: 'Test User',
                    email: 'test@example.com',
                    picture: 'https://example.com/pic.jpg',
                },
            },
            isLoading: false,
        });

        render(
            <BrowserRouter>
                <NavBar />
            </BrowserRouter>
        );

        expect(screen.getByText('Test User')).toBeInTheDocument();
        expect(screen.getByText('test@example.com')).toBeInTheDocument();
        expect(screen.getByText('Logout')).toBeInTheDocument();
        expect(screen.queryByText('Login')).not.toBeInTheDocument();
    });

    test('calls logout when Logout button is clicked', () => {
        mockUseAuth.mockReturnValue({
            data: {
                user: {
                    name: 'Test User',
                    email: 'test@example.com',
                },
            },
            isLoading: false,
        });

        render(
            <BrowserRouter>
                <NavBar />
            </BrowserRouter>
        );

        fireEvent.click(screen.getByText('Logout'));
        expect(mockLogoutFn).toHaveBeenCalledTimes(1);
    });
});
