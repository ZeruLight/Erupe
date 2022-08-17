package channelserver

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sort"
	"strings"
	"unicode"
)

type Player struct {
	CharName string
	QuestID  int
}

func getPlayerSlice(s *Server) []Player {
	var p []Player
	var questIndex int

	for _, channel := range s.Channels {
		for _, stage := range channel.stages {
			if len(stage.clients) == 0 {
				continue
			}
			questID := 0
			if stage.isQuest() {
				questIndex++
				questID = questIndex
			}
			for client := range stage.clients {
				p = append(p, Player{
					CharName: client.Name,
					QuestID:  questID,
				})
			}
		}
	}
	return p
}

func getCharacterList(s *Server) string {
	questEmojis := []string{
		":person_in_lotus_position:",
		":white_circle:",
		":red_circle:",
		":blue_circle:",
		":brown_circle:",
		":green_circle:",
		":purple_circle:",
		":yellow_circle:",
		":orange_circle:",
		":black_circle:",
	}

	playerSlice := getPlayerSlice(s)

	sort.SliceStable(playerSlice, func(i, j int) bool {
		return playerSlice[i].QuestID < playerSlice[j].QuestID
	})

	message := fmt.Sprintf("===== Online: %d =====\n", len(playerSlice))
	for _, player := range playerSlice {
		message += fmt.Sprintf("%s %s", questEmojis[player.QuestID], player.CharName)
	}

	return message
}

// onDiscordMessage handles receiving messages from discord and forwarding them ingame.
func (s *Server) onDiscordMessage(ds *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from our bot, or ones that are not in the correct channel.
	if m.Author.Bot || m.ChannelID != s.erupeConfig.Discord.RealtimeChannelID {
		return
	}

	paddedName := strings.TrimSpace(strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, m.Author.Username))

	for i := 0; i < 10-len(m.Author.Username); i++ {
		paddedName += " "
	}

	message := fmt.Sprintf("[D] %s > %s", paddedName, m.Content)
	s.BroadcastChatMessage(s.discordBot.NormalizeDiscordMessage(message))
}
