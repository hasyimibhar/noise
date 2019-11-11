package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"hash"

	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

type suiteFn func([]byte) (cipher.AEAD, error)
type hashFn func() hash.Hash

const sharedKeyLength = 32

// deriveCipherSuite derives an AEAD cipher suite given an ephemeral shared key
// typically produced from a handshake/key exchange protocol.
func deriveCipherSuite(suiteFn suiteFn, hashFn hashFn, ephemeralSharedKey []byte, context []byte) (cipher.AEAD, []byte, error) {
	deriver := hkdf.New(hashFn, ephemeralSharedKey, nil, context)

	sharedKey := make([]byte, sharedKeyLength)
	if _, err := deriver.Read(sharedKey); err != nil {
		return nil, nil, errors.Wrap(err, "failed to derive key via hkdf")
	}

	suite, err := suiteFn(sharedKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to derive aead suite")
	}

	return suite, sharedKey, nil
}

// AEAD via. AES-256 GCM (Galois Counter Mode).
func Aes256GCM() func(sharedKey []byte) (cipher.AEAD, error) {
	// 	if !cpu.Initialized || (cpu.Initialized && !cpu.ARM64.HasAES && !cpu.X86.HasAES && !cpu.S390X.HasAESGCM) {
	// 		panic("UNSUPPORTED: CPU does not support AES-NI instructions.")
	// 	}
	//
	return func(sharedKey []byte) (cipher.AEAD, error) {
		block, _ := aes.NewCipher(sharedKey)
		suite, _ := cipher.NewGCM(block)

		return suite, nil
	}
}

// AEAD via. ChaCha20 Poly1305. Expects a 256-bit shared key.
func Chacha20Poly1305() func(sharedKey []byte) (cipher.AEAD, error) {
	return func(sharedKey []byte) (cipher.AEAD, error) {
		return chacha20poly1305.New(sharedKey)
	}
}

// AEAD via. XChaCha20 Poly1305. Expected a 256-bit shared key.
func Xchacha20Poly1305() func(sharedKey []byte) (cipher.AEAD, error) {
	return func(sharedKey []byte) (cipher.AEAD, error) {
		return chacha20poly1305.NewX(sharedKey)
	}
}
