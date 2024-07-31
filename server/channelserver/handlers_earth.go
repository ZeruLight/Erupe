package channelserver

import (
	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"log"
	"time"
)

func doAckEarthSucceed(s *Session, ackHandle uint32, data []*byteframe.ByteFrame) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(uint32(len(data)))
	for i := range data {
		bf.WriteBytes(data[i].Data())
	}
	doAckBufSucceed(s, ackHandle, bf.Data())
}

func handleMsgMhfGetEarthValue(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthValue)
	type EarthValues struct {
		Value []uint32
	}

	var earthValues []EarthValues
	switch pkt.ReqType {
	case 1:
		earthValues = []EarthValues{
			// {Block, DureSlays, Unk, Unk, Unk, Unk}
			{[]uint32{1, 100, 0, 0, 0, 0}},
			{[]uint32{2, 100, 0, 0, 0, 0}},
		}
	case 2:
		earthValues = []EarthValues{
			// {Block, Floors?, Unk, Unk, Unk, Unk}
			{[]uint32{1, 5771, 0, 0, 0, 0}},
			{[]uint32{2, 1847, 0, 0, 0, 0}},
		}
	case 3:
		earthValues = []EarthValues{
			{[]uint32{1001, 36, 0, 0, 0, 0}},   //getTouhaHistory
			{[]uint32{9001, 3, 0, 0, 0, 0}},    //getKohouhinDropStopFlag  // something to do with ttcSetDisableFlag?
			{[]uint32{9002, 10, 300, 0, 0, 0}}, //getKohouhinForceValue
		}
	}

	var data []*byteframe.ByteFrame
	for _, i := range earthValues {
		bf := byteframe.NewByteFrame()
		for _, j := range i.Value {
			bf.WriteUint32(j)
		}
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}
func cleanupEarthStatus(s *Session) {
	s.server.db.Exec(`DELETE FROM events WHERE event_type='earth'`)
	s.server.db.Exec(`UPDATE characters SET conquest_data=NULL`)
}

func generateEarthStatusTimestamps(s *Session, start uint32, debug bool) []uint32 {
	timestamps := make([]uint32, 4)
	midnight := TimeMidnight()
	if start == 0 || TimeAdjusted().Unix() > int64(start)+1814400 {
		cleanupEarthStatus(s)
		start = uint32(midnight.Add(24 * time.Hour).Unix())
		s.server.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('earth', to_timestamp($1)::timestamp without time zone)", start)
	}
	if debug {
		timestamps[0] = uint32(TimeWeekStart().Unix())
		timestamps[1] = uint32(TimeWeekNext().Unix())
		timestamps[2] = uint32(TimeWeekNext().Add(time.Duration(7) * time.Hour * 24).Unix())
		timestamps[3] = uint32(TimeWeekNext().Add(time.Duration(14) * time.Hour * 24).Unix())
	} else {
		timestamps[0] = start
		timestamps[1] = timestamps[0] + 604800
		timestamps[2] = timestamps[1] + 604800
		timestamps[3] = timestamps[2] + 604800
	}
	return timestamps
}
func handleMsgMhfGetEarthStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthStatus)
	bf := byteframe.NewByteFrame()

	var earthTimestamps []uint32
	var debug = s.server.erupeConfig.EarthDebug
	earthId, earthStart := int32(0x01BEEFEE), uint32(0)
	rows, _ := s.server.db.Queryx("SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='earth'")
	if rows == nil {
		log.Println("No rows found")
	} else {
		for rows.Next() {
			rows.Scan(&earthId, &earthStart)
		}
	}
	earthTimestamps = generateEarthStatusTimestamps(s, earthStart, debug)

	// Conquest
	if uint32(TimeAdjusted().Unix()) > earthTimestamps[0] {
		bf.WriteUint32(earthTimestamps[0]) // Start
		bf.WriteUint32(earthTimestamps[1]) // End
		bf.WriteInt32(1)                   //Conquest Earth Status ID //1 and 2 UNK the difference
		bf.WriteInt32(earthId)             //ID
	} else {
		bf.WriteUint32(earthTimestamps[1]) // Start
		bf.WriteUint32(earthTimestamps[2]) // End
		bf.WriteInt32(2)                   //Conquest Earth Status ID //1 and 2 UNK the difference
		bf.WriteInt32(earthId)             //ID
	}
	for i, m := range s.server.erupeConfig.EarthMonsters {
		//Changed from G9 to G8 to get conquest working in g9.1
		if _config.ErupeConfig.RealClientMode <= _config.G8 {
			if i == 3 {
				break
			}
		}
		if i == 4 {
			break
		}
		bf.WriteInt32(m)
	}

	// Pallone
	if uint32(TimeAdjusted().Unix()) > earthTimestamps[1] {
		bf.WriteUint32(earthTimestamps[1]) // Start
		bf.WriteUint32(earthTimestamps[2]) // End
		bf.WriteInt32(11)                  //Pallone Earth Status ID //11 is Fest //12 is Reward
		bf.WriteInt32(earthId + 1)         //ID
	} else {
		bf.WriteUint32(earthTimestamps[2]) // Start
		bf.WriteUint32(earthTimestamps[3]) // End
		bf.WriteInt32(12)                  //Pallone Earth Status ID //11 is Fest //12 is Reward
		bf.WriteInt32(earthId + 1)         //ID
	}
	for i, m := range s.server.erupeConfig.EarthMonsters {
		//Changed from G9 to G8 to get conquest working in g9.1
		if _config.ErupeConfig.RealClientMode <= _config.G8 {
			if i == 3 {
				break
			}
		}
		if i == 4 {
			break
		}
		bf.WriteInt32(m)
	}

	// Tower
	if uint32(TimeAdjusted().Unix()) > earthTimestamps[2] {
		bf.WriteUint32(earthTimestamps[2]) // Start
		bf.WriteUint32(earthTimestamps[3]) // End
		bf.WriteInt32(21)                  //Tower Earth Status ID
		bf.WriteInt32(earthId + 2)         //ID
		for i, m := range s.server.erupeConfig.EarthMonsters {
			if _config.ErupeConfig.RealClientMode <= _config.G8 {
				if i == 3 {
					break
				}
			}
			if i == 4 {
				break
			}
			bf.WriteInt32(m)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
