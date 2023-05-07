package rcs380

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/google/gousb"
)

const (
	vid = 0x054c // SONY
	pid = 0x06c1 // RC-S380/S
)

var insetRFMap = map[rune][]byte{
	'F': {0x00, 0x01, 0x01, 0x0f, 0x01},
	'A': {0x00, 0x02, 0x03, 0x0f, 0x03},
	'B': {0x00, 0x03, 0x07, 0x0f, 0x07},
}

var insetProtocol2Map = map[rune][]byte{
	'F': {0x02, 0x00, 0x18},
	'A': {0x02, 0x00, 0x06, 0x01, 0x00, 0x02, 0x00, 0x05, 0x01, 0x07, 0x07},
	'B': {0x02, 0x00, 0x14, 0x09, 0x01, 0x0a, 0x01, 0x0b, 0x01, 0x0c, 0x01},
}

var requestMap = map[rune][]byte{
	'F': {0x04, 0x6e, 0x00, 0x06, 0x00, 0xff, 0xff, 0x01, 0x00},
	'A': {0x04, 0x6e, 0x00, 0x26},
	'B': {0x04, 0x6e, 0x00, 0x05, 0x00, 0x10},
}

type Device struct {
	dev *gousb.Device
	in  *gousb.InEndpoint
	out *gousb.OutEndpoint
}

func NewDevice() (*Device, error) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	dev, err := ctx.OpenDeviceWithVIDPID(vid, pid)
	if err != nil {
		return nil, err
	}

	if dev == nil {
		return nil, fmt.Errorf("RC-S380 not detected")
	}

	cfg, err := dev.Config(1)
	if err != nil {
		return nil, err
	}
	intf, err := cfg.Interface(0, 0)
	if err != nil {
		return nil, err
	}
	inEnd, err := intf.InEndpoint(1)
	if err != nil {
		return nil, err
	}
	outEnd, err := intf.OutEndpoint(2)
	if err != nil {
		return nil, err
	}

	return &Device{
		dev: dev,
		in:  inEnd,
		out: outEnd,
	}, nil
}

func (d *Device) Close() error {
	return d.dev.Close()
}

// ヘッダー + データ長さ + データ長さのチェックサム + データ + データのチェックサム + フッター
var packetHeader = [5]byte{0x00, 0x00, 0xff, 0xff, 0xff}

// checksum calculates two's complement.
func checksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return ^sum + 1
}

func (d *Device) Write(command []byte) error {
	var b bytes.Buffer
	command = append([]byte{0xd6}, command...)

	n := len(command)
	lenbuf := make([]byte, 2)                        // Create a 2-byte buffer to hold the binary data
	binary.LittleEndian.PutUint16(lenbuf, uint16(n)) // Convert n to a 2-byte binary representation in little-endian byte order
	ncsum := checksum(lenbuf)
	csum := checksum(command)

	b.Write(packetHeader[:])
	b.Write(lenbuf)
	b.WriteByte(ncsum)
	b.Write(command)
	b.WriteByte(csum)
	b.WriteByte(0x00)

	if _, err := d.out.Write(b.Bytes()); err != nil {
		return err
	}

	// receive ack/nck
	if _, err := d.Read(); err != nil {
		return err
	}

	return nil
}

func (d *Device) Read() ([]byte, error) {
	buf := make([]byte, d.in.Desc.MaxPacketSize)
	n, err := d.in.Read(buf)
	return buf[:n], err
}

func (d *Device) PacketInit() error {
	_, err := d.out.Write([]byte{0x00, 0x00, 0xff, 0x00, 0xff, 0x00})
	return err
}

func (d *Device) PacketSetCommandType() error {
	cmd := []byte{0x2a, 0x01}
	return d.Write(cmd)
}

func (d *Device) PacketSwitchRF() error {
	cmd := []byte{0x06, 0x00}
	return d.Write(cmd)
}

func (d *Device) PacketInsetRF(t rune) error {
	cmd, ok := insetRFMap[t]
	if !ok {
		return fmt.Errorf("undefined NFC type: %c", t)
	}
	return d.Write(cmd)
}

func (d *Device) PacketInsetProtocol1() error {
	cmd := []byte{0x02, 0x00, 0x18, 0x01, 0x01, 0x02, 0x01, 0x03, 0x00, 0x04, 0x00, 0x05, 0x00, 0x06, 0x00, 0x07, 0x08, 0x08, 0x00, 0x09, 0x00, 0x0a, 0x00, 0x0b, 0x00, 0x0c, 0x00, 0x0e, 0x04, 0x0f, 0x00, 0x10, 0x00, 0x11, 0x00, 0x12, 0x00, 0x13, 0x06}
	return d.Write(cmd)
}

func (d *Device) PacketInsetProtocol2(t rune) error {
	cmd, ok := insetProtocol2Map[t]
	if !ok {
		return fmt.Errorf("undefined NFC type: %c", t)
	}
	return d.Write(cmd)
}

func (d *Device) PacketSenseRequest(t rune) error {
	cmd, ok := requestMap[t]
	if !ok {
		return fmt.Errorf("undefined NFC type: %c", t)
	}
	return d.Write(cmd)
}
