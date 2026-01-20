// src/config/config.ts

type AppConfig = {
  apiBaseUrl: string;
  frontendBaseUrl: string;
};

function requireEnv(name: string): string {
  const value = process.env[name];

  if (!value) {
    throw new Error(
      `[Config error] Missing required environment variable: ${name}`
    );
  }

  return value.replace(/\/+$/, ''); // remove trailing slash
}

const config: AppConfig = {
  apiBaseUrl: requireEnv('REACT_APP_API_URL'),
  frontendBaseUrl: requireEnv('REACT_APP_FRONTEND_URL'),
};

export default config;