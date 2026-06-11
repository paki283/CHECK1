const { default: makeWASocket, useMultiFileAuthState, DisconnectReason, fetchLatestBaileysVersion, downloadContentFromMessage } = require('@whiskeysockets/baileys');
const express = require('express');
const path = require('path');
const fs = require('fs-extra');
const pino = require('pino');

const app = express();
const PORT = process.env.PORT || 3000;
app.use(express.json());
app.use(express.static('public'));

let sock = null;
let pairingCode = '';
let isConnected = false;
const prefix = '.';

// Commands load
const commands = new Map();
if (fs.existsSync('./commands')) {
    const commandFiles = fs.readdirSync('./commands').filter(f => f.endsWith('.js'));
    for (const file of commandFiles) {
        const cmd = require(`./commands/${file}`);
        commands.set(cmd.name, cmd);
    }
}

// Groups data - Crash fix
const dataFile = './data/groups.json';
fs.ensureFileSync(dataFile);
let groupSettings = {};
try {
    const data = fs.readFileSync(dataFile, 'utf8');
    groupSettings = data.trim()? JSON.parse(data) : {};
} catch(e) {
    groupSettings = {};
}
function saveSettings() {
    fs.writeJsonSync(dataFile, groupSettings, { spaces: 2 });
}

async function startBot(phoneNumber = '') {
    const { state, saveCreds } = await useMultiFileAuthState('session');
    const { version } = await fetchLatestBaileysVersion();

    sock = makeWASocket({
        version,
        auth: state,
        logger: pino({ level: 'silent' }),
        printQRInTerminal: false,
        markOnlineOnConnect: false,
        syncFullHistory: false
    });

    sock.ev.on('creds.update', saveCreds);

    sock.ev.on('connection.update', async (update) => {
        const { connection, lastDisconnect } = update;
        if (connection === 'open') {
            isConnected = true;
            pairingCode = '';
            console.log('✅ Bot Connected');
            setTimeout(async () => {
                await sock.updatePresence('unavailable'); // Lastseen freeze
                await sock.sendPresenceUpdate('unavailable'); // Private mode
            }, 3000);
        } else if (connection === 'close') {
            isConnected = false;
            const code = lastDisconnect.error?.output?.statusCode;
            if (code!== DisconnectReason.loggedOut) {
                setTimeout(() => startBot(), 5000);
            }
        }
    });

    if (phoneNumber &&!sock.authState.creds.registered) {
        await new Promise(r => setTimeout(r, 3000));
        pairingCode = await sock.requestPairingCode(phoneNumber.replace(/[^0-9]/g, ''));
        console.log('Pairing Code:', pairingCode);
    }

    sock.ev.on('messages.upsert', async ({ messages, type }) => {
        if (type!== 'notify') return;
        const msg = messages[0];
        if (!msg.message || msg.key.fromMe) return;

        const from = msg.key.remoteJid;
        const isGroup = from.endsWith('@g.us');
        const sender = msg.key.participant || from;
        const text = msg.message.conversation || msg.message.extendedTextMessage?.text || msg.message.imageMessage?.caption || '';

        if (isGroup &&!groupSettings[from]) {
            groupSettings[from] = { antivv: false, antidelete: false, antiedit: false };
            saveSettings();
        }

        // Auto ViewOnce
        if (isGroup && groupSettings[from]?.antivv) {
            const viewOnce = msg.message.viewOnceMessageV2 || msg.message.viewOnceMessage;
            if (viewOnce) {
                try {
                    const media = viewOnce.message.imageMessage || viewOnce.message.videoMessage;
                    if (media) {
                        const stream = await downloadContentFromMessage(media, media.mimetype.split('/')[0]);
                        let buffer = Buffer.from([]);
                        for await (const chunk of stream) buffer = Buffer.concat([buffer, chunk]);
                        await sock.sendMessage(from, {
                            [media.mimetype.startsWith('image')? 'image' : 'video']: buffer,
                            caption: `👁️ Auto ViewOnce\n@${sender.split('@')[0]}`,
                            mentions: [sender]
                        });
                    }
                } catch(e) {}
            }
        }

        // Commands
        if (!text.startsWith(prefix)) return;
        const args = text.slice(prefix.length).trim().split(/ +/);
        const cmdName = args.shift().toLowerCase();
        const cmd = commands.get(cmdName);
        if (cmd) {
            try {
                await cmd.execute(sock, msg, args, groupSettings, saveSettings);
            } catch (e) {
                console.log(e);
                sock.sendMessage(from, { text: 'Error ❌' });
            }
        }
    });
}

// API - Vercel ke liye
app.get('/status', (req, res) => res.json({ connected: isConnected, code: pairingCode }));
app.post('/pair', async (req, res) => {
    const { number } = req.body;
    if (!number) return res.json({ error: 'Number do +92300xxxxxxx' });

    // Vercel pe bot start na karo sirf message do
    if (process.env.VERCEL === '1') {
        return res.json({ error: 'Bot Vercel pe nahi chalega. Railway/Replit use karo' });
    }

    if (sock &&!sock.authState.creds.registered) {
        const code = await sock.requestPairingCode(number.replace(/[^0-9]/g, ''));
        pairingCode = code;
        res.json({ code });
    } else res.json({ error: isConnected? 'Connected' : 'Wait' });
});

app.get('/', (req, res) => res.sendFile(path.join(__dirname, 'public', 'index.html')));
app.listen(PORT, () => {
    console.log(`Web: http://localhost:${PORT}`);

    // Vercel pe bot start na karo - 500 error khatam
    if (process.env.VERCEL!== '1') {
        startBot();
    }
});
