// Package designed for interacting with the official CipherCord Discord bot.
package bot

import (
	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

// The public token of the CipherCord bot. Yes, this is meant to be seen by anyone.
const Token string = "MTI0MDg0MzIyNDkyNzU2Nzk4Mg." + "G42Qkq." + "oAK5X3SuhUCdKD3yLI9SsUbpGiCmkIB4a3rUQY"

// The messaging channel in the official CipherCord Discord server.
const ChannelID string = "1127831380567523408"

// The contents of any new message sent in the official CipherCord messaging channel will be sent through here.
var RawMessages = make(chan string)

// Starts the CipherCord bot.
func Init() error {
	var err error
	session, err = discordgo.New("Bot " + Token)
	if err != nil {
		return err
	}

	session.Identify.Intents = discordgo.IntentGuildMessages

	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID != ChannelID {
			return
		}

		RawMessages <- m.Content
	})

	return session.Open()
}

// Sends string to the official CipherCord messaging channel.
func Send(s string) error {
	_, err := session.ChannelMessageSend(ChannelID, s)
	return err
}

// FIXME: RawMessageHistory just freezes when you use it.

// Gets the amt most recent messages in raw form and sends it through RawMessages.
func RawMessageHistory(amt int) error {
	msgs, err := session.ChannelMessages(ChannelID, amt, "", "", "")
	if err != nil {
		return err
	}

	for _, m := range msgs {
		RawMessages <- m.Content
	}

	return nil
}
