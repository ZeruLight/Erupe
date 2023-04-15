package channelserver

func getLangStrings(s *Server) map[string]string {
	strings := make(map[string]string)
	switch s.erupeConfig.Language {
	case "jp":
		strings["language"] = "日本語"
		strings["cafeReset"] = "%d/%dにリセット"

		strings["commandDisabled"] = "%sのコマンドは無効です"
		strings["commandReload"] = "リロードします"
		strings["commandKqfGet"] = "現在のキークエストフラグ：%x"
		strings["commandKqfSetError"] = "キークエコマンドエラー　例：%s set xxxxxxxxxxxxxxxx"
		strings["commandKqfSetSuccess"] = "キークエストのフラグが更新されました。ワールド／ランドを移動してください"
		strings["commandRightsError"] = "コース更新コマンドエラー　例：%s x"
		strings["commandRightsSuccess"] = "コース情報を更新しました：%d"
		strings["commandCourseError"] = "コース確認コマンドエラー　例：%s <name>"
		strings["commandCourseDisabled"] = "%sコースは無効です"
		strings["commandCourseEnabled"] = "%sコースは有効です"
		strings["commandCourseLocked"] = "%sコースはロックされています"
		strings["commandTeleportError"] = "テレポートコマンドエラー　構文：%s x y"
		strings["commandTeleportSuccess"] = "%d %dにテレポート"
		strings["commandLinkPSNError"] = "PSN連携コマンドエラー　例：%s <psn id>"
		strings["commandLinkPSNSuccess"] = "PSN「%s」が連携されています"

		strings["commandRaviNoCommand"] = "ラヴィコマンドが指定されていません"
		strings["commandRaviStartSuccess"] = "大討伐を開始します"
		strings["commandRaviStartError"] = "大討伐は既に開催されています"
		strings["commandRaviMultiplier"] = "ラヴィダメージ倍率：ｘ%.2f"
		strings["commandRaviResSuccess"] = "復活支援を実行します"
		strings["commandRaviResError"] = "復活支援は実行されませんでした"
		strings["commandRaviSedSuccess"] = "鎮静支援を実行します"
		strings["commandRaviRequest"] = "鎮静支援を要請します"
		strings["commandRaviError"] = "ラヴィコマンドが認識されません"
		strings["commandRaviNoPlayers"] = "誰も大討伐に参加していません"

		strings["ravienteBerserk"] = "<大討伐：猛狂期>が開催されました！"
		strings["ravienteExtreme"] = "<大討伐：猛狂期【極】>が開催されました！"
		strings["ravienteExtremeLimited"] = "<大討伐：猛狂期【極】(制限付)>が開催されました！"
		strings["ravienteBerserkSmall"] = "<大討伐：猛狂期(小数)>が開催されました！"

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

		strings["commandDisabled"] = "%s command is disabled"
		strings["commandReload"] = "Reloading players..."
		strings["commandKqfGet"] = "KQF: %x"
		strings["commandKqfSetError"] = "Error in command. Format: %s set xxxxxxxxxxxxxxxx"
		strings["commandKqfSetSuccess"] = "KQF set, please switch Land/World"
		strings["commandRightsError"] = "Error in command. Format: %s x"
		strings["commandRightsSuccess"] = "Set rights integer: %d"
		strings["commandCourseError"] = "Error in command. Format: %s <name>"
		strings["commandCourseDisabled"] = "%s Course disabled"
		strings["commandCourseEnabled"] = "%s Course enabled"
		strings["commandCourseLocked"] = "%s Course is locked"
		strings["commandTeleportError"] = "Error in command. Format: %s x y"
		strings["commandTeleportSuccess"] = "Teleporting to %d %d"
		strings["commandLinkPSNError"] = "Error in command. Format: %s <psn id>"
		strings["commandLinkPSNSuccess"] = "Connected PSN ID: %s"

		strings["commandRaviNoCommand"] = "No Raviente command specified!"
		strings["commandRaviStartSuccess"] = "The Great Slaying will begin in a moment"
		strings["commandRaviStartError"] = "The Great Slaying has already begun!"
		strings["commandRaviMultiplier"] = "Raviente multiplier is currently %.2fx"
		strings["commandRaviResSuccess"] = "Sending resurrection support!"
		strings["commandRaviResError"] = "Resurrection support has not been requested!"
		strings["commandRaviSedSuccess"] = "Sending sedation support if requested!"
		strings["commandRaviRequest"] = "Requesting sedation support!"
		strings["commandRaviError"] = "Raviente command not recognised!"
		strings["commandRaviNoPlayers"] = "No one has joined the Great Slaying!"

		strings["ravienteBerserk"] = "<Great Slaying: Berserk> is being held!"
		strings["ravienteExtreme"] = "<Great Slaying: Extreme> is being held!"
		strings["ravienteExtremeLimited"] = "<Great Slaying: Extreme (Limited)> is being held!"
		strings["ravienteBerserkSmall"] = "<Great Slaying: Berserk (Small)> is being held!"

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
