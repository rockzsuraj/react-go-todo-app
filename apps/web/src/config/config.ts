// src/config/config.ts

type AppConfig = {
  apiBaseUrl: string;
  frontendBaseUrl: string;
};

function getEnv(name: string): string {
  const value = process.env[name];

  if (!value) {
    if (process.env.NODE_ENV === 'production') {
      throw new Error(
        `[Config error] Missing required environment variable: ${name}`
      );
    }

    // Dev / local fallback (prevents blank screen)
    console.warn(`[Config warning] ${name} is not defined`);
    return '';
  }

  return value.replace(/\/+$/, '');
}

const config: AppConfig = {
  apiBaseUrl: getEnv('REACT_APP_API_URL'),
  frontendBaseUrl: getEnv('REACT_APP_FRONTEND_URL'),
};

export default config;