module.exports = {
name: 'antidelete',
execute: async (sock, msg, args, groupSettings, saveSettings) => {
const from = msg.key.remoteJid;
if (!from.endsWith('@g.us')) return sock.sendMessage(from, {text: 'Group me use karo'});
const action = args[0];
if (action === 'set') {
groupSettings[from].antidelete = true;
saveSettings();
await sock.sendMessage(from, {text: '✅ Anti Delete ON - Jo delete karega main bhej dunga'});
} else {
groupSettings[from].antidelete = false;
saveSettings();
await sock.sendMessage(from, {text: '❌ Anti Delete OFF'});
}
