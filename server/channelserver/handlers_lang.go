package channelserver

func getLangStrings(s *Server) map[string]string {
	strings := make(map[string]string)
	switch s.erupeConfig.Language {
	case "jp":
		strings["language"] = "日本語"
		strings["cafeReset"] = "%d/%dにリセット"
	default:
		strings["language"] = "English"
		strings["cafeReset"] = "Resets on %d/%d"
	}
	return strings
}
