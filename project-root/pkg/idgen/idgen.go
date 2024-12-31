package idgen

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	counter int
	mu      sync.Mutex
)

func GenerateUniqueId() (string) {
	hostname, err := os.Hostname()
	if err != nil {}

	hostnameDigits := ""
	for _, r := range hostname {
		if r >= '0' && r <= '9' {
			hostnameDigits += string(r)
		}
	}

	if len(hostnameDigits) > 4 {
		hostnameDigits = hostnameDigits[:4]
	}
	hostnameDigits = fmt.Sprintf("%04s", hostnameDigits)

	timestamp := time.Now().UnixMilli()

	mu.Lock()
	counter = (counter + 1) % 1000
	counterValue := counter
	mu.Unlock()

	uniqueId := fmt.Sprintf("%d%03d%s", timestamp, counterValue, hostnameDigits)

	return uniqueId
}
