package remotecache

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"hash"
)

// SignatureSize is the length of the signature that starts every object.
const SignatureSize = sha256.Size

// ErrTampered is an object whose signature does not match its content.
var ErrTampered = errors.New("the remote cache object's signature does not match: it was not written by this team's key, or was changed")

// signer computes an object's signature: HMAC-SHA256 over its name (kind and
// key, so an object cannot be replayed under another key) and its content.
func signer(secret []byte, kind, key string) hash.Hash {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(kind + "/" + key + "\n"))
	return mac
}

// Sign returns the object to store: the signature, then the content.
func Sign(secret []byte, kind, key string, content []byte) []byte {
	mac := signer(secret, kind, key)
	mac.Write(content)
	return append(mac.Sum(nil), content...)
}

// Verify checks a stored object and returns its content.
func Verify(secret []byte, kind, key string, object []byte) ([]byte, error) {
	if len(object) < SignatureSize {
		return nil, ErrTampered
	}
	content := object[SignatureSize:]
	mac := signer(secret, kind, key)
	mac.Write(content)
	if !hmac.Equal(mac.Sum(nil), object[:SignatureSize]) {
		return nil, ErrTampered
	}
	return content, nil
}
