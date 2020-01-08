package stboot

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

type signature struct {
	Bytes []byte
	Cert  *x509.Certificate
}

// Signer is used by BootBall to hash, sign and varify the BootConfigs
// with appropriate algorithms
type Signer interface {
	Hash(files ...string) ([]byte, error)
	Sign(privKey string, data []byte) ([]byte, error)
	Verify(sig signature, hash []byte) error
}

// DummySigner creates signatures that are always valid.
type DummySigner struct{}

// Hash hashes the the provided files. I case of DummySigner
// just 8 random bytes are returned.
func (DummySigner) Hash(files ...string) ([]byte, error) {
	hash := make([]byte, 8)
	_, err := rand.Read(hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// Sign signes the provided data with privKey. In case of DummySigner
// just 8 random bytes are returned
func (DummySigner) Sign(privKey string, data []byte) ([]byte, error) {
	sig := make([]byte, 8)
	_, err := rand.Read(sig)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// Verify checks if sig contains a valid signature of hash. In case of
// DummySigner this is allwazs the case.
func (DummySigner) Verify(sig signature, hash []byte) error {
	return nil
}

// Sha512PssSigner uses SHA512 hashes ans PSS signatures along with
// x509 certificates.
type Sha512PssSigner struct{}

// Hash hashes the the provided files. I case of Sha512PssSigner
// it is a SHA512 hash.
func (Sha512PssSigner) Hash(files ...string) ([]byte, error) {
	h := sha512.New()
	h.Reset()

	for _, file := range files {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		h.Write(buf)
	}
	return h.Sum(nil), nil
}

// Sign signes the provided data with privKey. In case of Sha512PssSigner
// it is a PSS signature.
func (Sha512PssSigner) Sign(privKey string, data []byte) ([]byte, error) {
	buf, err := ioutil.ReadFile(privKey)
	if err != nil {
		return nil, err
	}

	privPem, _ := pem.Decode(buf)
	key, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, err
	}
	if key == nil {
		err = fmt.Errorf("key is empty")
		return nil, err
	}

	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}

	sig, err := rsa.SignPSS(rand.Reader, key, crypto.SHA512, data, opts)
	if err != nil {
		return nil, err
	}
	if sig == nil {
		return nil, fmt.Errorf("signature is nil")
	}
	return sig, nil
}

// Verify checks if sig contains a valid signature of hash. In case of
func (Sha512PssSigner) Verify(sig signature, hash []byte) error {
	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}
	err := rsa.VerifyPSS(sig.Cert.PublicKey.(*rsa.PublicKey), crypto.SHA512, hash, sig.Bytes, opts)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}
	return nil
}

// parseCertificate parses certificate from raw certificate
func parseCertificate(raw []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(raw)

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// certPool returns a x509 certificate pool from raw certificate
func certPool(pem []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(pem)
	if !ok {
		return nil, errors.New("Failed to parse root certificate")
	}
	return certPool, nil
}

// validateCertificate validates cert against certPool. If cert is not signed
// by a certificate of certPool an error is returned.
func validateCertificate(cert *x509.Certificate, cerPool *x509.CertPool) error {
	opts := x509.VerifyOptions{
		Roots: cerPool,
	}
	_, err := cert.Verify(opts)
	if err != nil {
		return err
	}
	return nil
}
