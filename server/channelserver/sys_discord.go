package channelserver

import (
	"erupe-ce/utils/db"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/crypto/bcrypt"
)

type Player struct {
	CharName string
	QuestID  int
}

func getPlayerSlice(server *ChannelServer) []Player {
	var p []Player
	var questIndex int

	for _, channel := range server.Channels {
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

func getCharacterList(server *ChannelServer) string {
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

	playerSlice := getPlayerSlice(server)

	sort.SliceStable(playerSlice, func(i, j int) bool {
		return playerSlice[i].QuestID < playerSlice[j].QuestID
	})

	message := fmt.Sprintf("===== Online: %d =====\n", len(playerSlice))
	for _, player := range playerSlice {
		message += fmt.Sprintf("%s %s", questEmojis[player.QuestID], player.CharName)
	}

	return message
}

// onInteraction handles slash commands
func (server *ChannelServer) onInteraction(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	db, err := db.GetDB()
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	switch i.Interaction.ApplicationCommandData().Name {
	case "link":
		var temp string
		err := db.QueryRow(`UPDATE users SET discord_id = $1 WHERE discord_token = $2 RETURNING discord_id`, i.Member.User.ID, i.ApplicationCommandData().Options[0].StringValue()).Scan(&temp)
		if err == nil {
			ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your Erupe account was linked successfully.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to link Erupe account.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	case "password":
		password, _ := bcrypt.GenerateFromPassword([]byte(i.ApplicationCommandData().Options[0].StringValue()), 10)
		_, err := db.Exec(`UPDATE users SET password = $1 WHERE discord_id = $2`, password, i.Member.User.ID)
		if err == nil {
			ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your Erupe account password has been updated.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to update Erupe account password.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	}
}

// onDiscordMessage handles receiving messages from discord and forwarding them ingame.
func (server *ChannelServer) onDiscordMessage(ds *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots, or messages that are not in the correct channel.
	if m.Author.Bot || m.ChannelID != server.erupeConfig.Discord.RelayChannel.RelayChannelID {
		return
	}

	paddedName := strings.TrimSpace(strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, m.Author.Username))
	for i := 0; i < 8-len(m.Author.Username); i++ {
		paddedName += " "
	}
	message := server.discordBot.NormalizeDiscordMessage(fmt.Sprintf("[D] %s > %s", paddedName, m.Content))
	if len(message) > server.erupeConfig.Discord.RelayChannel.MaxMessageLength {
		return
	}

	var messages []string
	lineLength := 61
	for i := 0; i < len(message); i += lineLength {
		end := i + lineLength
		if end > len(message) {
			end = len(message)
		}
		messages = append(messages, message[i:end])
	}
	for i := range messages {
		server.BroadcastChatMessage(messages[i])
	}
}
func (server *ChannelServer) DiscordChannelSend(charName string, content string) {
	if server.erupeConfig.Discord.Enabled && server.discordBot != nil {
		message := fmt.Sprintf("**%s**: %s", charName, content)
		server.discordBot.RealtimeChannelSend(message)
	}
}

func (server *ChannelServer) DiscordScreenShotSend(charName string, title string, description string, articleToken string) {
	if server.erupeConfig.Discord.Enabled && server.discordBot != nil {
		imageUrl := fmt.Sprintf("%s:%d/api/ss/bbs/%s", server.erupeConfig.Screenshots.Host, server.erupeConfig.Screenshots.Port, articleToken)
		message := fmt.Sprintf("**%s**: %s - %s %s", charName, title, description, imageUrl)
		server.discordBot.RealtimeChannelSend(message)
	}
}
