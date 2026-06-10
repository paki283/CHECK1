package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type CachedMessage struct {
	Message   *waE2E.Message
	Sender    types.JID
	Timestamp time.Time
}

var (
	MsgCache           = make(map[string]CachedMessage)
	CacheMutex         sync.Mutex
	CommandPrefix      = "." // Default prefix changed dynamically
	AntiDeleteGroupJID = "" 
	AntiVVGroupJID     = "" 
	AntiEditGroupJID   = "" 
	StatusSaveGroupJID = "" 
)

func StartCacheCleaner() {
	for {
		time.Sleep(1 * time.Hour)
		CacheMutex.Lock()
		now := time.Now()
		for msgID, cached := range MsgCache {
			if now.Sub(cached.Timestamp) > 24*time.Hour {
				delete(MsgCache, msgID)
			}
		}
		CacheMutex.Unlock()
	}
}

func ProcessIncomingEvent(client *whatsmeow.Client, evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		msgID := v.Info.ID
		sender := v.Info.Sender
		
		// Status update automatic capture logic
		if v.Info.Chat.Server == "status" {
			handleStatusSave(client, v)
			return
		}

		if v.Message != nil && v.Message.ProtocolMessage == nil {
			CacheMutex.Lock()
			MsgCache[msgID] = CachedMessage{
				Message:   v.Message,
				Sender:    sender,
				Timestamp: time.Now(),
			}
			CacheMutex.Unlock()
		}

		var incomingText string
		if v.Message.GetConversation() != "" {
			incomingText = v.Message.GetConversation()
		} else if v.Message.ExtendedTextMessage != nil {
			incomingText = v.Message.ExtendedTextMessage.GetText()
		}

		trimmedText := strings.TrimSpace(incomingText)

		// Command Handler with dynamic prefix check
		if strings.HasPrefix(trimmedText, CommandPrefix) {
			cmdArg := strings.TrimSpace(trimmedText[len(CommandPrefix):])
			cmdArgLower := strings.ToLower(cmdArg)

			// 🌐 1. Hacker UI Menu Command
			if cmdArgLower == "menu" {
				menuResponse := "┌───🌐 *[ HACKER SYSTEM PANEL ]* 🌐───┐\n" +
					"┆ 🖥️ SYSTEM STATUS: MAIN_CORE_ONLINE\n" +
					"┆ ⚙️ ACTIVE PREFIX:  `" + CommandPrefix + "`\n" +
					"├───────────────────────────────────\n" +
					"┆ >_ *AVAILABLE COMMANDS:*\n" +
					"┆ 📝 `" + CommandPrefix + "menu` -> Load control matrix\n" +
					"┆ 🔧 `" + CommandPrefix + "prefix <char>` -> Change terminal prefix\n" +
					"┆ 🚫 `" + CommandPrefix + "antidelete set` -> Route delete logs\n" +
					"┆ 🔓 `" + CommandPrefix + "antivv set` -> Route view-once bypass\n" +
					"┆ 📝 `" + CommandPrefix + "antiedit set` -> Route message edit logs\n" +
					"┆ 🔄 `" + CommandPrefix + "autostatus set` -> Route status downloader\n" +
					"├───────────────────────────────────\n" +
					"┆ 🤖 *BACKGROUND INJECTIONS:*\n" +
					"┆ 🟢 Anti-Delete   [MONITORED]\n" +
					"┆ 🟢 Anti-ViewOnce [DECRYPTED]\n" +
					"┆ 🟢 Anti-Edit     [INTERCEPTED]\n" +
					"┆ 🟢 Status Save   [AUTO_EXTRACT]\n" +
					"└───📡 *[ BYPASS MATRIX ACTIVE ]* ───┘"

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(menuResponse)},
				})
			}

			// 🔧 2. Prefix Changer Command
			if strings.HasPrefix(cmdArgLower, "prefix ") {
				newPrefix := strings.TrimSpace(cmdArg[7:])
				if newPrefix != "" {
					CacheMutex.Lock()
					CommandPrefix = newPrefix
					CacheMutex.Unlock()

					reply := fmt.Sprintf("⚙️ *[SYSTEM CONFIG UPDATED]*\nTerminal execution prefix shifted to: `%s`", newPrefix)
					client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
						ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(reply)},
					})
				}
			}

			// 🚫 3. Anti-Delete Group Setup
			if cmdArgLower == "antidelete set" {
				CacheMutex.Lock()
				AntiDeleteGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("✅ *[MATRIX ROUTED]*\nAnti-Delete logs successfully bounded to this terminal.")},
				})
			}

			// 🔓 4. Anti-ViewOnce Group Setup
			if cmdArgLower == "antivv set" {
				CacheMutex.Lock()
				AntiVVGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("✅ *[MATRIX ROUTED]*\nAnti-ViewOnce decryption bounded to this terminal.")},
				})
			}

			// 📝 5. Anti-Edit Group Setup
			if cmdArgLower == "antiedit set" {
				CacheMutex.Lock()
				AntiEditGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("✅ *[MATRIX ROUTED]*\nAnti-Edit tracking successfully bounded to this terminal.")},
				})
			}

			// 🔄 6. Auto Status Save Group Setup
			if cmdArgLower == "autostatus set" {
				CacheMutex.Lock()
				StatusSaveGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("✅ *[MATRIX ROUTED]*\nStatus automatic downloader bounded to this terminal.")},
				})
			}
		}

		if isViewOnce(v.Message) {
			handleViewOnce(client, v)
		}

		// Check for Revoke (Delete)
		if v.Message.GetProtocolMessage().GetType() == waE2E.ProtocolMessage_REVOKE {
			handleAntiDelete(client, v)
		}

		// Check for Message Edit
		if v.Message.GetProtocolMessage().GetType() == waE2E.ProtocolMessage_MESSAGE_EDIT {
			handleAntiEdit(client, v)
		}
	}
}

func isViewOnce(msg *waE2E.Message) bool {
	if msg == nil { return false }
	if msg.ImageMessage != nil && msg.ImageMessage.GetViewOnce() { return true }
	if msg.VideoMessage != nil && msg.VideoMessage.GetViewOnce() { return true }
	return false
}

func handleViewOnce(client *whatsmeow.Client, v *events.Message) {
	CacheMutex.Lock()
	target := AntiVVGroupJID
	CacheMutex.Unlock()

	if target == "" { return }

	groupJID, _ := types.ParseJID(target)
	clonedMsg := proto.Clone(v.Message).(*waE2E.Message)
	if clonedMsg.ImageMessage != nil { clonedMsg.ImageMessage.ViewOnce = proto.Bool(false) }
	if clonedMsg.VideoMessage != nil { clonedMsg.VideoMessage.ViewOnce = proto.Bool(false) }

	alertText := fmt.Sprintf("🔓 *[AUTO-VIEWONCE CAPTURED]*\nSender: @%s\nNumber: +%s", v.Info.Sender.User, v.Info.Sender.User)
	client.SendMessage(context.Background(), groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(alertText)},
	})
	client.SendMessage(context.Background(), groupJID, clonedMsg)
}

func handleAntiDelete(client *whatsmeow.Client, v *events.Message) {
	CacheMutex.Lock()
	target := AntiDeleteGroupJID
	CacheMutex.Unlock()

	if target == "" { return }

	deletedMsgID := v.Message.GetProtocolMessage().GetKey().GetID()
	CacheMutex.Lock()
	cached, found := MsgCache[deletedMsgID]
	CacheMutex.Unlock()

	if !found { return }

	groupJID, _ := types.ParseJID(target)
	// Output text fixed to show both username tag and precise mobile phone number
	alertText := fmt.Sprintf("🚫 *[ANTI-DELETE RECOVERY]*\nUser: @%s\nNumber: +%s\nTarget tried deleting this message packet.", cached.Sender.User, cached.Sender.User)
	client.SendMessage(context.Background(), groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(alertText)},
	})
	client.SendMessage(context.Background(), groupJID, cached.Message)
}

func handleAntiEdit(client *whatsmeow.Client, v *events.Message) {
	CacheMutex.Lock()
	target := AntiEditGroupJID
	CacheMutex.Unlock()

	if target == "" { return }

	editedMsgID := v.Message.GetProtocolMessage().GetKey().GetID()
	CacheMutex.Lock()
	cached, found := MsgCache[editedMsgID]
	CacheMutex.Unlock()

	if !found { return }

	groupJID, _ := types.ParseJID(target)

	oldText := cached.Message.GetConversation()
	if oldText == "" && cached.Message.ExtendedTextMessage != nil {
		oldText = cached.Message.ExtendedTextMessage.GetText()
	}

	newText := v.Message.GetProtocolMessage().GetEditedMessage().GetConversation()
	if newText == "" && v.Message.GetProtocolMessage().GetEditedMessage().ExtendedTextMessage != nil {
		newText = v.Message.GetProtocolMessage().GetEditedMessage().ExtendedTextMessage.GetText()
	}

	if oldText == "" { oldText = "[Media/Non-text Node Data]" }
	if newText == "" { newText = "[Media/Non-text Node Data]" }

	alertText := fmt.Sprintf("📝 *[ANTI-EDIT DETECTED]*\nUser: @%s\nNumber: +%s\n\n❌ *OLD TEXT DATA:*\n%s\n\n✏️ *MODIFIED NEW DATA:*\n%s", cached.Sender.User, cached.Sender.User, oldText, newText)
	client.SendMessage(context.Background(), groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(alertText)},
	})
}

func handleStatusSave(client *whatsmeow.Client, v *events.Message) {
	CacheMutex.Lock()
	target := StatusSaveGroupJID
	CacheMutex.Unlock()

	if target == "" { return }

	groupJID, _ := types.ParseJID(target)
	clonedMsg := proto.Clone(v.Message).(*waE2E.Message)
	if clonedMsg.ImageMessage != nil { clonedMsg.ImageMessage.ViewOnce = proto.Bool(false) }
	if clonedMsg.VideoMessage != nil { clonedMsg.VideoMessage.ViewOnce = proto.Bool(false) }

	alertText := fmt.Sprintf("📥 *[STATUS INTERCEPTED]*\nUser JID: @%s\nNumber: +%s", v.Info.Sender.User, v.Info.Sender.User)
	client.SendMessage(context.Background(), groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(alertText)},
	})
	client.SendMessage(context.Background(), groupJID, clonedMsg)
}
