package common

// Asymmetry подструктура для хранения пути ключей и названия ключей ассиметричного шифрования.
type Asymmetry struct {
	KeysPath   string `koanf:"path"`
	PublicKey  string `koanf:"public"`
	PrivateKey string `koanf:"private"`
}
