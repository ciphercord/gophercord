// Package designed for interacting with the official CipherCord Discord bot.
package bot

import (
	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

// The public token of the CipherCord bot. Yes, this is meant to be seen by anyone.
const Token string = "MTEyNjc0NDg5OTk5NjM1NjYzOA." + "GAl6qX." + "uu5QL6kv5lSqJ8Y5wcllXsyrInvvjjIDwjCUOA"

// The messaging channel in the official CipherCord Discord server.
const ChannelID string = "1127831380567523408"

// Any messages sent in the official CipherCord messaging channel will be sent through this chan.
var Messages = make(chan string)

// TODO: Change Messages to RawMessage

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

		Messages <- m.Content
	})

	return session.Open()
}

func Send(s string) error {
	_, err := session.ChannelMessageSend(ChannelID, s)
	return err
}
