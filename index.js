const express = require('express');
const path = require('path');
const fs = require('fs-extra');

const app = express();
const PORT = process.env.PORT || 3000;
app.use(express.json());
app.use(express.static('public'));

// Sirf pairing API - Bot code bilkul nahi
let pairingCode = '';
let isConnected = false;

app.get('/status', (req, res) => res.json({ 
    connected: false, 
    code: pairingCode,
    msg: 'Bot Vercel pe nahi chalta. Railway.app use karo'
}));

app.post('/pair', (req, res) => {
    res.json({ 
        error: 'WhatsApp bot Vercel serverless pe nahi chalta.\n1. Railway.app pe deploy karo\n2. Wahan terminal me pairing code milega' 
    });
});

app.get('/', (req, res) => res.sendFile(path.join(__dirname, 'public', 'index.html')));

app.listen(PORT, () => {
    console.log(`Web chal gaya: http://localhost:${PORT}`);
});
