package wkutil

import (
	"fmt"
	"strings"
)

type SlotBitMap struct {
	bits    []byte
	slotNum uint32
}

func NewSlotBitMap(slotNum uint32) *SlotBitMap {
	var bits []byte
	if slotNum%8 == 0 {
		bits = make([]byte, (slotNum / 8))
	} else {
		bits = make([]byte, (slotNum/8)+1)
	}
	return &SlotBitMap{bits: bits, slotNum: slotNum}
}

func (s *SlotBitMap) SetSlot(num uint32, v bool) {
	index := num / 8
	pos := num % 8
	if v {
		s.bits[index] |= 1 << pos
	} else {
		s.bits[index] = s.bits[index] & ^(1 << pos)
	}
}

func (s *SlotBitMap) SetSlotForRange(start, end uint32, v bool) {
	for i := start; i <= end; i++ {
		s.SetSlot(i, v)
	}
}

func (s *SlotBitMap) GetSlot(num uint32) bool {
	index := num / 8
	pos := num % 8
	return s.bits[index]&(1<<pos) != 0
}

func (s *SlotBitMap) Reset() {
	var bits []byte
	if s.slotNum%8 == 0 {
		bits = make([]byte, (s.slotNum / 8))
	} else {
		bits = make([]byte, (s.slotNum/8)+1)
	}
	s.bits = bits
}

func (s *SlotBitMap) GetBits() []byte {
	return s.bits
}

func (s *SlotBitMap) GetVaildSlotNum() int {
	var count = 0
	for i := 0; i < len(s.bits); i++ {
		b := s.bits[i]
		for j := 0; j < 8; j++ {
			vaild := (b >> j & 0x01) == 1
			if vaild {
				count++
			}
		}
	}
	return count
}

func (s *SlotBitMap) GetVaildSlots() []uint32 {
	var slots = make([]uint32, 0)
	for i := 0; i < len(s.bits); i++ {
		b := s.bits[i]
		for j := 0; j < 8; j++ {
			vaild := (b >> j & 0x01) == 1
			if vaild {
				slots = append(slots, uint32(i*8+j))
			}
		}
	}
	return slots
}

func (s *SlotBitMap) ExportSlots(num int) []byte {
	exportBits := make([]byte, len(s.bits))
	exportNum := num
	for i := len(s.bits) - 1; i >= 0; i-- {
		if exportNum <= 0 {
			break
		}
		b := s.bits[i]
		eb := exportBits[i]
		for j := 8 - 1; j >= 0; j-- {
			if exportNum <= 0 {
				break
			}
			vaild := (b >> j & 0x01) == 1
			if vaild {
				eb = eb | (0x01 << j)
				b = b & (^(0x01 << j))
				exportNum--
			}
		}
		s.bits[i] = b
		exportBits[i] = eb
	}
	return exportBits
}

func (s *SlotBitMap) CleanSlots(slots []byte) {
	if len(slots) == 0 {
		return
	}
	for i := len(s.bits) - 1; i >= 0; i-- {
		b := s.bits[i]
		if len(slots)-(len(s.bits)-i) >= 0 {
			v := slots[len(slots)-(len(s.bits)-i)]
			b = b & (^v)
		}
		s.bits[i] = b
	}
}

func (s *SlotBitMap) MergeSlots(bs ...[]byte) {
	if len(bs) == 0 {
		return
	}
	for i := 0; i < len(s.bits); i++ {
		b := s.bits[i]

		for j := 0; j < len(bs); j++ {
			if i < len(bs[j]) {
				v := bs[j][i]
				b = b | v
			}
		}
		s.bits[i] = b
	}
	return

}

func (s *SlotBitMap) FormatSlots() string {
	slots := s.GetVaildSlots()
	if len(slots) == 0 {
		return ""
	}
	formatStr := make([]string, 0)
	var start uint32 = slots[0]
	for i := 1; i < len(slots); i++ {
		if slots[i]-slots[i-1] != 1 {
			if start == slots[i-1] {
				formatStr = append(formatStr, fmt.Sprintf("%d", start))
			} else {
				formatStr = append(formatStr, fmt.Sprintf("%d-%d", start, slots[i-1]))
			}

			start = slots[i]
		}
		if i == len(slots)-1 {
			if start == slots[i] {
				formatStr = append(formatStr, fmt.Sprintf("%d", start))
			} else {
				formatStr = append(formatStr, fmt.Sprintf("%d-%d", start, slots[i]))
			}
		}
	}
	return strings.Join(formatStr, ",")
}
