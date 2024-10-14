package channelserver

import (
	"erupe-ce/config"
	"strings"
)

type v map[string]string

var translations = map[string]map[string]interface{}{
	"en": {
		"language":                       "English",
		"cafe.reset":                     "Resets on {day}/{month}",
		"timer":                          "Time: {hours}:{minutes}:{seconds}.{milliseconds} ({frames}f)",
		"commands.no_op":                 "You don't have permission to use this command",
		"commands.disabled":              "{command} command is disabled",
		"commands.reload":                "Reloading players...",
		"commands.kqf.get":               "KQF: {kqf}",
		"commands.kqf.set":               "KQF set, please switch Land/World",
		"commands.kqf.version":           "This command is disabled prior to MHFG10",
		"commands.kqf.error.syntax":      "Syntax error, format: {prefix}kqf <get | set xxxxxxxxxxxxxxxx>",
		"commands.rights.success":        "Set rights integer: {rights}",
		"commands.rights.error.syntax":   "Syntax error, format: {prefix}rights <x>",
		"commands.course.disabled":       "{course} Course disabled",
		"commands.course.enabled":        "{course} Course enabled",
		"commands.course.locked":         "{course} Course is locked",
		"commands.course.error.syntax":   "Syntax error, format: {prefix}course <course>",
		"commands.teleport.success":      "Teleporting to {x}, {y}",
		"commands.teleport.error.syntax": "Syntax error, format: {prefix}tp <x> <y>",
		"commands.psn.success":           "Connected PSN ID: {psn}",
		"commands.psn.error.exists":      "{psn} is connected to another account!",
		"commands.psn.error.syntax":      "Syntax error, format: {prefix}psn <psn id>",
		"commands.discord.success":       "Your Discord token: {token}",
		"commands.ban.success.permanent": "Successfully banned {username}",
		"commands.ban.success.temporary": "Successfully banned {username} until {expiry}",
		"commands.ban.error.invalid":     "Invalid Character ID",
		"commands.ban.error.syntax":      "Syntax error, format: {prefix}ban <id> [length]",
		"commands.timer.enabled":         "Quest timer enabled",
		"commands.timer.disabled":        "Quest timer disabled",
		"commands.ravi.start":            "The Great Slaying will begin shortly",
		"commands.ravi.multiplier":       "Raviente multiplier is {multiplier}x",
		"commands.ravi.resurrect.send":   "Sending resurrection support!",
		"commands.ravi.resurrect.error":  "Resurrection support could not be sent",
		"commands.ravi.sedation.send":    "Sending sedation support!",
		"commands.ravi.sedation.request": "Requesting sedation support!",
		"commands.ravi.version":          "This command is disabled prior to MHFZZ",
		"commands.ravi.error.start":      "The Great Slaying could not be started",
		"commands.ravi.error.no_players": "No one has joined the Great Slaying!",
		"commands.ravi.error.syntax":     "Syntax error, format: {prefix}ravi <start | multiplier | sr | ss | rs>",
		"raviente.berserk":               "<Great Slaying: Berserk> is being held!",
		"raviente.extreme":               "<Great Slaying: Extreme> is being held!",
		"raviente.extremelimited":        "<Great Slaying: Extreme (Limited)> is being held!",
		"raviente.berserksmall":          "<Great Slaying: Berserk (Small)> is being held!",
		"guild.invite.invite.title":      "Invited!",
		"guild.invite.invite.body":       "You have been invited to join\n「{guild}」\nDo you want to accept?",
		"guild.invite.success.title":     "Success!",
		"guild.invite.success.body":      "You have successfully joined\n「{guild}」.",
		"guild.invite.accepted.title":    "Accepted",
		"guild.invite.accepted.body":     "The invited Hunter has joined\n「{guild}」.",
		"guild.invite.rejected.title":    "Rejected",
		"guild.invite.rejected.body":     "You have rejected the invitation to join\n「{guild}」.",
		"guild.invite.declined.title":    "Declined",
		"guild.invite.declined.body":     "The invited Hunter has declined the invitation to join\n「{guild}」.",
	},
	"jp": {
		"language":                       "日本語",
		"cafe.reset":                     "{day}/{month}にリセット",
		"timer":                          "タイマー：{hours}'{minutes}\"{seconds}.{milliseconds} ({frames}f)",
		"commands.disabled":              "{command}のコマンドは無効です",
		"commands.reload":                "リロードします",
		"commands.kqf.get":               "現在のキークエストフラグ：{kqf}",
		"commands.kqf.set":               "キークエストのフラグが更新されました。ワールド／ランドを移動してください",
		"commands.kqf.error.syntax":      "エラー　例：{prefix}kqf set xxxxxxxxxxxxxxxx",
		"commands.rights.success":        "コース情報を更新しました：{rights}",
		"commands.rights.error.syntax":   "エラー　例：{prefix}rights <x>",
		"commands.course.disabled":       "{course}コースは無効です",
		"commands.course.enabled":        "{course}コースは有効です",
		"commands.course.locked":         "{course}コースはロックされています",
		"commands.course.error.syntax":   "エラー　例：{prefix}course <コース>",
		"commands.teleport.success":      "{x} {y}にテレポート",
		"commands.teleport.error.syntax": "エラー　例：{prefix}teleport <x> <y>",
		"commands.psn.success":           "PSN「{psn}」が連携されています",
		"commands.psn.error.exists":      "{psn}は既存のユーザに接続されています",
		"commands.psn.error.syntax":      "エラー　例：{prefix}psn <PSN>",
		"commands.discord.success":       "あなたのDiscordトークン：{token}",
		"commands.ban.error.syntax":      "エラー　例：{prefix}ban <ID> [期限]",
		"commands.ravi.start":            "大討伐を開始します",
		"commands.ravi.multiplier":       "ラヴィダメージ倍率：ｘ{multiplier}",
		"commands.ravi.resurrect.send":   "復活支援を実行します",
		"commands.ravi.resurrect.error":  "復活支援は実行されませんでした",
		"commands.ravi.sedation.send":    "鎮静支援を実行します",
		"commands.ravi.sedation.request": "鎮静支援を要請します",
		"commands.ravi.error.start":      "大討伐は既に開催されています",
		"commands.ravi.error.no_players": "誰も大討伐に参加していません",
		"commands.ravi.error.syntax":     "エラー　例：{prefix}ravi <start | multiplier | sr | ss | rs>",
		"raviente.berserk":               "<大討伐：猛狂期>が開催されました！",
		"raviente.extreme":               "<大討伐：猛狂期【極】>が開催されました！",
		"raviente.extremelimited":        "<大討伐：猛狂期【極】(制限付)>が開催されました！",
		"raviente.berserksmall":          "<大討伐：猛狂期(小数)>が開催されました！",
		"guild.invite.invite.title":      "猟団勧誘のご案内",
		"guild.invite.invite.body":       "猟団「{guild}」からの勧誘通知です。\n「勧誘に返答」より、返答を行ってください。",
		"guild.invite.success.title":     "成功",
		"guild.invite.success.body":      "あなたは「{guild}}」に参加できました。",
		"guild.invite.accepted.title":    "承諾されました",
		"guild.invite.accepted.body":     "招待した狩人が「{guild}}」への招待を承諾しました。",
		"guild.invite.rejected.title":    "却下しました",
		"guild.invite.rejected.body":     "あなたは「{guild}}」への参加を却下しました。",
		"guild.invite.declined.title":    "辞退しました",
		"guild.invite.declined.body":     "招待した狩人が「{guild}}」への招待を辞退しました。",
	},
}

// t retrieves the translation for a given key and locale
func t(key string, placeholders map[string]string) string {
	// Look for the translation directly
	if locTranslations, ok := translations[config.GetConfig().Language]; ok {
		if translation, found := locTranslations[key]; found {
			str := translation.(string)
			for placeholder, replacement := range placeholders {
				str = strings.ReplaceAll(str, "{"+placeholder+"}", replacement)
			}
			return str
		}
	}
	// Fallback to returning the key itself if translation is not found
	return key
}
