module.exports = {
  assetsDir: 'static/',
  devServer: {
    https: true,
    key: '../tls/key.pem',
    cert: '../tls/cert.pem',
    ca: '../tls/ca.pem',
    proxy: 'https://localhost:8080',
  },
  integrity: true,
};
