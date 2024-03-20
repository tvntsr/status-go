//go:build js
// +build js

// Package secp256k1 wraps the bitcoin secp256k1 C library.
package secp256k1
import (
	"errors"
	"fmt"
	"math/big"
)

var (
	ErrInvalidMsgLen       = errors.New("invalid message length, need 32 bytes")
	ErrInvalidSignatureLen = errors.New("invalid signature length")
	ErrInvalidRecoveryID   = errors.New("invalid signature recovery id")
	ErrInvalidKey          = errors.New("invalid private key")
	ErrInvalidPubkey       = errors.New("invalid public key")
	ErrSignFailed          = errors.New("signing failed")
	ErrRecoverFailed       = errors.New("recovery failed")
)

// Sign creates a recoverable ECDSA signature.
// The produced signature is in the 65-byte [R || S || V] format where V is 0 or 1.
//
// The caller is responsible for ensuring that msg cannot be chosen
// directly by an attacker. It is usually preferable to use a cryptographic
// hash function on any input before handing it to this function.
func Sign(msg []byte, seckey []byte) ([]byte, error) {
	// ::FIXME:: it is not implemented
	return nil, fmt.Errorf("Not implemented")
}

// RecoverPubkey returns the public key of the signer.
// msg must be the 32-byte hash of the message to be signed.
// sig must be a 65-byte compact ECDSA signature containing the
// recovery id as the last element.
func RecoverPubkey(msg []byte, sig []byte) ([]byte, error) {
	// ::FIXME:: it is not implemented
	return nil, fmt.Errorf("Not implemented")
}

// VerifySignature checks that the given pubkey created signature over message.
// The signature should be in [R || S] format.
func VerifySignature(pubkey, msg, signature []byte) bool {
	// ::FIXME:: it is not implemented
	return false
}

// DecompressPubkey parses a public key in the 33-byte compressed format.
// It returns non-nil coordinates if the public key is valid.
func DecompressPubkey(pubkey []byte) (x, y *big.Int) {
	// ::FIXME:: it is not implemented
	return nil, nil
}

// CompressPubkey encodes a public key to 33-byte compressed format.
func CompressPubkey(x, y *big.Int) []byte{
	// ::FIXME:: it is not implemented
	panic("libsecp256k1 error")
}

func checkSignature(sig []byte) error {
	// ::FIXME:: it is not implemented
	return fmt.Errorf("Not implemented")
}
