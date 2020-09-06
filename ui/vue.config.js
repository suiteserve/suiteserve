module.exports = {
  devServer: {
    host: 'localhost',
    port: 8081,
    proxy: 'https://localhost:8080',
    https: true,
    key: '../tls/key.pem',
    cert: '../tls/cert.pem',
    ca: '../tls/ca.pem',
  },
  integrity: true,
};
