module.exports = {
  devServer: {
    host: 'localhost',
    proxy: 'https://localhost:8080',
    https: true,
    key: '../config/key.pem',
    cert: '../config/cert.pem',
    ca: '../config/ca.pem',
  },
  integrity: true,
};
