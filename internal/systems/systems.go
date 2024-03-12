// Package systems вспомогательный пакет для сбора информации о системе.
package systems

import (
	"GophKeeper/internal/models"
	"fmt"
	"net"
)

// Systems описывает структуру для реализации интерфейса по получению информации о системе.
type Systems struct {
}

// MachineInfo возвращает собранную информацию о системе.
func (s *Systems) MachineInfo() (models.Machine, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return models.Machine{}, fmt.Errorf("cannot get interface addresses: %w", err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return models.Machine{IPAddress: ipnet.IP.String()}, nil
			}
		}
	}

	return models.Machine{}, nil
}
