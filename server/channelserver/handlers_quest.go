package channelserver

import (
	"database/sql"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/decryption"
	ps "erupe-ce/common/pascalstring"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

func handleMsgSysGetFile(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetFile)

	if pkt.IsScenario {
		if s.server.erupeConfig.DevModeOptions.QuestDebugTools && s.server.erupeConfig.DevMode {
			s.logger.Debug(
				"Scenario",
				zap.Uint8("CategoryID", pkt.ScenarioIdentifer.CategoryID),
				zap.Uint32("MainID", pkt.ScenarioIdentifer.MainID),
				zap.Uint8("ChapterID", pkt.ScenarioIdentifer.ChapterID),
				zap.Uint8("Flags", pkt.ScenarioIdentifer.Flags),
			)
		}
		filename := fmt.Sprintf("%d_0_0_0_S%d_T%d_C%d", pkt.ScenarioIdentifer.CategoryID, pkt.ScenarioIdentifer.MainID, pkt.ScenarioIdentifer.Flags, pkt.ScenarioIdentifer.ChapterID)
		// Read the scenario file.
		data, err := os.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("scenarios/%s.bin", filename)))
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to open file: %s/scenarios/%s.bin", s.server.erupeConfig.BinPath, filename))
			// This will crash the game.
			doAckBufSucceed(s, pkt.AckHandle, data)
			return
		}
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		if s.server.erupeConfig.DevModeOptions.QuestDebugTools && s.server.erupeConfig.DevMode {
			s.logger.Debug(
				"Quest",
				zap.String("Filename", pkt.Filename),
			)
		}

		if s.server.erupeConfig.GameplayOptions.SeasonOverride {
			pkt.Filename = seasonConversion(s, pkt.Filename)
		}

		data, err := os.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("quests/%s.bin", pkt.Filename)))
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to open file: %s/quests/%s.bin", s.server.erupeConfig.BinPath, pkt.Filename))
			// This will crash the game.
			doAckBufSucceed(s, pkt.AckHandle, data)
			return
		}
		doAckBufSucceed(s, pkt.AckHandle, data)
	}
}

func questSuffix(s *Session) string {
	// Determine the letter to append for day / night
	var timeSet string
	if TimeGameAbsolute() > 2880 {
		timeSet = "d"
	} else {
		timeSet = "n"
	}
	return fmt.Sprintf("%s%d", timeSet, s.server.Season())
}

func seasonConversion(s *Session, questFile string) string {
	filename := fmt.Sprintf("%s%s", questFile[:5], questSuffix(s))

	// Return original file if file doesn't exist
	if _, err := os.Stat(filename); err == nil {
		return filename
	} else {
		return questFile
	}
}

func handleMsgMhfLoadFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadFavoriteQuest)
	var data []byte
	err := s.server.db.QueryRow("SELECT savefavoritequest FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err == nil && len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgMhfSaveFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveFavoriteQuest)
	dumpSaveData(s, pkt.Data, "favquest")
	s.server.db.Exec("UPDATE characters SET savefavoritequest=$1 WHERE id=$2", pkt.Data, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func loadQuestFile(s *Session, questId int) []byte {
	file, err := os.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("quests/%05dd0.bin", questId)))
	if err != nil {
		return nil
	}

	decrypted := decryption.UnpackSimple(file)
	fileBytes := byteframe.NewByteFrameFromBytes(decrypted)
	fileBytes.SetLE()
	fileBytes.Seek(int64(fileBytes.ReadUint32()), 0)

	// The 320 bytes directly following the data pointer must go directly into the event's body, after the header and before the string pointers.
	questBody := byteframe.NewByteFrameFromBytes(fileBytes.ReadBytes(320))
	questBody.SetLE()
	// Find the master quest string pointer
	questBody.Seek(40, 0)
	fileBytes.Seek(int64(questBody.ReadUint32()), 0)
	questBody.Seek(40, 0)
	// Overwrite it
	questBody.WriteUint32(320)
	questBody.Seek(0, 2)

	// Rewrite the quest strings and their pointers
	var tempString []byte
	newStrings := byteframe.NewByteFrame()
	tempPointer := 352
	for i := 0; i < 8; i++ {
		questBody.WriteUint32(uint32(tempPointer))
		temp := int64(fileBytes.Index())
		fileBytes.Seek(int64(fileBytes.ReadUint32()), 0)
		tempString = fileBytes.ReadNullTerminatedBytes()
		fileBytes.Seek(temp+4, 0)
		tempPointer += len(tempString) + 1
		newStrings.WriteNullTerminatedBytes(tempString)
	}
	questBody.WriteBytes(newStrings.Data())

	return questBody.Data()
}

func makeEventQuest(s *Session, rows *sql.Rows) ([]byte, error) {
	var id, mark uint32
	var questId, activeDuration, inactiveDuration int
	var maxPlayers, questType uint8
	var startTime time.Time
	rows.Scan(&id, &maxPlayers, &questType, &questId, &mark, &startTime, &activeDuration, &inactiveDuration)

	data := loadQuestFile(s, questId)
	if data == nil {
		return nil, fmt.Errorf("failed to load quest file")
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(id)
	bf.WriteUint32(0)
	bf.WriteUint8(0) // Indexer
	switch questType {
	case 16:
		bf.WriteUint8(s.server.erupeConfig.GameplayOptions.RegularRavienteMaxPlayers)
	case 22:
		bf.WriteUint8(s.server.erupeConfig.GameplayOptions.ViolentRavienteMaxPlayers)
	case 40:
		bf.WriteUint8(s.server.erupeConfig.GameplayOptions.BerserkRavienteMaxPlayers)
	case 50:
		bf.WriteUint8(s.server.erupeConfig.GameplayOptions.ExtremeRavienteMaxPlayers)
	case 51:
		bf.WriteUint8(s.server.erupeConfig.GameplayOptions.SmallBerserkRavienteMaxPlayers)
	default:
		bf.WriteUint8(maxPlayers)
	}
	bf.WriteUint8(questType)
	if questType == 9 {
		bf.WriteBool(false)
	} else {
		bf.WriteBool(true)
	}
	bf.WriteUint16(0)
	if _config.ErupeConfig.RealClientMode >= _config.G1 {
		bf.WriteUint32(mark)
	}
	bf.WriteUint16(0)
	bf.WriteUint16(uint16(len(data)))
	bf.WriteBytes(data)
	ps.Uint8(bf, "", true) // What is this string for?

	return bf.Data(), nil
}

func handleMsgMhfEnumerateQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateQuest)
	var totalCount, returnedCount uint16
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)

	currentTime := time.Now()

	// Check the event_quests table to load the quests with rotation system
	rows, err := s.server.db.Query("SELECT id, COALESCE(max_players, 4) AS max_players, quest_type, quest_id, COALESCE(mark, 0) AS mark, start_time, active_duration, inactive_duration FROM event_quests")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, mark uint32
		var questId, activeDuration, inactiveDuration int
		var maxPlayers, questType uint8
		var startTime time.Time
		err := rows.Scan(&id, &maxPlayers, &questType, &questId, &mark, &startTime, &activeDuration, &inactiveDuration)
		if err != nil {
			continue
		}

		// Calculate the rotation time based on start time, active duration, and inactive duration
		rotationTime := startTime.Add(time.Duration(activeDuration+inactiveDuration) * 24 * time.Hour)
		if currentTime.After(rotationTime) {
			// The rotation time has passed, update the start time and reset the rotation
			_, err := s.server.db.Exec("UPDATE event_quests SET start_time = $1 WHERE quest_id = $2", rotationTime, questId)
			if err != nil {
				return
			}
		}

		// Check if the quest is currently active
		if currentTime.After(startTime) && currentTime.Sub(startTime) <= time.Duration(activeDuration) * 24 * time.Hour {
			data, err := makeEventQuest(s, rows)
			if err != nil {
				continue
			} else {
				if len(data) > 896 || len(data) < 352 {
					continue
				} else {
					totalCount++
					if totalCount > pkt.Offset && len(bf.Data()) < 60000 {
						returnedCount++
						bf.WriteBytes(data)
						continue
					}
				}
			}
		}
	}

	type tuneValue struct {
		ID    uint16
		Value uint16
	}

	tuneValues := []tuneValue{
		{ID: 20, Value: 1},
		{ID: 26, Value: 1},
		{ID: 27, Value: 1},
		{ID: 33, Value: 1},
		{ID: 40, Value: 1},
		{ID: 49, Value: 1},
		{ID: 53, Value: 1},
		{ID: 59, Value: 1},
		{ID: 67, Value: 1},
		{ID: 80, Value: 1},
		{ID: 94, Value: 1},
		{ID: 1010, Value: 300},
		{ID: 1011, Value: 300},
		{ID: 1012, Value: 300},
		{ID: 1013, Value: 300},
		{ID: 1014, Value: 200},
		{ID: 1015, Value: 200},
		{ID: 1021, Value: 400},
		{ID: 1023, Value: 8},
		{ID: 1024, Value: 150},
		{ID: 1025, Value: 1},
		{ID: 1026, Value: 999}, // get_grank_cap
		{ID: 1027, Value: 100},
		{ID: 1028, Value: 100},
		{ID: 1030, Value: 8},
		{ID: 1031, Value: 100},
		{ID: 1032, Value: 0},   // isValid_partner
		{ID: 1044, Value: 200}, // get_rate_tload_time_out
		{ID: 1045, Value: 0},   // get_rate_tower_treasure_preset
		{ID: 1046, Value: 99},
		{ID: 1048, Value: 0},  // get_rate_tower_log_disable
		{ID: 1049, Value: 10}, // get_rate_tower_gem_max
		{ID: 1050, Value: 1},  // get_rate_tower_gem_set
		{ID: 1051, Value: 200},
		{ID: 1052, Value: 200},
		{ID: 1063, Value: 50000},
		{ID: 1064, Value: 50000},
		{ID: 1065, Value: 25000},
		{ID: 1066, Value: 25000},
		{ID: 1067, Value: 90},  // get_lobby_member_upper_for_making_room Lv1?
		{ID: 1068, Value: 80},  // get_lobby_member_upper_for_making_room Lv2?
		{ID: 1069, Value: 70},  // get_lobby_member_upper_for_making_room Lv3?
		{ID: 1072, Value: 300}, // get_rate_premium_ravi_tama
		{ID: 1073, Value: 300}, // get_rate_premium_ravi_ax_tama
		{ID: 1074, Value: 300}, // get_rate_premium_ravi_g_tama
		{ID: 1078, Value: 0},
		{ID: 1079, Value: 1},
		{ID: 1080, Value: 1},
		{ID: 1081, Value: 1},
		{ID: 1082, Value: 4},
		{ID: 1083, Value: 2},
		{ID: 1084, Value: 10},
		{ID: 1085, Value: 1},
		{ID: 1086, Value: 4},
		{ID: 1087, Value: 2},
		{ID: 1088, Value: 10},
		{ID: 1089, Value: 1},
		{ID: 1090, Value: 3},
		{ID: 1091, Value: 2},
		{ID: 1092, Value: 10},
		{ID: 1093, Value: 2},
		{ID: 1094, Value: 5},
		{ID: 1095, Value: 2},
		{ID: 1096, Value: 10},
		{ID: 1097, Value: 2},
		{ID: 1098, Value: 5},
		{ID: 1099, Value: 2},
		{ID: 1100, Value: 10},
		{ID: 1101, Value: 2},
		{ID: 1102, Value: 5},
		{ID: 1103, Value: 2},
		{ID: 1104, Value: 10},
		{ID: 1145, Value: 200},
		{ID: 1146, Value: 0}, // isTower_invisible
		{ID: 1147, Value: 0}, // isVenom_playable
		{ID: 1149, Value: 20},
		{ID: 1152, Value: 1130},
		{ID: 1154, Value: 0}, // isDisabled_object_season
		{ID: 1158, Value: 1},
		{ID: 1160, Value: 300},
		{ID: 1162, Value: 1},
		{ID: 1163, Value: 3},
		{ID: 1164, Value: 5},
		{ID: 1165, Value: 1},
		{ID: 1166, Value: 5},
		{ID: 1167, Value: 1},
		{ID: 1168, Value: 3},
		{ID: 1169, Value: 3},
		{ID: 1170, Value: 5},
		{ID: 1171, Value: 1},
		{ID: 1172, Value: 1},
		{ID: 1173, Value: 1},
		{ID: 1174, Value: 2},
		{ID: 1175, Value: 4},
		{ID: 1176, Value: 10},
		{ID: 1177, Value: 4},
		{ID: 1178, Value: 10},
		{ID: 1179, Value: 2},
		{ID: 1180, Value: 5},
		{ID: 3000, Value: 100},
		{ID: 3001, Value: 100},
		{ID: 3002, Value: 100},
		{ID: 3003, Value: 100},
		{ID: 3004, Value: 100},
		{ID: 3005, Value: 100},
		{ID: 3006, Value: 100},
		{ID: 3007, Value: 100},
		{ID: 3008, Value: 100},
		{ID: 3009, Value: 100},
		{ID: 3010, Value: 100},
		{ID: 3011, Value: 100},
		{ID: 3012, Value: 100},
		{ID: 3013, Value: 100},
		{ID: 3014, Value: 100},
		{ID: 3015, Value: 100},
		{ID: 3016, Value: 100},
		{ID: 3017, Value: 100},
		{ID: 3018, Value: 100},
		{ID: 3019, Value: 100},
		{ID: 3020, Value: 100},
		{ID: 3021, Value: 100},
		{ID: 3022, Value: 100},
		{ID: 3023, Value: 100},
		{ID: 3024, Value: 100},
		{ID: 3025, Value: 100},
		{ID: 3286, Value: 200},
		{ID: 3287, Value: 200},
		{ID: 3288, Value: 200},
		{ID: 3289, Value: 200},
		{ID: 3290, Value: 200},
		{ID: 3291, Value: 200},
		{ID: 3292, Value: 200},
		{ID: 3293, Value: 200},
		{ID: 3294, Value: 200},
		{ID: 3295, Value: 200},
		{ID: 3296, Value: 200},
		{ID: 3297, Value: 200},
		{ID: 3298, Value: 200},
		{ID: 3299, Value: 200},
		{ID: 3300, Value: 200},
		{ID: 3301, Value: 200},
		{ID: 3302, Value: 200},
		{ID: 3303, Value: 200},
		{ID: 3304, Value: 200},
		{ID: 3305, Value: 200},
		{ID: 3306, Value: 200},
		{ID: 3307, Value: 200},
		{ID: 3308, Value: 200},
		{ID: 3309, Value: 200},
		{ID: 3310, Value: 200},
		{ID: 3311, Value: 200},
		{ID: 3312, Value: 300},
		{ID: 3313, Value: 300},
		{ID: 3314, Value: 300},
		{ID: 3315, Value: 300},
		{ID: 3316, Value: 300},
		{ID: 3317, Value: 300},
		{ID: 3318, Value: 300},
		{ID: 3319, Value: 300},
		{ID: 3320, Value: 300},
		{ID: 3321, Value: 300},
		{ID: 3322, Value: 300},
		{ID: 3323, Value: 300},
		{ID: 3324, Value: 300},
		{ID: 3325, Value: 300},
		{ID: 3326, Value: 300},
		{ID: 3327, Value: 300},
		{ID: 3328, Value: 300},
		{ID: 3329, Value: 300},
		{ID: 3330, Value: 300},
		{ID: 3331, Value: 300},
		{ID: 3332, Value: 300},
		{ID: 3333, Value: 300},
		{ID: 3334, Value: 300},
		{ID: 3335, Value: 300},
		{ID: 3336, Value: 300},
		{ID: 3337, Value: 300},
		{ID: 3338, Value: 100},
		{ID: 3339, Value: 100},
		{ID: 3340, Value: 100},
		{ID: 3341, Value: 100},
		{ID: 3342, Value: 100},
		{ID: 3343, Value: 100},
		{ID: 3344, Value: 100},
		{ID: 3345, Value: 100},
		{ID: 3346, Value: 100},
		{ID: 3347, Value: 100},
		{ID: 3348, Value: 100},
		{ID: 3349, Value: 100},
		{ID: 3350, Value: 100},
		{ID: 3351, Value: 100},
		{ID: 3352, Value: 100},
		{ID: 3353, Value: 100},
		{ID: 3354, Value: 100},
		{ID: 3355, Value: 100},
		{ID: 3356, Value: 100},
		{ID: 3357, Value: 100},
		{ID: 3358, Value: 100},
		{ID: 3359, Value: 100},
		{ID: 3360, Value: 100},
		{ID: 3361, Value: 100},
		{ID: 3362, Value: 100},
		{ID: 3363, Value: 100},
		{ID: 3364, Value: 100},
		{ID: 3365, Value: 100},
		{ID: 3366, Value: 100},
		{ID: 3367, Value: 100},
		{ID: 3368, Value: 100},
		{ID: 3369, Value: 100},
		{ID: 3370, Value: 100},
		{ID: 3371, Value: 100},
		{ID: 3372, Value: 100},
		{ID: 3373, Value: 100},
		{ID: 3374, Value: 100},
		{ID: 3375, Value: 100},
		{ID: 3376, Value: 100},
		{ID: 3377, Value: 100},
		{ID: 3378, Value: 100},
		{ID: 3379, Value: 100},
		{ID: 3380, Value: 100},
		{ID: 3381, Value: 100},
		{ID: 3382, Value: 100},
		{ID: 3383, Value: 100},
		{ID: 3384, Value: 100},
		{ID: 3385, Value: 100},
		{ID: 3386, Value: 100},
		{ID: 3387, Value: 100},
		{ID: 3388, Value: 100},
		{ID: 3389, Value: 100},
		{ID: 3390, Value: 100},
		{ID: 3391, Value: 100},
		{ID: 3392, Value: 100},
		{ID: 3393, Value: 100},
		{ID: 3394, Value: 100},
		{ID: 3395, Value: 100},
		{ID: 3396, Value: 100},
		{ID: 3397, Value: 100},
		{ID: 3398, Value: 100},
		{ID: 3399, Value: 100},
		{ID: 3400, Value: 100},
		{ID: 3401, Value: 100},
		{ID: 3402, Value: 100},
		{ID: 3416, Value: 100},
		{ID: 3417, Value: 100},
		{ID: 3418, Value: 100},
		{ID: 3419, Value: 100},
		{ID: 3420, Value: 100},
		{ID: 3421, Value: 100},
		{ID: 3422, Value: 100},
		{ID: 3423, Value: 100},
		{ID: 3424, Value: 100},
		{ID: 3425, Value: 100},
		{ID: 3426, Value: 100},
		{ID: 3427, Value: 100},
		{ID: 3428, Value: 100},
		{ID: 3442, Value: 100},
		{ID: 3443, Value: 100},
		{ID: 3444, Value: 100},
		{ID: 3445, Value: 100},
		{ID: 3446, Value: 100},
		{ID: 3447, Value: 100},
		{ID: 3448, Value: 100},
		{ID: 3449, Value: 100},
		{ID: 3450, Value: 100},
		{ID: 3451, Value: 100},
		{ID: 3452, Value: 100},
		{ID: 3453, Value: 100},
		{ID: 3454, Value: 100},
		{ID: 3468, Value: 100},
		{ID: 3469, Value: 100},
		{ID: 3470, Value: 100},
		{ID: 3471, Value: 100},
		{ID: 3472, Value: 100},
		{ID: 3473, Value: 100},
		{ID: 3474, Value: 100},
		{ID: 3475, Value: 100},
		{ID: 3476, Value: 100},
		{ID: 3477, Value: 100},
		{ID: 3478, Value: 100},
		{ID: 3479, Value: 100},
		{ID: 3480, Value: 100},
		{ID: 3494, Value: 0},
		{ID: 3495, Value: 0},
		{ID: 3496, Value: 0},
		{ID: 3497, Value: 0},
		{ID: 3498, Value: 0},
		{ID: 3499, Value: 0},
		{ID: 3500, Value: 0},
		{ID: 3501, Value: 0},
		{ID: 3502, Value: 0},
		{ID: 3503, Value: 0},
		{ID: 3504, Value: 0},
		{ID: 3505, Value: 0},
		{ID: 3506, Value: 0},
		{ID: 3520, Value: 0},
		{ID: 3521, Value: 0},
		{ID: 3522, Value: 0},
		{ID: 3523, Value: 0},
		{ID: 3524, Value: 0},
		{ID: 3525, Value: 0},
		{ID: 3526, Value: 0},
		{ID: 3527, Value: 0},
		{ID: 3528, Value: 0},
		{ID: 3529, Value: 0},
		{ID: 3530, Value: 0},
		{ID: 3531, Value: 0},
		{ID: 3532, Value: 0},
		{ID: 3546, Value: 0},
		{ID: 3547, Value: 0},
		{ID: 3548, Value: 0},
		{ID: 3549, Value: 0},
		{ID: 3550, Value: 0},
		{ID: 3551, Value: 0},
		{ID: 3552, Value: 0},
		{ID: 3553, Value: 0},
		{ID: 3554, Value: 0},
		{ID: 3555, Value: 0},
		{ID: 3556, Value: 0},
		{ID: 3557, Value: 0},
		{ID: 3558, Value: 0},
		{ID: 3572, Value: 0},
		{ID: 3573, Value: 0},
		{ID: 3574, Value: 0},
		{ID: 3575, Value: 0},
		{ID: 3576, Value: 0},
		{ID: 3577, Value: 0},
		{ID: 3578, Value: 0},
		{ID: 3579, Value: 0},
		{ID: 3580, Value: 0},
		{ID: 3581, Value: 0},
		{ID: 3582, Value: 0},
		{ID: 3583, Value: 0},
		{ID: 3584, Value: 0},
	}

	tuneValues = append(tuneValues, tuneValue{1020, uint16(s.server.erupeConfig.GameplayOptions.GCPMultiplier * 100)})

	tuneValues = append(tuneValues, tuneValue{1029, s.server.erupeConfig.GameplayOptions.GUrgentRate})

	if s.server.erupeConfig.GameplayOptions.DisableHunterNavi {
		tuneValues = append(tuneValues, tuneValue{1037, 1})
	}

	if s.server.erupeConfig.GameplayOptions.EnableKaijiEvent {
		tuneValues = append(tuneValues, tuneValue{1106, 1})
	} else {
		tuneValues = append(tuneValues, tuneValue{1106, 0})
	}

	if s.server.erupeConfig.GameplayOptions.EnableHiganjimaEvent {
		tuneValues = append(tuneValues, tuneValue{1144, 1})
	} else {
		tuneValues = append(tuneValues, tuneValue{1144, 0})
	}

	if s.server.erupeConfig.GameplayOptions.EnableNierEvent {
		tuneValues = append(tuneValues, tuneValue{1153, 1})
	} else {
		tuneValues = append(tuneValues, tuneValue{1153, 0})
	}

	if s.server.erupeConfig.GameplayOptions.DisableRoad {
		tuneValues = append(tuneValues, tuneValue{1155, 1})
	} else {
		tuneValues = append(tuneValues, tuneValue{1155, 0})
	}

	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3026, uint16(s.server.erupeConfig.GameplayOptions.GRPMultiplier * 100)})
	}

	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3039, uint16(s.server.erupeConfig.GameplayOptions.GSRPMultiplier * 100)})
	}

	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3052, uint16(s.server.erupeConfig.GameplayOptions.GZennyMultiplier * 100)})
	}
	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3078, uint16(s.server.erupeConfig.GameplayOptions.GZennyMultiplier * 100)})
	}

	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3104, uint16(s.server.erupeConfig.GameplayOptions.MaterialMultiplier * 100)})
	}
	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3130, uint16(s.server.erupeConfig.GameplayOptions.MaterialMultiplier * 100)})
	}

	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3156, s.server.erupeConfig.GameplayOptions.ExtraCarves})
	}
	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3182, s.server.erupeConfig.GameplayOptions.ExtraCarves})
	}
	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3208, s.server.erupeConfig.GameplayOptions.ExtraCarves})
	}
	for i := uint16(0); i < 13; i++ {
		tuneValues = append(tuneValues, tuneValue{i + 3234, s.server.erupeConfig.GameplayOptions.ExtraCarves})
	}

	offset := uint16(time.Now().Unix())
	bf.WriteUint16(offset)

	if _config.ErupeConfig.RealClientMode <= _config.F5 {
		tuneValues = tuneValues[:256]
	} else if _config.ErupeConfig.RealClientMode <= _config.G3 {
		tuneValues = tuneValues[:283]
	} else if _config.ErupeConfig.RealClientMode <= _config.GG {
		tuneValues = tuneValues[:315]
	} else if _config.ErupeConfig.RealClientMode <= _config.G61 {
		tuneValues = tuneValues[:332]
	} else if _config.ErupeConfig.RealClientMode <= _config.G7 {
		tuneValues = tuneValues[:339]
	} else if _config.ErupeConfig.RealClientMode <= _config.G81 {
		tuneValues = tuneValues[:396]
	} else if _config.ErupeConfig.RealClientMode <= _config.G91 {
		tuneValues = tuneValues[:694]
	} else if _config.ErupeConfig.RealClientMode <= _config.G101 {
		tuneValues = tuneValues[:704]
	} else if _config.ErupeConfig.RealClientMode <= _config.Z2 {
		tuneValues = tuneValues[:750]
	}

	bf.WriteUint16(uint16(len(tuneValues)))
	for i := range tuneValues {
		bf.WriteUint16(tuneValues[i].ID ^ offset)
		bf.WriteUint16(offset)
		bf.WriteBytes(make([]byte, 4))
		bf.WriteUint16(tuneValues[i].Value ^ offset)
	}

	vsQuestItems := []uint16{1580, 1581, 1582, 1583, 1584, 1585, 1587, 1588, 1589, 1595, 1596, 1597, 1598, 1599, 1600, 1601, 1602, 1603, 1604}
	vsQuestBets := []struct {
		IsTicket bool
		Quantity uint32
	}{
		{true, 5},
		{false, 1000},
		{false, 5000},
		{false, 10000},
	}
	bf.WriteUint16(uint16(len(vsQuestItems)))
	bf.WriteUint16(0) // Unk array of uint16s
	bf.WriteUint16(uint16(len(vsQuestBets)))
	bf.WriteUint16(0) // Unk

	for i := range vsQuestItems {
		bf.WriteUint16(vsQuestItems[i])
	}
	for i := range vsQuestBets {
		bf.WriteBool(vsQuestBets[i].IsTicket)
		bf.WriteUint8(9)
		bf.WriteUint16(7)
		bf.WriteUint32(vsQuestBets[i].Quantity)
	}

	bf.WriteUint16(totalCount)
	bf.WriteUint16(pkt.Offset)
	bf.Seek(0, io.SeekStart)
	bf.WriteUint16(returnedCount)

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnterTournamentQuest(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdBonusQuestInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdBonusQuestInfo)

	udBonusQuestInfos := []struct {
		Unk0      uint8
		Unk1      uint8
		StartTime uint32 // Unix timestamp (seconds)
		EndTime   uint32 // Unix timestamp (seconds)
		Unk4      uint32
		Unk5      uint8
		Unk6      uint8
	}{} // Blank stub array.

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udBonusQuestInfos)))
	for _, q := range udBonusQuestInfos {
		resp.WriteUint8(q.Unk0)
		resp.WriteUint8(q.Unk1)
		resp.WriteUint32(q.StartTime)
		resp.WriteUint32(q.EndTime)
		resp.WriteUint32(q.Unk4)
		resp.WriteUint8(q.Unk5)
		resp.WriteUint8(q.Unk6)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}
