package discordbot

import (
	"erupe-ce/config"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"regexp"
)

type DiscordBot struct {
	Session         *discordgo.Session
	config          *_config.Config
	logger          *zap.Logger
	MainGuild       *discordgo.Guild
	RealtimeChannel *discordgo.Channel
}

type Options struct {
	Config *_config.Config
	Logger *zap.Logger
}

func NewDiscordBot(options Options) (discordBot *DiscordBot, err error) {
	session, err := discordgo.New("Bot " + options.Config.Discord.BotToken)

	if err != nil {
		options.Logger.Fatal("Discord failed", zap.Error(err))
		return nil, err
	}

	var realtimeChannel *discordgo.Channel

	if options.Config.Discord.RealTimeChannel.Enabled {
		realtimeChannel, err = session.Channel(options.Config.Discord.RealTimeChannel.RealtimeChannelID)
	}

	if err != nil {
		options.Logger.Fatal("Discord failed to create realtimeChannel", zap.Error(err))
		return nil, err
	}

	discordBot = &DiscordBot{
		config:          options.Config,
		logger:          options.Logger,
		Session:         session,
		RealtimeChannel: realtimeChannel,
	}

	return
}

func (bot *DiscordBot) Start() (err error) {
	err = bot.Session.Open()

	return
}

// Replace all mentions to real name from the message.
func (bot *DiscordBot) NormalizeDiscordMessage(message string) string {
	userRegex := regexp.MustCompile(`<@!?(\d{17,19})>`)
	emojiRegex := regexp.MustCompile(`(?:<a?)?:(\w+):(?:\d{18}>)?`)

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

	return result
}

func (bot *DiscordBot) RealtimeChannelSend(message string) (err error) {
	if bot.RealtimeChannel == nil {
		return
	}

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
