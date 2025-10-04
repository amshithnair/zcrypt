package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

// GenerateKeyPair creates and saves a new Ed25519 keypair
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	os.WriteFile("zcrypt_private.key", priv, 0600)
	os.WriteFile("zcrypt_public.key", pub, 0644)
	fmt.Println("✅ Keys generated and saved locally as zcrypt_private.key / zcrypt_public.key.")
	return pub, priv, nil
}

// LoadKey loads the existing keypair
func LoadKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	priv, err := os.ReadFile("zcrypt_private.key")
	if err != nil {
		return nil, nil, fmt.Errorf("❌ private key not found: %v", err)
	}
	pub, err := os.ReadFile("zcrypt_public.key")
	if err != nil {
		return nil, nil, fmt.Errorf("❌ public key not found: %v", err)
	}
	return ed25519.PublicKey(pub), ed25519.PrivateKey(priv), nil
}

// SignMessage signs a log message with the private key
func SignMessage(priv ed25519.PrivateKey, msg []byte) string {
	sig := ed25519.Sign(priv, msg)
	return hex.EncodeToString(sig)
}

// VerifySignature verifies if the provided signature is valid
func VerifySignature(pub ed25519.PublicKey, msg []byte, sigHex string) bool {
	sig, _ := hex.DecodeString(sigHex)
	return ed25519.Verify(pub, msg, sig)
}
