package channelserver

func getLangStrings(s *Server) map[string]string {
	strings := make(map[string]string)
	switch s.erupeConfig.Language {
	case "jp":
		strings["language"] = "日本語"
		strings["cafeReset"] = "%d/%dにリセット"
		strings["ravienteBerserk"] = "<大討伐：猛狂期>が開催されました！"
		strings["ravienteExtreme"] = "<大討伐：猛狂期【極】>が開催されました！"
		strings["ravienteBerserkSmall"] = "<大討伐：猛狂期（小数）>が開催されました！"

		strings["guildInviteName"] = "猟団勧誘のご案内"
		strings["guildInvite"] = "猟団「%s」からの勧誘通知です。\n「勧誘に返答」より、返答を行ってください。"

		strings["guildInviteSuccessName"] = "成功"
		strings["guildInviteSuccess"] = "あなたは「%s」に参加できました。"

		strings["guildInviteAcceptedName"] = "承諾されました"
		strings["guildInviteAccepted"] = "招待した狩人が「%s」への招待を承諾しました。"

		strings["guildInviteRejectName"] = "却下しました"
		strings["guildInviteReject"] = "あなたは「%s」への参加を却下しました。"

		strings["guildInviteDeclinedName"] = "辞退しました"
		strings["guildInviteDeclined"] = "招待した狩人が「%s」への招待を辞退しました。"
	default:
		strings["language"] = "English"
		strings["cafeReset"] = "Resets on %d/%d"
		strings["ravienteBerserk"] = "<Great Slaying: Berserk> is being held!"
		strings["ravienteExtreme"] = "<Great Slaying: Extreme> is being held!"
		strings["ravienteBerserkSmall"] = "<Great Slaying: Berserk Small> is being held!"

		strings["guildInviteName"] = "Invitation!"
		strings["guildInvite"] = "You have been invited to join\n「%s」\nDo you want to accept?"

		strings["guildInviteSuccessName"] = "Success!"
		strings["guildInviteSuccess"] = "You have successfully joined\n「%s」."

		strings["guildInviteAcceptedName"] = "Accepted"
		strings["guildInviteAccepted"] = "The recipient accepted your invitation to join\n「%s」."

		strings["guildInviteRejectName"] = "Rejected"
		strings["guildInviteReject"] = "You rejected the invitation to join\n「%s」."

		strings["guildInviteDeclinedName"] = "Declined"
		strings["guildInviteDeclined"] = "The recipient declined your invitation to join\n「%s」."
	}
	return strings
}
