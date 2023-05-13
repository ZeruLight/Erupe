package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
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
		if _, err := os.Stat(filepath.Join(s.server.erupeConfig.BinPath, "quest_override.bin")); err == nil {
			data, err := os.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "quest_override.bin"))
			if err != nil {
				panic(err)
			}
			doAckBufSucceed(s, pkt.AckHandle, data)
		} else {
			if s.server.erupeConfig.DevModeOptions.QuestDebugTools && s.server.erupeConfig.DevMode {
				s.logger.Debug(
					"Quest",
					zap.String("Filename", pkt.Filename),
				)
			}
			// Get quest file.
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

func handleMsgMhfEnumerateQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateQuest)
	var totalCount, returnedCount uint16
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	err := filepath.Walk(fmt.Sprintf("%s/events/", s.server.erupeConfig.BinPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		} else {
			if len(data) > 850 || len(data) < 400 {
				return nil // Could be more or less strict with size limits
			} else {
				totalCount++
				if totalCount > pkt.Offset && len(bf.Data()) < 60000 {
					returnedCount++
					bf.WriteBytes(data)
					return nil
				}
			}
		}
		return nil
	})
	if err != nil || totalCount == 0 {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 18))
		return
	}

	tuneValues := []struct {
		ID    uint16
		Value uint16
	}{
		{ID: 608, Value: 1},
		{ID: 612, Value: 0},
		{ID: 613, Value: 0},
		{ID: 614, Value: 1130},
		{ID: 615, Value: 0},
		{ID: 616, Value: 5},
		{ID: 617, Value: 1},
		{ID: 618, Value: 5},
		{ID: 619, Value: 1},
		{ID: 620, Value: 1},
		{ID: 621, Value: 3},
		{ID: 622, Value: 300},
		{ID: 624, Value: 2},
		{ID: 625, Value: 4},
		{ID: 626, Value: 1},
		{ID: 627, Value: 1},
		{ID: 628, Value: 5},
		{ID: 629, Value: 1},
		{ID: 630, Value: 3},
		{ID: 631, Value: 3},
		{ID: 634, Value: 5},
		{ID: 636, Value: 10},
		{ID: 637, Value: 2},
		{ID: 638, Value: 10},
		{ID: 639, Value: 4},
		{ID: 667, Value: 20},
		{ID: 668, Value: 0},
		{ID: 669, Value: 0},
		{ID: 670, Value: 0},
		{ID: 671, Value: 200},
		{ID: 672, Value: 5},
		{ID: 673, Value: 2},
		{ID: 674, Value: 10},
		{ID: 675, Value: 2},
		{ID: 676, Value: 3},
		{ID: 677, Value: 2},
		{ID: 678, Value: 10},
		{ID: 679, Value: 1},
		{ID: 680, Value: 5},
		{ID: 681, Value: 2},
		{ID: 682, Value: 10},
		{ID: 683, Value: 2},
		{ID: 684, Value: 5},
		{ID: 685, Value: 2},
		{ID: 686, Value: 10},
		{ID: 687, Value: 2},
		{ID: 692, Value: 0},
		{ID: 694, Value: 10},
		{ID: 705, Value: 50000},
		{ID: 714, Value: 80},
		{ID: 715, Value: 70},
		{ID: 716, Value: 25000},
		{ID: 717, Value: 90},
		{ID: 718, Value: 50000},
		{ID: 719, Value: 25000},
		{ID: 720, Value: 0},
		{ID: 721, Value: 1},
		{ID: 724, Value: 300},
		{ID: 726, Value: 300},
		{ID: 727, Value: 300},
		{ID: 728, Value: 4},
		{ID: 729, Value: 2},
		{ID: 730, Value: 10},
		{ID: 731, Value: 1},
		{ID: 732, Value: 4},
		{ID: 733, Value: 2},
		{ID: 734, Value: 1},
		{ID: 735, Value: 1},
		{ID: 736, Value: 8},
		{ID: 737, Value: 100},
		{ID: 738, Value: 100},
		{ID: 739, Value: 30},
		{ID: 740, Value: 999},
		{ID: 741, Value: 100},
		{ID: 742, Value: 150},
		{ID: 743, Value: 1},
		{ID: 752, Value: 99},
		{ID: 762, Value: 200},
		{ID: 765, Value: 200},
		{ID: 1296, Value: 200},
		{ID: 1297, Value: 200},
		{ID: 1298, Value: 300},
		{ID: 1299, Value: 300},
		{ID: 1300, Value: 300},
		{ID: 1301, Value: 300},
		{ID: 1305, Value: 8},
		{ID: 1306, Value: 100},
		{ID: 1307, Value: 400},
		{ID: 1701, Value: 1},
		{ID: 1718, Value: 1},
		{ID: 1720, Value: 1},
		{ID: 1735, Value: 1},
		{ID: 1742, Value: 1},
		{ID: 1747, Value: 1},
		{ID: 1751, Value: 1},
		{ID: 1757, Value: 1},
		{ID: 1778, Value: 1},
		{ID: 1788, Value: 1},
		{ID: 1789, Value: 1},
		{ID: 2278, Value: 0},
		{ID: 2560, Value: 200},
		{ID: 2561, Value: 200},
		{ID: 2562, Value: 200},
		{ID: 2563, Value: 200},
		{ID: 2564, Value: 200},
		{ID: 2565, Value: 200},
		{ID: 2566, Value: 200},
		{ID: 2567, Value: 200},
		{ID: 2568, Value: 200},
		{ID: 2569, Value: 200},
		{ID: 2570, Value: 200},
		{ID: 2571, Value: 200},
		{ID: 2572, Value: 200},
		{ID: 2573, Value: 200},
		{ID: 2574, Value: 200},
		{ID: 2575, Value: 200},
		{ID: 2576, Value: 300},
		{ID: 2577, Value: 300},
		{ID: 2578, Value: 300},
		{ID: 2579, Value: 300},
		{ID: 2580, Value: 300},
		{ID: 2581, Value: 300},
		{ID: 2582, Value: 300},
		{ID: 2583, Value: 300},
		{ID: 2584, Value: 300},
		{ID: 2585, Value: 300},
		{ID: 2586, Value: 300},
		{ID: 2587, Value: 300},
		{ID: 2588, Value: 300},
		{ID: 2589, Value: 300},
		{ID: 2590, Value: 300},
		{ID: 2591, Value: 300},
		{ID: 2608, Value: 200},
		{ID: 2609, Value: 200},
		{ID: 2616, Value: 200},
		{ID: 2617, Value: 200},
		{ID: 2618, Value: 200},
		{ID: 2619, Value: 200},
		{ID: 2620, Value: 200},
		{ID: 2621, Value: 200},
		{ID: 2622, Value: 200},
		{ID: 2623, Value: 200},
		{ID: 2624, Value: 0},
		{ID: 2625, Value: 0},
		{ID: 2626, Value: 0},
		{ID: 2627, Value: 0},
		{ID: 2628, Value: 0},
		{ID: 2629, Value: 0},
		{ID: 2632, Value: 0},
		{ID: 2634, Value: 0},
		{ID: 2635, Value: 0},
		{ID: 2636, Value: 0},
		{ID: 2637, Value: 0},
		{ID: 2638, Value: 0},
		{ID: 2639, Value: 0},
		{ID: 2664, Value: 0},
		{ID: 2665, Value: 0},
		{ID: 2666, Value: 0},
		{ID: 2667, Value: 0},
		{ID: 2668, Value: 0},
		{ID: 2669, Value: 0},
		{ID: 2670, Value: 0},
		{ID: 2671, Value: 0},
		{ID: 2674, Value: 0},
		{ID: 2676, Value: 0},
		{ID: 2677, Value: 0},
		{ID: 2678, Value: 0},
		{ID: 2679, Value: 0},
		{ID: 2694, Value: 0},
		{ID: 2696, Value: 0},
		{ID: 2697, Value: 0},
		{ID: 2704, Value: 0},
		{ID: 2705, Value: 0},
		{ID: 2706, Value: 0},
		{ID: 2707, Value: 0},
		{ID: 2708, Value: 0},
		{ID: 2709, Value: 0},
		{ID: 2710, Value: 0},
		{ID: 2711, Value: 0},
		{ID: 2716, Value: 0},
		{ID: 2718, Value: 0},
		{ID: 2719, Value: 0},
		{ID: 2720, Value: 100},
		{ID: 2722, Value: 100},
		{ID: 2723, Value: 100},
		{ID: 2724, Value: 100},
		{ID: 2725, Value: 100},
		{ID: 2726, Value: 100},
		{ID: 2727, Value: 100},
		{ID: 2736, Value: 0},
		{ID: 2737, Value: 0},
		{ID: 2738, Value: 0},
		{ID: 2739, Value: 0},
		{ID: 2744, Value: 0},
		{ID: 2745, Value: 0},
		{ID: 2746, Value: 0},
		{ID: 2747, Value: 0},
		{ID: 2748, Value: 0},
		{ID: 2749, Value: 0},
		{ID: 2750, Value: 0},
		{ID: 2751, Value: 0},
		{ID: 2752, Value: 100},
		{ID: 2753, Value: 100},
		{ID: 2754, Value: 100},
		{ID: 2755, Value: 100},
		{ID: 2756, Value: 100},
		{ID: 2757, Value: 100},
		{ID: 2758, Value: 100},
		{ID: 2759, Value: 100},
		{ID: 2762, Value: 100},
		{ID: 2764, Value: 100},
		{ID: 2765, Value: 100},
		{ID: 2766, Value: 100},
		{ID: 2767, Value: 100},
		{ID: 2776, Value: 100},
		{ID: 2777, Value: 100},
		{ID: 2778, Value: 100},
		{ID: 2779, Value: 100},
		{ID: 2780, Value: 100},
		{ID: 2781, Value: 100},
		{ID: 2784, Value: 100},
		{ID: 2785, Value: 100},
		{ID: 2792, Value: 100},
		{ID: 2793, Value: 100},
		{ID: 2794, Value: 100},
		{ID: 2795, Value: 100},
		{ID: 2796, Value: 100},
		{ID: 2797, Value: 100},
		{ID: 2798, Value: 100},
		{ID: 2799, Value: 100},
		{ID: 2804, Value: 100},
		{ID: 2806, Value: 100},
		{ID: 2807, Value: 100},
		{ID: 2816, Value: 0},
		{ID: 2818, Value: 0},
		{ID: 2819, Value: 0},
		{ID: 2820, Value: 0},
		{ID: 2821, Value: 0},
		{ID: 2822, Value: 0},
		{ID: 2823, Value: 0},
		{ID: 2832, Value: 0},
		{ID: 2833, Value: 0},
		{ID: 2834, Value: 0},
		{ID: 2835, Value: 0},
		{ID: 2840, Value: 0},
		{ID: 2841, Value: 0},
		{ID: 2842, Value: 0},
		{ID: 2843, Value: 0},
		{ID: 2844, Value: 0},
		{ID: 2845, Value: 0},
		{ID: 2846, Value: 0},
		{ID: 2847, Value: 0},
		{ID: 2848, Value: 0},
		{ID: 2849, Value: 0},
		{ID: 2850, Value: 0},
		{ID: 2851, Value: 0},
		{ID: 2852, Value: 0},
		{ID: 2853, Value: 0},
		{ID: 2854, Value: 0},
		{ID: 2855, Value: 0},
		{ID: 2858, Value: 0},
		{ID: 2860, Value: 0},
		{ID: 2861, Value: 0},
		{ID: 2862, Value: 0},
		{ID: 2863, Value: 0},
		{ID: 2872, Value: 0},
		{ID: 2873, Value: 0},
		{ID: 2874, Value: 0},
		{ID: 2875, Value: 0},
		{ID: 2876, Value: 0},
		{ID: 2877, Value: 0},
		{ID: 2880, Value: 0},
		{ID: 2881, Value: 0},
		{ID: 2888, Value: 0},
		{ID: 2889, Value: 0},
		{ID: 2890, Value: 0},
		{ID: 2891, Value: 0},
		{ID: 2892, Value: 0},
		{ID: 2893, Value: 0},
		{ID: 2894, Value: 0},
		{ID: 2895, Value: 0},
		{ID: 2900, Value: 0},
		{ID: 2902, Value: 0},
		{ID: 2903, Value: 0},
		{ID: 2920, Value: 100},
		{ID: 2921, Value: 100},
		{ID: 2922, Value: 100},
		{ID: 2923, Value: 100},
		{ID: 2928, Value: 100},
		{ID: 2929, Value: 100},
		{ID: 2930, Value: 100},
		{ID: 2931, Value: 100},
		{ID: 2932, Value: 100},
		{ID: 2933, Value: 100},
		{ID: 2934, Value: 100},
		{ID: 2935, Value: 100},
		{ID: 2942, Value: 100},
		{ID: 2946, Value: 100},
		{ID: 2948, Value: 100},
		{ID: 2949, Value: 100},
		{ID: 2950, Value: 100},
		{ID: 2951, Value: 100},
		{ID: 2960, Value: 100},
		{ID: 2961, Value: 100},
		{ID: 2962, Value: 100},
		{ID: 2963, Value: 100},
		{ID: 2964, Value: 100},
		{ID: 2965, Value: 100},
		{ID: 2968, Value: 100},
		{ID: 2970, Value: 100},
		{ID: 2971, Value: 100},
		{ID: 2972, Value: 100},
		{ID: 2973, Value: 100},
		{ID: 2974, Value: 100},
		{ID: 2975, Value: 100},
		{ID: 2976, Value: 100},
		{ID: 2977, Value: 100},
		{ID: 2978, Value: 100},
		{ID: 2979, Value: 100},
		{ID: 2980, Value: 100},
		{ID: 2981, Value: 100},
		{ID: 2982, Value: 100},
		{ID: 2983, Value: 100},
		{ID: 2988, Value: 100},
		{ID: 2990, Value: 100},
		{ID: 2991, Value: 100},
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
		{ID: 3026, Value: 100},
		{ID: 3027, Value: 100},
		{ID: 3028, Value: 100},
		{ID: 3029, Value: 100},
		{ID: 3030, Value: 100},
		{ID: 3031, Value: 100},
		{ID: 3032, Value: 100},
		{ID: 3033, Value: 100},
		{ID: 3034, Value: 100},
		{ID: 3035, Value: 100},
		{ID: 3036, Value: 100},
		{ID: 3037, Value: 100},
		{ID: 3038, Value: 100},
		{ID: 3039, Value: 100},
		{ID: 3040, Value: 300},
		{ID: 3041, Value: 300},
		{ID: 3042, Value: 300},
		{ID: 3043, Value: 300},
		{ID: 3044, Value: 300},
		{ID: 3045, Value: 300},
		{ID: 3046, Value: 300},
		{ID: 3047, Value: 300},
		{ID: 3048, Value: 100},
		{ID: 3049, Value: 100},
		{ID: 3050, Value: 100},
		{ID: 3051, Value: 100},
		{ID: 3052, Value: 100},
		{ID: 3053, Value: 100},
		{ID: 3054, Value: 300},
		{ID: 3055, Value: 300},
		{ID: 3056, Value: 100},
		{ID: 3057, Value: 100},
		{ID: 3058, Value: 100},
		{ID: 3059, Value: 100},
		{ID: 3060, Value: 100},
		{ID: 3061, Value: 100},
		{ID: 3062, Value: 100},
		{ID: 3063, Value: 100},
		{ID: 3064, Value: 100},
		{ID: 3065, Value: 100},
		{ID: 3066, Value: 100},
		{ID: 3067, Value: 100},
		{ID: 3068, Value: 100},
		{ID: 3069, Value: 100},
		{ID: 3070, Value: 100},
		{ID: 3071, Value: 100},
		{ID: 3328, Value: 100},
		{ID: 3329, Value: 100},
		{ID: 3330, Value: 100},
		{ID: 3331, Value: 100},
		{ID: 3332, Value: 100},
		{ID: 3333, Value: 100},
		{ID: 3334, Value: 100},
		{ID: 3335, Value: 100},
		{ID: 3336, Value: 100},
		{ID: 3337, Value: 100},
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
		{ID: 3358, Value: 100},
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
		{ID: 3416, Value: 100},
		{ID: 3417, Value: 100},
		{ID: 3418, Value: 100},
		{ID: 3419, Value: 100},
		{ID: 3420, Value: 100},
		{ID: 3421, Value: 100},
		{ID: 3422, Value: 100},
		{ID: 3423, Value: 100},
	}
	//offset := uint16(time.Now().Unix())
	offset := uint16(1766)
	bf.WriteUint16(offset)
	bf.WriteUint16(uint16(len(tuneValues)))
	for i := range tuneValues {
		bf.WriteUint16(tuneValues[i].ID)
		bf.WriteUint16(offset)
		bf.WriteUint32(0xD4D4D400)
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
	bf.WriteUint32(uint32(len(vsQuestBets)))
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
