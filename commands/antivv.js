module.exports = {
    name: 'antivv',
    execute: async (sock, msg, args, settings, save) => {
        const from = msg.key.remoteJid;
        if (!from.endsWith('@g.us')) return sock.sendMessage(from, { text: 'Group me use karo' });
        if (args[0] === 'set') {
            settings[from].antivv = true;
            save();
            sock.sendMessage(from, { text: '✅ Auto ViewOnce ON' });
        } else if (args[0] === 'unset') {
            settings[from].antivv = false;
            save();
            sock.sendMessage(from, { text: '❌ Auto ViewOnce OFF' });
        } else {
            sock.sendMessage(from, { text: '.antivv set ya.antivv unset' });
        }
    }
}
