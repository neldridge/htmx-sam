const express = require('express');
const path = require('path');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();
const PORT = process.env.PORT || 3000;

// Serve static files (including index.html) from ./public
app.use(express.static(path.join(__dirname, '.')));

// Serve index.html on root "/"
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'index.html'));
});

// Proxy all other paths (e.g. /api, /dashboard) to localhost:3001
app.use(
  createProxyMiddleware({
    target: 'http://localhost:3001',
    changeOrigin: true,
     pathRewrite: {'^/api' : ''}
  })
);

app.listen(PORT, () => {
  console.log(`Frontend server running at http://localhost:${PORT}`);
});
