module.exports = {
name: 'antiedit',
execute: async (sock, msg, args, groupSettings, saveSettings) => {
const from = msg.key.remoteJid;
if (!from.endsWith('@g.us')) return sock.sendMessage(from, {text: 'Group me use karo'});
const action = args[0];
if (action === 'set') {
groupSettings[from].antiedit = true;
saveSettings();
await sock.sendMessage(from, {text: '✅ Anti Edit ON'});
} else {
groupSettings[from].antiedit = false;
saveSettings();
await sock.sendMessage(from, {text: '❌ Anti Edit OFF'});
}
