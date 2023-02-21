package vm

import (
	"crypto/rand"
	"fmt"
)

func GenerateMAC() string {
	mac := make([]byte, 6)
	_, err := rand.Read(mac)
	if err != nil {
		panic(err)
	}
	mac[0] &= 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
