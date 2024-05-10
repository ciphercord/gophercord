// Package designed for interacting with the official CipherCord Discord bot.
package bot

import (
	"github.com/bwmarrin/discordgo"
)

// The public token of the CipherCord bot. Yes, this is meant to be seen by anyone.
const Token string = "MTEyNjc0NDg5OTk5NjM1NjYzOA." + "GAl6qX." + "uu5QL6kv5lSqJ8Y5wcllXsyrInvvjjIDwjCUOA"

// The messaging channel in the official CipherCord Discord server.
const ChannelID string = "1127831380567523408"

// Any messages sent in the official CipherCord messaging channel will be sent through this chan.
var Messages = make(chan string)

// Starts the CipherCord bot.
func Init() error {
	s, err := discordgo.New(Token)
	if err != nil {
		return err
	}

	s.Identify.Intents = discordgo.IntentGuildMessages

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID != ChannelID {
			return
		}

		Messages <- m.Content
	})

	return s.Open()
}
