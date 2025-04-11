package speeddaemon

import (
	"bytes"
)

type ClientMessage interface {
	GetType() U8
	Encode() ([]byte, error)
}

// ERROR MESSAGE

type ErrorMessage struct {
	Content Str
}

func (message *ErrorMessage) GetType() U8 {
	return ErrorMsgType
}

func (message *ErrorMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	if err := writeStr(&buffer, message.Content); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeError(r *SdBuffer) (*ErrorMessage, error) {
	content, err := readStr(r)
	if err != nil {
		return nil, err
	}
	return &ErrorMessage{Content: content}, nil
}

// PLATE MESSAGE

type PlateMessage struct {
	Plate     Str
	Timestamp U32
}

func (message *PlateMessage) GetType() U8 {
	return PlateMsgType
}

func (message *PlateMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	if err := writeStr(&buffer, message.Plate); err != nil {
		return nil, err
	}
	if err := writeU32(&buffer, message.Timestamp); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodePlate(r *SdBuffer) (*PlateMessage, error) {
	plate, err := readStr(r)
	if err != nil {
		return nil, err
	}
	timestamp, err := readU32(r)
	if err != nil {
		return nil, err
	}
	return &PlateMessage{Plate: plate, Timestamp: timestamp}, nil
}

// TICKET MESSAGE

type TicketMessage struct {
	Plate        Str
	Road         U16
	MileOne      U16
	TimeStampOne U32
	MileTwo      U16
	TimeStampTwo U32
	Speed        U16 // 100 * MPH
}

func (message *TicketMessage) GetType() U8 {
	return TicketMsgType
}

func (message *TicketMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	if err := writeStr(&buffer, message.Plate); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.Road); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.MileOne); err != nil {
		return nil, err
	}
	if err := writeU32(&buffer, message.TimeStampOne); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.MileTwo); err != nil {
		return nil, err
	}
	if err := writeU32(&buffer, message.TimeStampTwo); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.Speed); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeTicket(r *SdBuffer) (*TicketMessage, error) {
	plate, err := readStr(r)
	if err != nil {
		return nil, err
	}
	road, err := readU16(r)
	if err != nil {
		return nil, err
	}
	mileOne, err := readU16(r)
	if err != nil {
		return nil, err
	}
	timeStampOne, err := readU32(r)
	if err != nil {
		return nil, err
	}
	mileTwo, err := readU16(r)
	if err != nil {
		return nil, err
	}
	timeStampTwo, err := readU32(r)
	if err != nil {
		return nil, err
	}
	speed, err := readU16(r)
	if err != nil {
		return nil, err
	}
	return &TicketMessage{
		Plate:        plate,
		Road:         road,
		MileOne:      mileOne,
		TimeStampOne: timeStampOne,
		MileTwo:      mileTwo,
		TimeStampTwo: timeStampTwo,
		Speed:        speed,
	}, nil
}

// WANTHEARTBEAT MESSAGE

type WantHeartbeatMessage struct {
	Interval U32 // In deciseconds, 10 per second. 25 = message every 2.5 seconds
}

func (message *WantHeartbeatMessage) GetType() U8 {
	return WantHeartbeatType
}

func (message *WantHeartbeatMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	if err := writeU32(&buffer, message.Interval); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeWantHeartbeat(r *SdBuffer) (*WantHeartbeatMessage, error) {
	interval, err := readU32(r)
	if err != nil {
		return nil, err
	}
	return &WantHeartbeatMessage{Interval: interval}, nil
}

// HEARTBEAT MESSAGE

type HeartbeatMessage struct{}

func (message *HeartbeatMessage) GetType() U8 {
	return HeartbeatType
}

func (message *HeartbeatMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeHeartbeat(r *SdBuffer) (*HeartbeatMessage, error) {
	return &HeartbeatMessage{}, nil
}

// I AM CAMERA MESSAGE

type IAmCameraMessage struct {
	Road  U16
	Mile  U16
	Limit U16 // Miles per hour
}

func (message *IAmCameraMessage) GetType() U8 {
	return IAmCameraType
}

func (message *IAmCameraMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.Road); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.Mile); err != nil {
		return nil, err
	}
	if err := writeU16(&buffer, message.Limit); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeIAmCamera(r *SdBuffer) (*IAmCameraMessage, error) {
	road, err := readU16(r)
	if err != nil {
		return nil, err
	}
	mile, err := readU16(r)
	if err != nil {
		return nil, err
	}
	limit, err := readU16(r)
	if err != nil {
		return nil, err
	}
	return &IAmCameraMessage{Road: road, Mile: mile, Limit: limit}, nil
}

// I AM DISPATCHER MESSAGE

type IAmDispatcherMessage struct {
	Numroads U8
	Roads    []U16
}

func (message *IAmDispatcherMessage) GetType() U8 {
	return IAmDispatcherType
}

func (message *IAmDispatcherMessage) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeU8(&buffer, message.GetType()); err != nil {
		return nil, err
	}
	if err := writeU8(&buffer, message.Numroads); err != nil {
		return nil, err
	}
	for _, road := range message.Roads {
		if err := writeU16(&buffer, road); err != nil {
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

func decodeIAmDispatcher(r *SdBuffer) (*IAmDispatcherMessage, error) {
	numroads, err := readU8(r)
	if err != nil {
		return nil, err
	}
	roads := make([]U16, numroads)
	for i := U8(0); i < numroads; i++ {
		road, err := readU16(r)
		if err != nil {
			return nil, err
		}
		roads[i] = road
	}
	return &IAmDispatcherMessage{Numroads: numroads, Roads: roads}, nil
}
