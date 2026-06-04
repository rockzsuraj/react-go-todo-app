const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = (app) => {
  app.use(
    '/api',
    createProxyMiddleware({
      target: 'http://backend:8080',
      changeOrigin: false,
      timeout: 60000,
      proxyTimeout: 60000,
      cookieDomainRewrite: '',
      onProxyRes: (proxyRes) => {
        const cookies = proxyRes.headers['set-cookie'];
        if (cookies) {
          proxyRes.headers['set-cookie'] = cookies.map((cookie) =>
            cookie
              .replace(/;\s*secure/gi, '')
              .replace(/;\s*samesite=strict/gi, '; SameSite=Lax'),
          );
        }
      },
    }),
  );
};
