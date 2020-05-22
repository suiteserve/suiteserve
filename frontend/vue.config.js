module.exports = {
  devServer: {
    host: 'localhost',
    https: true,
    key: '../config/key.pem',
    cert: '../config/cert.pem',
    ca: '../config/ca.pem',
    proxy: 'https://localhost:8080',
  },
  integrity: true,
};
