const express = require('express');
const proxy = require('express-http-proxy');

const app = express();

app.use('/', proxy('http://localhost:3001'));

app.listen(3000, () => {
  console.log('Frontend proxy listening on http://localhost:3000');
});
