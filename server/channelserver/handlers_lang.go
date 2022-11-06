package channelserver

import "go.uber.org/zap"

func getLangStrings(s *Session) map[string]string {
	var lang string
	strings := make(map[string]string)
	err := s.server.db.QueryRow(`SELECT language FROM users, characters WHERE characters.id = $1 AND users.id = characters.user_id`, s.charID).Scan(&lang)
	if err != nil {
		s.logger.Warn("No language set for user", zap.Uint32("CID", s.charID))
	}
	switch lang {
	case "jp":
		strings["language"] = "日本語"
		strings["cafeReset"] = "%d/%dにリセット"
	default:
		strings["language"] = "English"
		strings["cafeReset"] = "Resets on %d/%d"
	}
	return strings
}
