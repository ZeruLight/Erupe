package channelserver

func getLangStrings(s *Server) map[string]string {
	strings := make(map[string]string)
	switch s.erupeConfig.Language {
	case "jp":
		strings["language"] = "日本語"
		strings["cafeReset"] = "%d/%dにリセット"
		strings["guildInviteName"] = "猟団勧誘のご案内"
		strings["guildInvite"] = "猟団「%s」からの勧誘通知です。\n「勧誘に返答」より、返答を行ってください。"
	default:
		strings["language"] = "English"
		strings["cafeReset"] = "Resets on %d/%d"
		strings["guildInviteName"] = "Invitation!"
		strings["guildInvite"] = "You have been invited to join\n「%s」\nDo you want to accept?"
	}
	return strings
}
