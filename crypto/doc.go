// This package provdes crypto-functions for catgi.
//
// It provides a secure and fast HMAC function using Blake2b
//
// It provides a secure and deterministic key derivation mechanism using
// keypaths and Blake2b.
//
// It provides a secure and easy symmetric encryption using Polycha20 and
// the key derivatino mechanism. Each encryption gets it's own key
// thusly preventing nonce reuse for multiple calls to the encryption
// method.
package crypto
