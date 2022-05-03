package channelserver

import (

	"encoding/hex"
	"encoding/binary"
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) { // RAVIENTE USE THIS
	// RAVI EVENT
	pkt := p.(*mhfpacket.MsgSysOperateRegister)
		var RefID uint8
		var Phase1HP, Phase2HP, Phase3HP, Phase4HP, Phase5HP, Phase6HP, Phase7HP, Phase8HP, Phase9HP, Unknown1, Unknown2, Unknown3, Unknown4, Unknown5, Unknown6, Unknown7, Unknown8, Unknown9, Unknown10, Unknown11, Unknown12, Unknown13, Unknown14, Unknown15, Unknown16, Unknown17, Unknown18, Unknown19, Unknown20 uint32
		var DamageMultiplier uint32

		var NextRavi, RaviStarted, RaviPostTime, RaviType, MaxPlayers, RaviKilled, CarveQuest, Register1, Register2, Register3, Register4, Register5 uint32

		var Support1, Support2, Support3, Support4, Support5, Support6, Support7, Support8, Support9, Support10, Support11, Support12, Support13, Support14, Support15, Support16, Support17, Support18, Support19, Support20, Support21, Support22, Support23, Support24, Support25 uint32
		raviState, err := s.server.db.Query("SELECT refid, phase1hp, phase2hp, phase3hp, phase4hp, phase5hp, phase6hp, phase7hp, phase8hp, phase9hp, unknown1, unknown2, unknown3, unknown4, unknown5, unknown6, unknown7, unknown8, unknown9, unknown10, unknown11, unknown12, unknown13, unknown14, unknown15, unknown16, unknown17, unknown18, unknown19, unknown20, damagemultiplier FROM ravistate WHERE RefID=$1", 29)
			if err != nil {
		panic(err)
		}
		for raviState.Next() {
		err = raviState.Scan(&RefID, &Phase1HP, &Phase2HP, &Phase3HP, &Phase4HP, &Phase5HP, &Phase6HP, &Phase7HP, &Phase8HP, &Phase9HP, &Unknown1, &Unknown2, &Unknown3, &Unknown4, &Unknown5, &Unknown6, &Unknown7, &Unknown8, &Unknown9, &Unknown10, &Unknown11, &Unknown12, &Unknown13, &Unknown14, &Unknown15, &Unknown16, &Unknown17, &Unknown18, &Unknown19, &Unknown20, &DamageMultiplier)
		if err != nil {
			panic("Error in ravistate")
			}
		}
		raviRegister, err := s.server.db.Query("SELECT refid, nextravi, ravistarted, raviposttime, ravitype, maxplayers, ravikilled, carvequest, register1, register2, register3, register4, register5 FROM raviregister WHERE RefID=$1", 12)
			if err != nil {
		panic(err)
		}
		for raviRegister.Next() {
			err = raviRegister.Scan(&RefID, &NextRavi, &RaviStarted, &RaviPostTime, &RaviType, &MaxPlayers, &RaviKilled, &CarveQuest, &Register1, &Register2, &Register3, &Register4, &Register5)
			if err != nil {
				panic("Error in raviregister")
			}
		}
		raviSupport, err := s.server.db.Query("SELECT refid, support1, support2, support3, support4, support5, support6, support7, support8, support9, support10, support11, support12, support13, support14, support15, support16, support17, support18, support19, support20, support21, support22, support23, support24, support25 FROM ravisupport WHERE RefID=$1", 25)
			if err != nil {
		panic(err)
		}
		for raviSupport.Next() {
			err = raviSupport.Scan(&RefID, &Support1, &Support2, &Support3, &Support4, &Support5, &Support6, &Support7, &Support8, &Support9, &Support10, &Support11, &Support12, &Support13, &Support14, &Support15, &Support16, &Support17, &Support18, &Support19, &Support20, &Support21, &Support22, &Support23, &Support24, &Support25)
			if err != nil {
				panic("Error in ravisupport")
			}
		}

	switch pkt.RegisterID {

	case 786461:
		resp := byteframe.NewByteFrame()
		size := 6
		var j int
		for i := 0; i < len(pkt.RawDataPayload)-1; i += size {
			j += size
			if j > len(pkt.RawDataPayload) {
				j = len(pkt.RawDataPayload)
			}
			AddData := binary.BigEndian.Uint32(pkt.RawDataPayload[i+2:j])
			resp.WriteUint8(1)
			resp.WriteUint8(pkt.RawDataPayload[i+1])
			switch pkt.RawDataPayload[i+1] {
				case 0:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET nextravi = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 1:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET ravistarted = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 2:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET ravikilled = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 3:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET raviposttime = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 4:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Register1)
							resp.WriteUint32(Register1 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE raviregister SET register1 = $1 WHERE refid = $2", Register1 + uint32(AddData), 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register1 = $1 WHERE refid = $2", AddData, 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register1 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
					}
				case 5:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET carvequest = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 6:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Register2)
							resp.WriteUint32(Register2 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE raviregister SET register2 = $1 WHERE refid = $2", Register2 + uint32(AddData), 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register2 = $1 WHERE refid = $2", AddData, 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register2 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
					}
				case 7:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Register3)
							resp.WriteUint32(Register3 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE raviregister SET register3 = $1 WHERE refid = $2", Register3 + uint32(AddData), 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register3 = $1 WHERE refid = $2", AddData, 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register3 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
					}
				case 8:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Register4)
							resp.WriteUint32(Register4 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE raviregister SET register4 = $1 WHERE refid = $2", Register4 + uint32(AddData), 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register4 = $1 WHERE refid = $2", AddData, 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register4 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
					}
				case 9:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET maxplayers = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 10:
					resp.WriteUint32(0)
					resp.WriteUint32(AddData)
					_, err = s.server.db.Exec("UPDATE raviregister SET ravitype = $1 WHERE refid = $2", AddData, 12)
					if err != nil {
						s.logger.Fatal("Failed to update raviregister in db")
					}
				case 11:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Register5)
							resp.WriteUint32(Register5 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE raviregister SET register5 = $1 WHERE refid = $2", Register5 + uint32(AddData), 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register5 = $1 WHERE refid = $2", AddData, 12)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE raviregister SET register5 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update raviregister in db")
								}
					}
				default:
					resp.WriteUint32(0)
					resp.WriteUint32(0)
			}
		}
		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		s.notifyplayer()



	case 917533:
		resp := byteframe.NewByteFrame()
		size := 6
		var j int
		for i := 0; i < len(pkt.RawDataPayload)-1; i += size {
			j += size
			if j > len(pkt.RawDataPayload) {
				j = len(pkt.RawDataPayload)
			}
			AddData := binary.BigEndian.Uint32(pkt.RawDataPayload[i+2:j])
			resp.WriteUint8(1)
			resp.WriteUint8(pkt.RawDataPayload[i+1])
			switch pkt.RawDataPayload[i+1] {
				case 0:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase1HP)
							resp.WriteUint32(Phase1HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase1HP = $1 WHERE refid = $2", Phase1HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase1HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase1HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 1:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase2HP)
							resp.WriteUint32(Phase2HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase2HP = $1 WHERE refid = $2", Phase2HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase2HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase2HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 2:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase3HP)
							resp.WriteUint32(Phase3HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase3HP = $1 WHERE refid = $2", Phase3HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase3HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase3HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 3:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase4HP)
							resp.WriteUint32(Phase4HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase4HP = $1 WHERE refid = $2", Phase4HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase4HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase4HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 4:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase5HP)
							resp.WriteUint32(Phase5HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase5HP = $1 WHERE refid = $2", Phase5HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase5HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase5HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 5:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase6HP)
							resp.WriteUint32(Phase6HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase6HP = $1 WHERE refid = $2", Phase6HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase6HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase6HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 6:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase7HP)
							resp.WriteUint32(Phase7HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase7HP = $1 WHERE refid = $2", Phase7HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase7HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase7HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 7:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase8HP)
							resp.WriteUint32(Phase8HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase8HP = $1 WHERE refid = $2", Phase8HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase8HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase8HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 8:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Phase9HP)
							resp.WriteUint32(Phase9HP + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase9HP = $1 WHERE refid = $2", Phase9HP + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase9HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Phase9HP = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 9:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown1)
							resp.WriteUint32(Unknown1 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown1 = $1 WHERE refid = $2", Unknown1 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown1 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown1 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 10:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown2)
							resp.WriteUint32(Unknown2 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown2 = $1 WHERE refid = $2", Unknown2 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown2 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown2 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 11:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown3)
							resp.WriteUint32(Unknown3 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown3 = $1 WHERE refid = $2", Unknown3 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown3 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown3 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 12:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown4)
							resp.WriteUint32(Unknown4 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown4 = $1 WHERE refid = $2", Unknown4 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown4 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown4 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 13:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown5)
							resp.WriteUint32(Unknown5 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown5 = $1 WHERE refid = $2", Unknown5 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown5 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown5 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 14:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown6)
							resp.WriteUint32(Unknown6 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown6 = $1 WHERE refid = $2", Unknown6 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown6 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown6 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 15:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown7)
							resp.WriteUint32(Unknown7 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown7 = $1 WHERE refid = $2", Unknown7 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown7 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown7 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 16:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown8)
							resp.WriteUint32(Unknown8 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown8 = $1 WHERE refid = $2", Unknown8 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown8 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown8 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 17:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown9)
							resp.WriteUint32(Unknown9 + (uint32(AddData) * DamageMultiplier))
							if DamageMultiplier == 1 {
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown9 = $1 WHERE refid = $2", Unknown9 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
							}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown9 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown9 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 18:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown10)
							resp.WriteUint32(Unknown10 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown10 = $1 WHERE refid = $2", Unknown10 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown10 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown10 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 19:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown11)
							resp.WriteUint32(Unknown11 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown11 = $1 WHERE refid = $2", Unknown11 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown11 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown11 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 20:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown12)
							resp.WriteUint32(Unknown12 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown12 = $1 WHERE refid = $2", Unknown12 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown12 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown12 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 21:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown13)
							resp.WriteUint32(Unknown13 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown13 = $1 WHERE refid = $2", Unknown13 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown13 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown13 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 22:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown14)
							resp.WriteUint32(Unknown14 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown14 = $1 WHERE refid = $2", Unknown14 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown14 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown14 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 23:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown15)
							resp.WriteUint32(Unknown15 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown15 = $1 WHERE refid = $2", Unknown15 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown15 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown15 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 24:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown16)
							resp.WriteUint32(Unknown16 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown16 = $1 WHERE refid = $2", Unknown16 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown16 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown16 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 25:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown17)
							resp.WriteUint32(Unknown17 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown17 = $1 WHERE refid = $2", Unknown17 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown17 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown17 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 26:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown18)
							resp.WriteUint32(Unknown18 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown18 = $1 WHERE refid = $2", Unknown18 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown18 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown18 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 27:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown19)
							resp.WriteUint32(Unknown19 + (uint32(AddData) * DamageMultiplier))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown19 = $1 WHERE refid = $2", Unknown19 + (uint32(AddData) * DamageMultiplier), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown19 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown19 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				case 28:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Unknown20)
							resp.WriteUint32(Unknown20 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown20 = $1 WHERE refid = $2", Unknown20 + uint32(AddData), 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown20 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravistate SET Unknown20 = $1 WHERE refid = $2", AddData, 29)
							if err != nil {
								s.logger.Fatal("Failed to update ravistate in db")
								}
					}

				default:
					resp.WriteUint32(0)
					resp.WriteUint32(0)
				}
			}

		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		s.notifyplayer()



	case 851997:
		resp := byteframe.NewByteFrame()
		size := 6
		var j int
		for i := 0; i < len(pkt.RawDataPayload)-1; i += size {
			j += size
			if j > len(pkt.RawDataPayload) {
				j = len(pkt.RawDataPayload)
			}
			AddData := binary.BigEndian.Uint32(pkt.RawDataPayload[i+2:j])
			resp.WriteUint8(1)
			resp.WriteUint8(pkt.RawDataPayload[i+1])
			switch pkt.RawDataPayload[i+1] {
				case 0:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support1)
							resp.WriteUint32(Support1 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support1 = $1 WHERE refid = $2", Support1 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support1 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support1 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 1:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support2)
							resp.WriteUint32(Support2 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support2 = $1 WHERE refid = $2", Support2 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support2 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support2 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 2:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support3)
							resp.WriteUint32(Support3 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support3 = $1 WHERE refid = $2", Support3 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support3 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support3 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 3:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support4)
							resp.WriteUint32(Support4 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support4 = $1 WHERE refid = $2", Support4 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support4 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support4 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 4:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support5)
							resp.WriteUint32(Support5 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support5 = $1 WHERE refid = $2", Support5 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support5 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support5 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 5:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support6)
							resp.WriteUint32(Support6 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support6 = $1 WHERE refid = $2", Support6 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support6 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support6 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 6:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support7)
							resp.WriteUint32(Support7 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support7 = $1 WHERE refid = $2", Support7 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support7 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support7 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 7:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support8)
							resp.WriteUint32(Support8 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support8 = $1 WHERE refid = $2", Support8 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support8 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support8 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 8:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support9)
							resp.WriteUint32(Support9 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support9 = $1 WHERE refid = $2", Support9 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support9 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support9 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 9:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support10)
							resp.WriteUint32(Support10 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support10 = $1 WHERE refid = $2", Support10 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support10 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support10 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 10:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support11)
							resp.WriteUint32(Support11 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support11 = $1 WHERE refid = $2", Support11 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support11 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support11 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 11:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support12)
							resp.WriteUint32(Support12 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support12 = $1 WHERE refid = $2", Support12 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support12 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support12 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 12:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support13)
							resp.WriteUint32(Support13 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support13 = $1 WHERE refid = $2", Support13 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support13 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support13 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 13:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support14)
							resp.WriteUint32(Support14 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support14 = $1 WHERE refid = $2", Support14 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support14 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support14 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 14:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support15)
							resp.WriteUint32(Support15 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support15 = $1 WHERE refid = $2", Support15 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support15 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support15 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 15:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support16)
							resp.WriteUint32(Support16 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support16 = $1 WHERE refid = $2", Support16 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support16 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support16 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 16:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support17)
							resp.WriteUint32(Support17 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support17 = $1 WHERE refid = $2", Support17 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support17 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support17 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 17:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support18)
							resp.WriteUint32(Support18 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support18 = $1 WHERE refid = $2", Support18 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support18 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support18 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 18:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support19)
							resp.WriteUint32(Support19 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support19 = $1 WHERE refid = $2", Support19 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support19 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support19 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 19:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support20)
							resp.WriteUint32(Support20 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support20 = $1 WHERE refid = $2", Support20 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support20 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support20 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 20:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support21)
							resp.WriteUint32(Support21 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support21 = $1 WHERE refid = $2", Support21 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support21 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support21 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 21:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support22)
							resp.WriteUint32(Support22 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support22 = $1 WHERE refid = $2", Support22 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support22 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support22 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 22:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support23)
							resp.WriteUint32(Support23 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support23 = $1 WHERE refid = $2", Support23 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support23 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support23 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 23:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support24)
							resp.WriteUint32(Support24 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support24 = $1 WHERE refid = $2", Support24 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support24 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support24 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				case 24:
					switch pkt.RawDataPayload[i] {
						case 2:
							resp.WriteUint32(Support25)
							resp.WriteUint32(Support25 + uint32(AddData))
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support25 = $1 WHERE refid = $2", Support25 + uint32(AddData), 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support25 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(AddData)
							_, err = s.server.db.Exec("UPDATE ravisupport SET Support25 = $1 WHERE refid = $2", AddData, 25)
							if err != nil {
								s.logger.Fatal("Failed to update ravisupport in db")
								}
					}

				default:
					resp.WriteUint32(0)
					resp.WriteUint32(0)


			}
		}
		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		s.notifyplayer()


	}
}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLoadRegister)

	// ORION TEMPORARY DISABLE (IN WORK)
	// ravi response
	r := pkt.Unk1
	switch r {
	case 12:
		var count int

		err := s.server.db.QueryRow("SELECT COUNT(*) FROM raviregister").Scan(&count)
		switch {
		case err != nil:
    			panic(err)
		default:
			if count == 0 {
				s.server.db.Exec("CALL raviinit()")
			}
		}
		if pkt.RegisterID == 983077 {
			data, _ := hex.DecodeString("000C000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			doAckBufFail(s, pkt.AckHandle, data)
		} else if pkt.RegisterID == 983069 {
			data, _ := hex.DecodeString("000C000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			doAckBufFail(s, pkt.AckHandle, data)
		}
		raviRegister, err := s.server.db.Query("SELECT refid, nextravi, ravistarted, raviposttime, ravitype, maxplayers, ravikilled, carvequest, register1, register2, register3, register4, register5 FROM raviregister WHERE RefID=$1", pkt.Unk1)
		if err != nil {
			panic(err)
		}
		var RefID uint8
		var NextRavi, RaviStarted, RaviPostTime, RaviType, MaxPlayers, RaviKilled, CarveQuest, Register1, Register2, Register3, Register4, Register5 uint32
		resp := byteframe.NewByteFrame()
		for raviRegister.Next() {
			err = raviRegister.Scan(&RefID, &NextRavi, &RaviStarted, &RaviPostTime, &RaviType, &MaxPlayers, &RaviKilled, &CarveQuest, &Register1, &Register2, &Register3, &Register4, &Register5)
			if err != nil {
				panic("Error in raviregister")
			}
				resp.WriteUint8(0)
				resp.WriteUint8(RefID)
				resp.WriteUint32(NextRavi)
				resp.WriteUint32(RaviStarted)
				resp.WriteUint32(RaviKilled)
				resp.WriteUint32(RaviPostTime)
				resp.WriteUint32(Register1)
				resp.WriteUint32(CarveQuest)
				resp.WriteUint32(Register2)
				resp.WriteUint32(Register3)
				resp.WriteUint32(Register4)
				resp.WriteUint32(MaxPlayers)
				resp.WriteUint32(RaviType)
				resp.WriteUint32(Register5)
		}
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	case 29:
		raviState, err := s.server.db.Query("SELECT refid, phase1hp, phase2hp, phase3hp, phase4hp, phase5hp, phase6hp, phase7hp, phase8hp, phase9hp, unknown1, unknown2, unknown3, unknown4, unknown5, unknown6, unknown7, unknown8, unknown9, unknown10, unknown11, unknown12, unknown13, unknown14, unknown15, unknown16, unknown17, unknown18, unknown19, unknown20 FROM ravistate WHERE RefID=$1", pkt.Unk1)
			if err != nil {
		panic(err)
		}
		var RefID uint8
		var Phase1HP, Phase2HP, Phase3HP, Phase4HP, Phase5HP, Phase6HP, Phase7HP, Phase8HP, Phase9HP, Unknown1, Unknown2, Unknown3, Unknown4, Unknown5, Unknown6, Unknown7, Unknown8, Unknown9, Unknown10, Unknown11, Unknown12, Unknown13, Unknown14, Unknown15, Unknown16, Unknown17, Unknown18, Unknown19, Unknown20 uint32
		resp := byteframe.NewByteFrame()
		for raviState.Next() {
			err = raviState.Scan(&RefID, &Phase1HP, &Phase2HP, &Phase3HP, &Phase4HP, &Phase5HP, &Phase6HP, &Phase7HP, &Phase8HP, &Phase9HP, &Unknown1, &Unknown2, &Unknown3, &Unknown4, &Unknown5, &Unknown6, &Unknown7, &Unknown8, &Unknown9, &Unknown10, &Unknown11, &Unknown12, &Unknown13, &Unknown14, &Unknown15, &Unknown16, &Unknown17, &Unknown18, &Unknown19, &Unknown20)
			if err != nil {
				panic("Error in ravistate")
			}
				resp.WriteUint8(0)
				resp.WriteUint8(RefID)
				resp.WriteUint32(Phase1HP)
				resp.WriteUint32(Phase2HP)
				resp.WriteUint32(Phase3HP)
				resp.WriteUint32(Phase4HP)
				resp.WriteUint32(Phase5HP)
				resp.WriteUint32(Phase6HP)
				resp.WriteUint32(Phase7HP)
				resp.WriteUint32(Phase8HP)
				resp.WriteUint32(Phase9HP)
				resp.WriteUint32(Unknown1)
				resp.WriteUint32(Unknown2)
				resp.WriteUint32(Unknown3)
				resp.WriteUint32(Unknown4)
				resp.WriteUint32(Unknown5)
				resp.WriteUint32(Unknown6)
				resp.WriteUint32(Unknown7)
				resp.WriteUint32(Unknown8)
				resp.WriteUint32(Unknown9)
				resp.WriteUint32(Unknown10)
				resp.WriteUint32(Unknown11)
				resp.WriteUint32(Unknown12)
				resp.WriteUint32(Unknown13)
				resp.WriteUint32(Unknown14)
				resp.WriteUint32(Unknown15)
				resp.WriteUint32(Unknown16)
				resp.WriteUint32(Unknown17)
				resp.WriteUint32(Unknown18)
				resp.WriteUint32(Unknown19)
				resp.WriteUint32(Unknown20)
		}
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	case 25:
		raviSupport, err := s.server.db.Query("SELECT refid, support1, support2, support3, support4, support5, support6, support7, support8, support9, support10, support11, support12, support13, support14, support15, support16, support17, support18, support19, support20, support21, support22, support23, support24, support25 FROM ravisupport WHERE RefID=$1", pkt.Unk1)
			if err != nil {
		panic(err)
		}
		var RefID uint8
		var Support1, Support2, Support3, Support4, Support5, Support6, Support7, Support8, Support9, Support10, Support11, Support12, Support13, Support14, Support15, Support16, Support17, Support18, Support19, Support20, Support21, Support22, Support23, Support24, Support25 uint32
		resp := byteframe.NewByteFrame()
		for raviSupport.Next() {
			err = raviSupport.Scan(&RefID, &Support1, &Support2, &Support3, &Support4, &Support5, &Support6, &Support7, &Support8, &Support9, &Support10, &Support11, &Support12, &Support13, &Support14, &Support15, &Support16, &Support17, &Support18, &Support19, &Support20, &Support21, &Support22, &Support23, &Support24, &Support25)
			if err != nil {
				panic("Error in ravisupport")
			}
				resp.WriteUint8(0)
				resp.WriteUint8(RefID)
				resp.WriteUint32(Support1)
				resp.WriteUint32(Support2)
				resp.WriteUint32(Support3)
				resp.WriteUint32(Support4)
				resp.WriteUint32(Support5)
				resp.WriteUint32(Support6)
				resp.WriteUint32(Support7)
				resp.WriteUint32(Support8)
				resp.WriteUint32(Support9)
				resp.WriteUint32(Support10)
				resp.WriteUint32(Support11)
				resp.WriteUint32(Support12)
				resp.WriteUint32(Support13)
				resp.WriteUint32(Support14)
				resp.WriteUint32(Support15)
				resp.WriteUint32(Support16)
				resp.WriteUint32(Support17)
				resp.WriteUint32(Support18)
				resp.WriteUint32(Support19)
				resp.WriteUint32(Support20)
				resp.WriteUint32(Support21)
				resp.WriteUint32(Support22)
				resp.WriteUint32(Support23)
				resp.WriteUint32(Support24)
				resp.WriteUint32(Support25)
		}
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

func (s *Session) notifyplayer() {

		s.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0E, 0x00, 0x1D})

		s.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0D, 0x00, 0x1D})

		s.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0C, 0x00, 0x1D})

}

func (s *Session) notifyall() {

	for session := range s.server.semaphore["hs_l0u3B51J9k3"].clients {
		session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0E, 0x00, 0x1D})
		}

	for session := range s.server.semaphore["hs_l0u3B51J9k3"].clients {
		session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0D, 0x00, 0x1D})
		}

	for session := range s.server.semaphore["hs_l0u3B51J9k3"].clients {
		session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0C, 0x00, 0x1D})
		}


}

func (s *Session) notifyticker() {

	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		s.server.semaphoreLock.Lock()
		getSemaphore := s.server.semaphore["hs_l0u3B51J9k3"]
		s.server.semaphoreLock.Unlock()
			if _, exists := getSemaphore.reservedClientSlots[s.charID]; exists {
			s.notifyall()
		}
	}
}

func handleMsgSysNotifyRegister(s *Session, p mhfpacket.MHFPacket) {}