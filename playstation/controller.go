package playstation

import (
	"errors"
)

type DS4Frame struct {
	LeftStickX  int
	LeftStickY  int
	RightStickX int
	RightStickY int

	L1 bool
	L2 bool
	L3 bool
	R1 bool
	R2 bool
	R3 bool

	DPadUp    bool
	DPadDown  bool
	DPadLeft  bool
	DPadRight bool

	Square   bool
	Cross    bool
	Circle   bool
	Triangle bool

	Share    bool
	Options  bool
	TrackPad bool
	PS       bool

	TrackPadTouch0   bool
	TrackPadTouch0ID int
	TrackPadTouch0X  int
	TrackPadTouch0Y  int
	TrackPadTouch1   bool
	TrackPadTouch1ID int
	TrackPadTouch1X  int
	TrackPadTouch1Y  int

	Timestamp int64

	BatterLevel int
}

func (ds4 *DS4Frame) UnmarshalText(data []byte) error {
	var err error
	*ds4, err = ParseDS4Frame(data)
	return err
}

func ParseDS4Frame(data []byte) (DS4Frame, error) {
	ds4 := DS4Frame{}

	if len(data) < 43 {
		return ds4, errors.New("invalid data frame")
	}

	ds4.LeftStickX = int(data[1])
	ds4.LeftStickY = int(data[2])
	ds4.RightStickX = int(data[3])
	ds4.RightStickY = int(data[4])

	ds4.TrackPadTouch0ID = int(data[35] & 0x7f)
	ds4.TrackPadTouch0X = int(data[36])
	ds4.TrackPadTouch0Y = int(data[38] << 4)
	ds4.TrackPadTouch1ID = int(data[39] & 0x7f)
	ds4.TrackPadTouch1X = int(data[41])
	ds4.TrackPadTouch1Y = int(data[42] << 4)

	ds4.Timestamp = int64(data[7] >> 2)

	ds4.BatterLevel = int(data[12])

	if data[6]&0x01 > 0 {
		ds4.L1 = true
	}
	if data[6]&0x04 > 0 {
		ds4.L2 = true
	}
	if data[6]&0x40 > 0 {
		ds4.L3 = true
	}
	if data[6]&0x02 > 0 {
		ds4.R1 = true
	}
	if data[6]&0x08 > 0 {
		ds4.R2 = true
	}
	if data[6]&0x80 > 0 {
		ds4.R3 = true
	}

	if data[5]&16 > 0 {
		ds4.Square = true
	}
	if data[5]&32 > 0 {
		ds4.Cross = true
	}
	if data[5]&64 > 0 {
		ds4.Circle = true
	}
	if data[5]&128 > 0 {
		ds4.Triangle = true
	}

	dpad := data[5] & 15
	if dpad == 0 || dpad == 1 || dpad == 7 {
		ds4.DPadUp = true
	}
	if dpad == 3 || dpad == 4 || dpad == 5 {
		ds4.DPadDown = true
	}
	if dpad == 5 || dpad == 6 || dpad == 7 {
		ds4.DPadLeft = true
	}
	if dpad == 1 || dpad == 2 || dpad == 3 {
		ds4.DPadRight = true
	}

	if data[6]&0x10 > 0 {
		ds4.Share = true
	}
	if data[6]&0x20 > 0 {
		ds4.Options = true
	}
	if data[7]&2 > 0 {
		ds4.TrackPad = true
	}
	if data[7]&1 > 0 {
		ds4.PS = true
	}

	if data[35]>>7 == 0 {
		ds4.TrackPadTouch0 = true
	}
	if (data[37]&0x0f)<<8 > 0 {
		ds4.TrackPadTouch0X = int((data[37] & 0x0f) << 8)
	}
	if (data[37]&0xf0)>>4 > 0 {
		ds4.TrackPadTouch0Y = int((data[37] & 0xf0) >> 4)
	}
	if data[39]>>7 == 0 {
		ds4.TrackPadTouch1 = true
	}
	if (data[41]&0x0f)<<8 > 0 {
		ds4.TrackPadTouch1X = int((data[41] & 0x0f) << 8)
	}
	if (data[41]&0xf0)>>4 > 0 {
		ds4.TrackPadTouch1Y = int((data[41] & 0xf0) >> 4)
	}

	return ds4, nil
}
