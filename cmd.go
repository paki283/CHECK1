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
	CommandPrefix      = "." // Default prefix
	AntiDeleteGroupJID = ""  // Command se set hoga
	AntiVVGroupJID     = ""  // Command se set hoga
	AntiEditGroupJID   = ""  // Command se set hoga
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

		// Command Handler with Dynamic Prefix Check
		if strings.HasPrefix(trimmedText, CommandPrefix) {
			// 🔒 STRICT PRIVATE MODE: Only bot owner can run commands
			if client.Store.ID != nil && sender.User != client.Store.ID.User {
				return
			}

			commandBody := strings.TrimSpace(trimmedText[len(CommandPrefix):])
			cleanedCmd := strings.ToLower(commandBody)

			// 1. menu Command
			if cleanedCmd == "menu" {
				menuResponse := fmt.Sprintf("⚡ *[ PRIVATE BOT COMMANDS ]* ⚡\n\n"+
					"🤖 *Available Commands:*\n"+
					"📝 `%smenu` - Show all active commands\n"+
					"🚫 `%santidelete set` - Target group for deleted messages\n"+
					"🔓 `%santivv set` - Target group for View-Once messages\n"+
					"📝 `%santiedit set` - Target group for Edited messages\n"+
					"🔧 `%sprefix <char>` - Change bot command prefix\n\n"+
					"📌 *Status:* System fully functional and working!", 
					CommandPrefix, CommandPrefix, CommandPrefix, CommandPrefix, CommandPrefix)

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: proto.String(menuResponse),
					},
				})
			}

			// 2. antidelete set Command Logic
			if cleanedCmd == "antidelete set" {
				CacheMutex.Lock()
				AntiDeleteGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: proto.String("✅ *[ANTI-DELETE TARGET SET]*\nAb is group mein saare deleted messages recover hokar ayenge!"),
					},
				})
			}

			// 3. antivv set Command Logic
			if cleanedCmd == "antivv set" {
				CacheMutex.Lock()
				AntiVVGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: proto.String("✅ *[ANTI-VIEWONCE TARGET SET]*\nAb is group mein saare 1-time timer messages open hokar ayenge!"),
					},
				})
			}

			// 4. antiedit set Command Logic
			if cleanedCmd == "antiedit set" {
				CacheMutex.Lock()
				AntiEditGroupJID = v.Info.Chat.String()
				CacheMutex.Unlock()

				client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: proto.String("✅ *[ANTI-EDIT TARGET SET]*\nAb is group mein saare edited messages track hokar ayenge!"),
					},
				})
			}

			// 5. prefix Command Logic
			if strings.HasPrefix(cleanedCmd, "prefix ") {
				newPrefix := strings.TrimSpace(commandBody[7:])
				if newPrefix != "" {
					CacheMutex.Lock()
					CommandPrefix = newPrefix
					CacheMutex.Unlock()

					reply := fmt.Sprintf("✅ *[PREFIX CHANGED]*\nBot ka naya prefix ab `%s` ho chuka hai!", newPrefix)
					client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
						ExtendedTextMessage: &waE2E.ExtendedTextMessage{
							Text: proto.String(reply),
						},
					})
				}
			}
		}

		if isViewOnce(v.Message) {
			handleViewOnce(client, v)
		}

		if v.Message.GetProtocolMessage().GetType() == waE2E.ProtocolMessage_REVOKE {
			handleAntiDelete(client, v)
		}

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

	alertText := fmt.Sprintf("🔓 *[AUTO-VIEWONCE CAPTURED]*\nFrom: @%s", v.Info.Sender.User)
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
	alertText := fmt.Sprintf("🚫 *[ANTI-DELETE RECOVERY]*\nUser @%s tried to delete a message.", cached.Sender.User)
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

	editedMsg := v.Message.GetProtocolMessage().GetEditedMessage()
	newText := editedMsg.GetConversation()
	if newText == "" && editedMsg != nil && editedMsg.ExtendedTextMessage != nil {
		newText = editedMsg.ExtendedTextMessage.GetText()
	}

	if oldText == "" { oldText = "[Non-text or Media Message]" }
	if newText == "" { newText = "[Non-text or Media Message]" }

	alertText := fmt.Sprintf("📝 *[ANTI-EDIT DETECTED]*\nUser: @%s\n\n❌ *Old Message:*\n%s\n\n✏️ *New Message:*\n%s", cached.Sender.User, oldText, newText)
	client.SendMessage(context.Background(), groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String(alertText)},
	})
}
