package security

// Obfuscate XORs data with key to prevent simple string extraction.
func Obfuscate(data []byte, key byte) []byte {
	res := make([]byte, len(data))
	for i := range data {
		res[i] = data[i] ^ key
	}
	return res
}

// Deobfuscate is an alias for Obfuscate since XOR is symmetric.
func Deobfuscate(data []byte, key byte) string {
	return string(Obfuscate(data, key))
}
