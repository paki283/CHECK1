module.exports = {
    name: 'antiedit',
    execute: async (sock, msg, args, settings, save) => {
        const from = msg.key.remoteJid;
        if (!from.endsWith('@g.us')) return sock.sendMessage(from, { text: 'Group me use karo' });
        if (args[0] === 'set') {
            settings[from].antiedit = true;
            save();
            sock.sendMessage(from, { text: '✅ Anti-Edit ON' });
        } else if (args[0] === 'unset') {
            settings[from].antiedit = false;
            save();
            sock.sendMessage(from, { text: '❌ Anti-Edit OFF' });
        }
    }
}
