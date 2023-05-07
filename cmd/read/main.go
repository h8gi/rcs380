package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/h8gi/rcs380"
)

func main() {
	dev, err := rcs380.NewDevice()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	if err := dev.PacketInit(); err != nil {
		log.Fatal(err)
	}

	if err := dev.PacketSetCommandType(); err != nil {
		log.Fatal(err)

	}

	if err := dev.PacketSwitchRF(); err != nil {
		log.Fatal(err)
	}

	if err := dev.PacketInsetRF('F'); err != nil {
		log.Fatal(err)
	}

	if err := dev.PacketInsetProtocol1(); err != nil {
		log.Fatal(err)
	}

	if err := dev.PacketInsetProtocol2('F'); err != nil {
		log.Fatal(err)
	}

	if err := dev.PacketSenseRequest('F'); err != nil {
		log.Fatal(err)
	}

	if _, err := dev.Read(); err != nil {
		log.Fatal(err)
	}

	resp, err := dev.Read()
	if err != nil {
		log.Fatal(err)
	}
	len := binary.LittleEndian.Uint16(resp[5:7])
	fmt.Printf("%d, %x\n", len, resp)

}
