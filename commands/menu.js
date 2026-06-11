module.exports = {
    name: 'menu',
    execute: async (sock, msg) => {
        const text = `*🤖 Bot Menu*
.ping - Bot check
.menu - Ye menu
.invisible on/off - Ghost mode
.antivv set/unset - Auto viewonce
.antidelete set/unset - Anti delete
.antiedit set/unset - Anti edit`;
        await sock.sendMessage(msg.key.remoteJid, { text });
    }
}
