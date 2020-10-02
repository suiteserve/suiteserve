const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  app.use(
    '/v1',
    createProxyMiddleware({
      target: 'https://localhost:8080',
      secure: false,
      changeOrigin: true,
    }),
  );
};
