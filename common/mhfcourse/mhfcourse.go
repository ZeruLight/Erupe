package mhfcourse

import (
	"golang.org/x/exp/slices"
	"math"
	"time"
)

type Course struct {
	ID     uint16
	Expiry time.Time
}

func (c Course) Aliases() []string {
	aliases := make(map[uint16][]string)
	aliases[1] = []string{"Trial", "TL"}
	aliases[2] = []string{"HunterLife", "HL"}
	aliases[3] = []string{"Extra", "ExtraA", "EX"}
	aliases[4] = []string{"ExtraB"}
	aliases[5] = []string{"Mobile"}
	aliases[6] = []string{"Premium"}
	aliases[7] = []string{"Pallone", "ExtraC"}
	aliases[8] = []string{"Assist", "***ist", "Legend", "Rasta"}
	aliases[9] = []string{"N"}
	aliases[10] = []string{"Hiden", "Secret"}
	aliases[11] = []string{"HunterSupport", "HunterAid", "Support", "Aid", "Royal"}
	aliases[12] = []string{"NBoost", "NetCafeBoost", "Boost"}
	aliases[20] = []string{"DEBUG"}
	aliases[21] = []string{"COG_LINK_EXPIRED"}
	aliases[22] = []string{"360_GOLD"}
	aliases[23] = []string{"PS3_TROP"}
	aliases[24] = []string{"COG"}
	aliases[25] = []string{"CAFE_SP"}
	aliases[26] = []string{"NetCafe", "Cafe", "OfficialCafe", "Official"}
	aliases[27] = []string{"HLRenewing", "HLR", "HLRenewal", "HLRenew", "CardHL"}
	aliases[28] = []string{"EXRenewing", "EXR", "EXRenewal", "EXRenew", "CardEX"}
	aliases[29] = []string{"Free"}
	return aliases[c.ID]
}

func Courses() []Course {
	courses := make([]Course, 32)
	for i := range courses {
		courses[i].ID = uint16(i)
		courses[i].Expiry = time.Time{}
	}
	return courses
}

func (c Course) Value() uint32 {
	return uint32(math.Pow(2, float64(c.ID)))
}

// CourseExists returns true if the named course exists in the given slice
func CourseExists(ID uint16, c []Course) bool {
	for _, course := range c {
		if course.ID == ID {
			return true
		}
	}
	return false
}

// GetCourseStruct returns a slice of Course(s) from a rights integer
func GetCourseStruct(rights uint32) ([]Course, uint32) {
	resp := []Course{{ID: 1, Expiry: time.Time{}}}
	s := Courses()
	slices.SortStableFunc(s, func(i, j Course) bool {
		return i.ID > j.ID
	})
	var normalCafeCourseSet, netcafeCourseSet bool
	for _, course := range s {
		if rights-course.Value() < 0x80000000 {
			switch course.ID {
			case 26:
				if normalCafeCourseSet {
					break
				}
				normalCafeCourseSet = true
				resp = append(resp, Course{ID: 25, Expiry: time.Time{}})
				fallthrough
			case 9:
				if netcafeCourseSet {
					break
				}
				netcafeCourseSet = true
				resp = append(resp, Course{ID: 30, Expiry: time.Time{}})
			}
			course.Expiry = time.Date(2030, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))
			resp = append(resp, course)
			rights -= course.Value()
		}
	}
	rights = 0
	for _, course := range resp {
		rights += course.Value()
	}
	return resp, rights
}
