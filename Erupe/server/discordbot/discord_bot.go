package discordbot

import (
	"regexp"

	"github.com/Solenataris/Erupe/config"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordBot struct {
	Session         *discordgo.Session
	config          *config.Config
	logger          *zap.Logger
	MainGuild       *discordgo.Guild
	RealtimeChannel *discordgo.Channel
}

type DiscordBotOptions struct {
	Config *config.Config
	Logger *zap.Logger
}

func NewDiscordBot(options DiscordBotOptions) (discordBot *DiscordBot, err error) {
	session, err := discordgo.New("Bot " + options.Config.Discord.BotToken)

	if err != nil {
		options.Logger.Fatal("Discord failed", zap.Error(err))
		return nil, err
	}

	mainGuild, err := session.Guild(options.Config.Discord.ServerID)

	if err != nil {
		options.Logger.Fatal("Discord failed to get main guild", zap.Error(err))
		return nil, err
	}

	realtimeChannel, err := session.Channel(options.Config.Discord.RealtimeChannelID)

	if err != nil {
		options.Logger.Fatal("Discord failed to create realtimeChannel", zap.Error(err))
		return nil, err
	}

	discordBot = &DiscordBot{
		config:          options.Config,
		logger:          options.Logger,
		Session:         session,
		MainGuild:       mainGuild,
		RealtimeChannel: realtimeChannel,
	}

	return
}

func (bot *DiscordBot) Start() (err error) {
	err = bot.Session.Open()

	return
}

func (bot *DiscordBot) FindRoleByID(id string) *discordgo.Role {
	for _, role := range bot.MainGuild.Roles {
		if role.ID == id {
			return role
		}
	}

	return nil
}

// Replace all mentions to real name from the message.
func (bot *DiscordBot) NormalizeDiscordMessage(message string) string {
	userRegex := regexp.MustCompile(`<@!?(\d{17,19})>`)
	emojiRegex := regexp.MustCompile(`(?:<a?)?:(\w+):(?:\d{18}>)?`)
	roleRegex := regexp.MustCompile(`<@&(\d{17,19})>`)

	result := ReplaceTextAll(message, userRegex, func(userId string) string {
		user, err := bot.Session.User(userId)

		if err != nil {
			return "@unknown" // @Unknown
		}

		return "@" + user.Username
	})

	result = ReplaceTextAll(result, emojiRegex, func(emojiName string) string {
		return ":" + emojiName + ":"
	})

	result = ReplaceTextAll(result, roleRegex, func(roleId string) string {
		role := bot.FindRoleByID(roleId)

		if role != nil {
			return "@!" + role.Name
		}

		return "@!unknown"
	})

	return string(result)
}

func (bot *DiscordBot) RealtimeChannelSend(message string) (err error) {
	_, err = bot.Session.ChannelMessageSend(bot.RealtimeChannel.ID, message)

	return
}

func ReplaceTextAll(text string, regex *regexp.Regexp, handler func(input string) string) string {
	result := regex.ReplaceAllFunc([]byte(text), func(s []byte) []byte {
		input := regex.ReplaceAllString(string(s), `$1`)

		return []byte(handler(input))
	})

	return string(result)
}
