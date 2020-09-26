export default {
  port: 8081,
  https: true,
  httpsOptions: {
    ca: '../tls/ca.pem',
    cert: '../tls/cert.pem',
    key: '../tls/key.pem',
  },
  proxy: {
    '/v1/': {
      target: 'https://localhost:8080',
      secure: false,
      changeOrigin: true,
    },
  },
};
