# Lessons Learned from doing Crypto in Go

## Comparing (H)MACs

In most languages, the naive approach to checking a MAC would
be to write something like this:

```go 
func verifymac(mac, key, data []byte) error {
    computedMac := calculateMac(key, data)
    if mac != computedMac {
        return errors.New("MAC Fail")
    }
    return nil
}
```

If you do it this way, it's going to blow up in your face the moment
somebody tries even trivial timing attacks.

The first step is to swap the naive comparison with a constant-time
compare.

```go
func verifymac(mac, key, data []byte) bool {
    computedMac := calculateMac(key, data)
    var result byte
    for k := range mac {
        result |= mac[k] ^ computedMac[k]
    }
    if result != 0 {
        return errors.New("MAC Fail")
    }
    return nil
}
```

But this opens a new attack vector. If the keys are of a different size
an attacker could exploit this to reintroduce timing attacks.

So we have to pad or trim one of the macs to the correct size.

```go
func verifymac(mac, key, data []byte) bool {
    computedMac := calculateMac(key, data)
    if len(computedMac) != len(mac) {
        mac = padmac(mac, len(computedMac))
    }
    var result byte
    for k := range mac {
        result |= mac[k] ^ computedMac[k]
    }
    if result != 0 {
        return errors.New("MAC Fail")
    }
    return nil
}
```
The function `padmac` is defined as follows:

* if the length of the mac is smaller than the given size, append missing bytes
* if the length of the mac is bigger than the given size, trim to target size

It's very unlikely that an attacker somehow manages to generate a valid MAC
this way since it'll either be trimmed, in which case the attacker wastes computing
power or it will be padded, in which case the attacker faces a difficulty curve
that so far has secured the Bitcoin network; finding hashes with lots of 0s
is hard.

So now we must be safe, right?

Not quite yet. There are still two timing attacks in this function.

The first is the allocation of `errors.New()` and the other is the `if{}`.

Let's take care of `errors.New()` first.

```go
var macError = errors.New("MAC Fail")
func verifymac(mac, key, data []byte) bool {
    computedMac := calculateMac(key, data)
    if len(computedMac) != len(mac) {
        mac = padmac(mac, len(computedMac))
    }
    var result byte
    for k := range mac {
        result |= mac[k] ^ computedMac[k]
    }
    if result != 0 {
        return macError
    }
    return nil
}
```

This avoids an allocation on the stack before returning and brings us
closer to constant time operation.

Now to the `if{}`; the problem here is that it takes longer for the
function to return from a success-comparison than it takes to
return from an non-success.

While there are no timing attacks I'm aware of that could lead to problems
with this directly, it can under certain circumstances allow an attacker
to verify if an HMAC is correct even if your server never returns the error
to the client.

So I'd categorize it under "not problematic normally but let's fix it".

```go
var macError = errors.New("MAC Fail")
func verifymac(mac, key, data []byte) bool {
    var returnError error
    computedMac := calculateMac(key, data)
    if len(computedMac) != len(mac) {
        mac = padmac(mac, len(computedMac))
    }
    var result byte
    for k := range mac {
        result |= mac[k] ^ computedMac[k]
    }
    if result != 0 {
        returnError = macError
    }
    return returnError
}
```

The return of the function is now always at the same place. The time it
takes to assign `macError` to `returnError` is much to small to *really*
matter over a networked connection.

You *can* implement the following:

```go
var macError = errors.New("MAC Fail")
var noMacError = errors.New("No MAC Fail")
func verifymac(mac, key, data []byte) bool {
    var returnError error
    computedMac := calculateMac(key, data)
    if len(computedMac) != len(mac) {
        mac = padmac(mac, len(computedMac))
    }
    var result byte
    for k := range mac {
        result |= mac[k] ^ computedMac[k]
    }
    if result != 0 {
        returnError = macError
    } else {
        returnError = noMacError
    }
    return returnError
}
```

This function will now take almost exactly as long if it fails as it
will if it doesn't and it's irrelevant which parts of the key are wrong
or even if the key has the wrong size.

The final function I use goes as follows:

```go
func VerifyHMAC(hmac, key []byte, data io.Reader) (err error) {
	var verifyHMAC = make([]byte, len(hmac))
	verifyHMAC, err = HMAC(key, data)

	lenhmac := len(hmac)
	lenvmac := len(verifyHMAC)
	if lenhmac != lenvmac {
		if lenhmac > lenvmac {
			hmac = hmac[:lenvmac]
		} else {
			hmac = append(hmac,
				make([]byte, lenvmac-lenhmac)...)
		}
	}

	var result byte
	for k := range hmac {
		result |= hmac[k] ^ verifyHMAC[k]
	}
	var retErr error
	if result != 0 {
		retErr = HMACVerificationFail
	}
	if lenhmac != lenvmac {
		retErr = HMACVerificationFail
	}
	if err != nil {
		retErr = HMACVerificationFail
	}

	return retErr
}
```

This function has more error conditions, which means I had to give 
up the last optimization in favor of function simplicity.

### Benchmark Results

The following Benchmark results use the last function that does not employ
the `if{}else{}`.

It uses three function, HMAC, Verify and VerifyWrong that test the speed
of creating a HMAC, verifying a correct HMAC and Verifying a Wrong HMAC.

Three Levels are employed; a KeySize level which uses 64 bytes of data,
a 4K level which uses 4KiB of data and a 64M level which uses 64MiB of data.

The first level tests the core of the function, namely the resources needed
for an almost minimal setup and verification.

The second level is meant to simulate a typical file or data verification.

The third level is meant to simulate a large file or data transfer, like
uploading a youtube video. At 64MiB the setup of the function is the 
least significant part and most time will be spent generating the HMAC.

The benchmark was performed on a i5 2500 at somewhat normal frequency and some 
DDR3-1333 nonECC RAM. The total performance should be slightly above average
for today's standards.

```
```

## One Secret to Rule Them All 

As of writing this document, Catgi requires using several secrets
for each component that uses secret keys or data.

To make this easier for the user and me alike, I devised the keyman.go
file.

It defines a type, SecretKey, that can be used to generate arbitrary
subsecrets and pull an arbitrary amount of secret data using HKDF from
each.

The path notation of the secret key for documentation seperates the
strings with a `/` and surrounds it with `"` if it contains `/` or
whitespace. 

Examples:

```
master/internal/cookie_secret
master/backends/encrypt/"first slot"
master/backends/fcache/localfs/encrypt
```

Each backend or use can therefore use their own key for their own
purposes and derive subkeys if necessary that they can pass down
to other uses without compromising their secret.

## Symmetric Block Encryption

This is a collection of notes on how I'm implementing the Symmetric
Encryption.

The first problem is that I don't want to encrypt inplace, I'd rather
move towards using streams to reduce the memory usage of Catgi.

As a threat model I've chosen to go a bit simple; an attacker who has full
man-in-the-middle control over the data stream should only be able to make
decryption fail with an error.

My implementation of Encryption is planned to produce blocks looking like
this:

```
type BlockLength uint64 // [4]byte
type EncryptedBlock struct {
    Cipher  uint16   `msgpack:"cipher"`
    Padding uint16   `msgpack:"padding"`
    Nonce   [12]byte `msgpack:"nonce"`
    Salt    []byte   `msgpack:"salt"`
    Data    []byte   `msgpack:"data"`
}
```

Cipher is a constant that indicates which encryption was used.

Nonce is the nonce used by the encryption.

Salt is a key salt added when starting a new encryption. This prevents
key reuse, as each encryption uses a different key, which also reduces
the probability of nonce reuse. (2^(8*24) is a lot but this is safer)

Data is the ciphertext.

BlockLength is a header byte that encodes the size of the following block.
This way the decryption has it easier to decode the blocks.

For Streaming, the Block size may vary so I'll have to use several methods
of encrypting:

* Whole, which encrypts everything at once
* Stream, which encrypts as long it can
* Block, which works like stream but always produces fixed size blocks

This mode is not encoded in the ciphertext and must be determined by the
application.

The Go API does not expose a proper streaming AEAD interface, so all 
ciphers use the same block-based encryption underneath.

All blocks operate independently, you can decode any part of a stream
provided you have the key the same as block ciphering.

The block cipher will only pad the data it encrypts, it makes no guarantees
that the encoded blocks are all of the same size (ie best effort mode).

### Size Limits

In Block Mode the amount of padding cannot exceed 64KiB.

Since blocks are 4KiB, this should not be a problem.

Additionally, Block Mode is limited to 2^63-1 total block size.

In Stream Mode, the padding is not used and the block size is up to 2^64-1.

In Whole mode, you may encrypt up to 2^64-1 bytes of data, not accounting for encoding
overhead.

### notes

An attacker could corrupt the ciphertext by changing the BlockSizeHeader,
which would cause the application to read the wrong amount of data
from the wire.

However, this attack vector leads only to the application not decrypting
the data, which an attacker could do by flipping any byte in the block,
so I don't see it as a problem tbh.