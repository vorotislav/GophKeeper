package systems

import "GophKeeper/internal/models"

type Systems struct {
}

func (s *Systems) MachineInfo() models.Machine {
	return models.Machine{
		IPAddress:  "127.0.0.1",
		MACAddress: "00:AB:CD:EF:11:22",
		PublicKey:  "public key",
	}
}
