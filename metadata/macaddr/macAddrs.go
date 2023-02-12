package macaddr

import (
	"bytes"
	"log"
	"net"
)

func GetMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
			// Don't use random as we have a real address
			addr = i.HardwareAddr.String()
			break
		}
	}
	return
}
