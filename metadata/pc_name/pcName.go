package pcname

import (
	"log"

	"golang.org/x/sys/windows"
)

func GetPCName() (name string) {
	name, err := windows.ComputerName()
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}
