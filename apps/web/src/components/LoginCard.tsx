import config from '../config/config';

export default function LoginCard() {
    const loginUrl = new URL(
        '/api/auth/google/login',
        config.apiBaseUrl
    );

    loginUrl.searchParams.set(
        'redirect',
        `${config.frontendBaseUrl}/oauth/callback`
    );

    return (
        <div className="container mt-5 text-center">
            <div className="card p-4">
                <h4 className="mb-3">Welcome to Todos Manager 👋</h4>
                <p className="text-muted">Please login with Google to create and manage your todos.</p>
                <a className="btn btn-danger mt-2" href={loginUrl.toString()}>
                    Login with Google
                </a>
            </div>
        </div>
    );
}
