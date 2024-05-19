// Package designed with functions to aid in creating, packaging, and unpackaging CipherCord messages.
package message

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// Advanced Encryption Standard (256-bit) / Galois/Counter Mode / Base64 (RAW)
const EncryptionType string = "aes-256/gcm/b64r"

// Secure Hash Algorithm (256-bit) / Base64 (RAW) / Cut 32
const HashingType string = "sha256/b64r/:32"

// MessagePack / Base64 (RAW)
const PackagingType string = "msgpack/b64r"

// The major API version number.
const Version = "2"

// UnencryptedMessage represents a package of unencrypted information that will later be encrypted.
// Nothing in this struct will ever be sent over the wire.
//
// The * indicates the field will be seen as hashed when decrypted.
type UnencryptedMessage struct {
	Key      string // Secret password in plain text
	Room     string // *Room name
	FileType string // File type of Content. Empty when Content is a message and not a file.
	Content  string // Message/file content
	Author   string // Author's nickname
}

// A package of encrypted data that is ready to be sent out in the world.
type EncryptedMessage struct {
	Key        string // Hash of key32.
	Version    string // Unencrypted API version.
	Encryption string // Unencrypted encryption type.
	Hashing    string // Unencrypted hashing type.
	Packaging  string // Unencrypted packaging type.
	Room       string // Hash of room name.
	FileType   string // Unencrypted file type. Empty when Content is a message and not a file.
	Content    string // Encrypted message/file content.
	Author     string // Encrypted nickname of author.
}

// Converts an UnencryptedMessage into an EncryptedMessage.
func EncryptMessage(umsg UnencryptedMessage) (EncryptedMessage, error) {
	var emsg EncryptedMessage

	key32 := Hash32(umsg.Key)
	keyHash := Hash32(key32)

	emsg.Key = keyHash

	emsg.Version = Version
	emsg.Encryption = EncryptionType
	emsg.Hashing = HashingType
	emsg.Packaging = PackagingType
	emsg.Room = Hash32(umsg.Room)

	content, err := Encrypt(umsg.Content, key32)
	if err != nil {
		return EncryptedMessage{}, err
	}
	emsg.Content = content

	author, err := Encrypt(umsg.Author, key32)
	if err != nil {
		return EncryptedMessage{}, err
	}
	emsg.Author = author

	return emsg, nil
}

// Not a serious error message, this just means you have the wrong key. Probably user error.
var ErrKeyUnmatched error = fmt.Errorf("ciphercord: one or more unmatched fields")

// Usually means message was packaged with a different client.
var ErrOtherUnmatched error = fmt.Errorf("ciphercord: one or more unmatched fields")

// Converts an EncryptedMessage into an UnencryptedMessage.
func DecryptMessage(emsg EncryptedMessage, key string) (UnencryptedMessage, error) {
	key32 := Hash32(key)
	keyHash := Hash32(key32)

	if emsg.Key != keyHash {
		return UnencryptedMessage{}, ErrKeyUnmatched
	}
	if emsg.Encryption != EncryptionType || emsg.Hashing != HashingType || emsg.Packaging != PackagingType || emsg.Version == Version {
		return UnencryptedMessage{}, ErrOtherUnmatched
	}

	var umsg UnencryptedMessage

	umsg.Room = emsg.Room

	content, err := Decrypt(emsg.Content, key32)
	if err != nil {
		return UnencryptedMessage{}, err
	}
	umsg.Content = content

	author, err := Decrypt(emsg.Author, key32)
	if err != nil {
		return UnencryptedMessage{}, err
	}
	umsg.Author = author

	return umsg, nil
}

// The same as DecryptMessage but it assumes everything matches up.
// It doesn't check to see if encryption types, hashing types, or keys match up.
// Will not return ErrUnmatched.
func DecryptMessageUnstable(emsg EncryptedMessage, key string) (UnencryptedMessage, error) {
	key32 := Hash32(key)

	var umsg UnencryptedMessage

	umsg.Room = emsg.Room

	content, err := Decrypt(emsg.Content, key32)
	if err != nil {
		return UnencryptedMessage{}, err
	}
	umsg.Content = content

	author, err := Decrypt(emsg.Author, key32)
	if err != nil {
		return UnencryptedMessage{}, err
	}
	umsg.Author = author

	return umsg, nil
}

// Encodes an EncryptedMessage into a plain text string.
func Encode(msg EncryptedMessage) (string, error) {
	b, err := msgpack.Marshal(msg)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(b), nil
}

// Decodes a plain text string back into an EncryptedMessage.
func Decode(s string) (EncryptedMessage, error) {
	b, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return EncryptedMessage{}, err
	}

	var emsg EncryptedMessage

	err = msgpack.Unmarshal(b, &emsg)
	if err != nil {
		return EncryptedMessage{}, err
	}

	return emsg, nil
}

// Packages up an UnencryptedMessage to string to be ready for sending.
func Package(umsg UnencryptedMessage) (string, error) {
	emsg, err := EncryptMessage(umsg)
	if err != nil {
		return "", err
	}

	encoded, err := Encode(emsg)
	if err != nil {
		return "", err
	}

	return encoded, nil
}

// Unpackages a string to UnencryptedMessage be ready for parsing.
func Unpackage(s string, key string) (UnencryptedMessage, error) {
	emsg, err := Decode(s)
	if err != nil {
		return UnencryptedMessage{}, err
	}

	umsg, err := DecryptMessage(emsg, key)
	if err != nil {
		return UnencryptedMessage{}, err
	}

	return umsg, nil
}

// Encrypts s into an encrypted string. Argument key32 means a key of 32 characters.
func Encrypt(s string, key32 string) (string, error) {
	c, err := aes.NewCipher([]byte(key32))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nonce[:], nonce[:], []byte(s), nil)
	return base64.RawStdEncoding.EncodeToString(encrypted), nil
}

// Cipher text is smaller than the nonce size.
var ErrTooSmall = fmt.Errorf("ciphercord: cipher text is smaller than the nonce size")

// Decrypts s into plain text. Argument key32 means a key of 32 characters.
func Decrypt(s string, key32 string) (string, error) {
	cipherBytes, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher([]byte(key32))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()

	if len(cipherBytes) < nonceSize {
		return "", ErrTooSmall
	}

	nonce, cipherBytes := cipherBytes[:nonceSize], cipherBytes[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// Takes string and hashes it to be 32 characters. This is how the other functions convert key to key32
func Hash32(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hash := h.Sum(nil)
	return base64.RawStdEncoding.EncodeToString(hash)[:32]
}
