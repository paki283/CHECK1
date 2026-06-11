module.exports = {
name: 'autoviewonce',
execute: async (sock, msg, args, groupSettings, saveSettings) => {
const from = msg.key.remoteJid;
if (!from.endsWith('@g.us')) return sock.sendMessage(from, {text: 'Group me use karo'});
const action = args[0];
if (action === 'open') {
groupSettings[from].antivv = true;
saveSettings();
await sock.sendMessage(from, {text: '✅ Auto ViewOnce ON - Group set'});
} else {
groupSettings[from].antivv = false;
saveSettings();
await sock.sendMessage(from, {text: '❌ Auto ViewOnce OFF'});
}
