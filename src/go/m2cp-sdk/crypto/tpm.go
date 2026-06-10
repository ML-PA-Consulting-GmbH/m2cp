package crypto

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"github.com/google/uuid"
)

var (
	pathTpmDevice = "/dev/tpmrm0"
	tpmLock       sync.Mutex
)

const tpmSigningKeyHandle = tpmutil.Handle(0x81010200)

var tpmSigningKeyTemplate = tpm2.Public{
	Type:       tpm2.AlgRSA,
	NameAlg:    tpm2.AlgSHA256,
	Attributes: tpm2.FlagSignerDefault,
	RSAParameters: &tpm2.RSAParams{
		Sign: &tpm2.SigScheme{
			Alg:  tpm2.AlgRSASSA,
			Hash: tpm2.AlgSHA256,
		},
		KeyBits: 2048,
	},
}

const tpmEncryptKeyHandle = tpmutil.Handle(0x81010201) // fixed encryption key handle
var tpmEncryptKeyTemplate = tpm2.Public{
	Type:       tpm2.AlgRSA,
	NameAlg:    tpm2.AlgSHA256,
	Attributes: tpm2.FlagDecrypt | tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth,
	RSAParameters: &tpm2.RSAParams{
		KeyBits: 2048,
	},
}

const tpmStorageKeyHandle = tpmutil.Handle(0x81010202) // fixed storage key handle
var tpmStorageKeyTemplate = tpm2.Public{
	Type:       tpm2.AlgRSA,
	NameAlg:    tpm2.AlgSHA256,
	Attributes: tpm2.FlagStorageDefault,
	RSAParameters: &tpm2.RSAParams{
		Symmetric: &tpm2.SymScheme{
			Alg:     tpm2.AlgAES,
			KeyBits: 128,
			Mode:    tpm2.AlgCFB,
		},
		KeyBits: 2048,
	},
}

const symKeySize = 32
const tpmSealedSymKeyHandle = tpmutil.Handle(0x81010203) //fixed sym key handle
var sealedTemplate = tpm2.Public{
	Type:       tpm2.AlgKeyedHash,
	NameAlg:    tpm2.AlgSHA256,
	Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagUserWithAuth,
	AuthPolicy: nil,
}

// decodeAttributes was invented for debugging only
func decodeAttributes(attr tpm2.KeyProp) string {
	type Flag struct {
		Bitmask tpm2.KeyProp
		Name    string
	}

	flags := []Flag{
		{tpm2.FlagFixedTPM, "FlagFixedTPM"},
		{tpm2.FlagStClear, "FlagStClear"},
		{tpm2.FlagFixedParent, "FlagFixedParent"},
		{tpm2.FlagSensitiveDataOrigin, "FlagSensitiveDataOrigin"},
		{tpm2.FlagUserWithAuth, "FlagUserWithAuth"},
		{tpm2.FlagAdminWithPolicy, "FlagAdminWithPolicy"},
		{tpm2.FlagNoDA, "FlagNoDA"},
		{tpm2.FlagRestricted, "FlagRestricted"},
		{tpm2.FlagDecrypt, "FlagDecrypt"},
		{tpm2.FlagSign, "FlagSign"},
	}
	result := ""
	for _, flag := range flags {
		if attr&flag.Bitmask != 0 {
			result += fmt.Sprintf("|%s", flag.Name)
		}
	}
	return result
}

// prettyPrintPublicKey was invented for debugging only
func prettyPrintPublicKey(key tpm2.Public) string {
	data := map[string]interface{}{
		"Type":       key.Type,
		"NameAlg":    key.NameAlg,
		"Attributes": decodeAttributes(key.Attributes),
		"AuthPolicy": key.AuthPolicy,
	}
	jsonBytes, _ := json.MarshalIndent(data, "", "    ")

	return string(jsonBytes)
}

// TpmSelectDevice allows using a non-standard device for tpm operations
func TpmSelectDevice(device string) error {
	info, err := os.Stat(device)
	if os.IsNotExist(err) {
		return fmt.Errorf("device path does not exist")
	} else if err != nil {
		return fmt.Errorf("failed to access the device path: %s", err)
	}

	if info.IsDir() {
		return fmt.Errorf("device path points to a directory, not a device")
	}

	if (info.Mode() & os.ModeDevice) == 0 {
		return fmt.Errorf("device path is not a valid device")
	}

	pathTpmDevice = device
	return nil
}

// TpmSignBytes signs the given bytes using the m2cp key (derived from EK).
func TpmSignBytes(toSign []byte) (signature []byte, err error) {
	var sig []byte

	if err = withTpm(tpmSigningKeyHandle, tpmSigningKeyTemplate, func(key *client.Key) error {
		fmt.Printf("signing %d bytes with key %s, payload (base64 encoded):\n%s\n",
			len(toSign),
			rsaPublicKey(key.PublicKey().(*rsa.PublicKey)).ID(),
			base64.StdEncoding.EncodeToString(toSign),
		)

		digest := tpmHashBytes(toSign)
		fmt.Printf("Step 1: hashed the payload with SHA3_384. digest base64 encoded:\n%s\n", base64.StdEncoding.EncodeToString(digest))

		fmt.Printf("Step 2: sign the digest (hash again with sha2_256 and encrypt with RSA key)\n")
		sig, err = key.SignData(digest)
		if err != nil {
			return fmt.Errorf("failed to sign: %s", err)
		}

		fmt.Printf("Step 3: created signature, base64 encoded:\n%s\n", base64.StdEncoding.EncodeToString(sig))

		if !tpmVerifyEkSignature(key.PublicKey(), toSign, sig) {
			return fmt.Errorf("signature verification failed")
		}

		id := rsaPublicKey(key.PublicKey().(*rsa.PublicKey)).ID()
		fmt.Printf("signed message with key %s\n", id)

		return nil
	}); err != nil {
		return nil, err
	}

	return sig, nil
}

// TpmDecryptRSA decrypts the given encrypted bytes asymmetrically using the m2cp encryption key.
func TpmDecryptRSA(encrypted []byte) (decrypted []byte, err error) {
	var plaintext []byte

	// ensure key exists and get public key for logging purposes
	pub, err := TpmGetEncryptionPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption public key: %s", err)
	}

	fmt.Printf("decrypting %d bytes with key %s, encrypted data (base64 encoded):\n%s\n",
		len(encrypted),
		pub.ID(),
		base64.StdEncoding.EncodeToString(encrypted),
	)

	tpmLock.Lock()
	defer tpmLock.Unlock()

	// Use tpm2.RSADecrypt with OAEP scheme
	rwc, err := openTpm()
	if err != nil {
		return nil, fmt.Errorf("failed to open TPM for decryption: %s", err)
	}
	defer rwc.Close()

	scheme := &tpm2.AsymScheme{
		Alg:  tpm2.AlgOAEP,
		Hash: tpm2.AlgSHA256,
	}

	plaintext, err = tpm2.RSADecrypt(rwc, tpmEncryptKeyHandle, "", encrypted, scheme, "")
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %s", err)
	}

	fmt.Printf("decrypted %d bytes with key %s\n", len(plaintext), pub.ID())

	return plaintext, nil
}

// TpmGetEndorsementPublicKey returns the public key of m2cp key (derived from EK).
func TpmGetEndorsementPublicKey() (PublicKey, error) {
	var pub PublicKey

	if err := withTpm(tpmSigningKeyHandle, tpmSigningKeyTemplate, func(key *client.Key) error {
		pub = rsaPublicKey(key.PublicKey().(*rsa.PublicKey))
		return nil
	}); err != nil {
		return nil, err
	}
	return pub, nil
}

func TpmGetEndorsementPublicKeyBase64() (string, error) {
	ekPub, err := TpmGetEndorsementPublicKey()
	if err != nil {
		return "", err
	}
	return encodeKeyBase64(ekPub)
}

// TpmGetEncryptionPublicKey returns the public key for encryption operations.
func TpmGetEncryptionPublicKey() (PublicKey, error) {
	var pub PublicKey

	if err := withTpm(tpmEncryptKeyHandle, tpmEncryptKeyTemplate, func(key *client.Key) error {
		pub = rsaPublicKey(key.PublicKey().(*rsa.PublicKey))
		return nil
	}); err != nil {
		return nil, err
	}
	return pub, nil
}

func TpmGetEncryptionPublicKeyBase64() (string, error) {
	encPub, err := TpmGetEncryptionPublicKey()
	if err != nil {
		return "", err
	}
	return encodeKeyBase64(encPub)
}

// TpmGetEncryptionPublicKeyPEM returns the encryption public key in PEM format.
// This format can be used with OpenSSL for RSA-OAEP encryption.
func TpmGetEncryptionPublicKeyPEM() (string, error) {
	var rsaPub *rsa.PublicKey

	if err := withTpm(tpmEncryptKeyHandle, tpmEncryptKeyTemplate, func(key *client.Key) error {
		rsaPub = key.PublicKey().(*rsa.PublicKey)
		return nil
	}); err != nil {
		return "", err
	}

	derBytes, err := x509.MarshalPKIXPublicKey(rsaPub)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %s", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}

	return string(pem.EncodeToMemory(pemBlock)), nil
}

func CreateNewCashedSymKey() (err error) {
	_, _, err = newCashedSymKey()
	return err
}

// newCashedSymKey gets or creates a sealed symmetric key at the specified handle
func newCashedSymKey() (symKey []byte, new bool, err error) {
	var unsealErr error

	defer func() {
		if err != nil && unsealErr != nil {
			err = fmt.Errorf("%w\n - unseal error: (%s)", err, unsealErr)
		}
	}()

	// Lock first to ensure atomic operation with the same TPM context
	tpmLock.Lock()
	defer tpmLock.Unlock()

	rwc, err := openTpm()
	if err != nil {
		err = fmt.Errorf("failed to open TPM: %s", err)
		return
	}
	defer rwc.Close()

	// Ensure the storage primary key exists in THIS TPM context
	// Use client.LoadCachedKey to just load it, or NewCachedKey to create if needed
	storageKey, err := client.NewCachedKey(rwc, tpm2.HandleOwner, tpmStorageKeyTemplate, tpmStorageKeyHandle)
	if err != nil {
		err = fmt.Errorf("failed to get storage key: %s", err)
		return
	}
	defer storageKey.Close()

	symKey, unsealErr = tpm2.Unseal(rwc, tpmSealedSymKeyHandle, "")
	if unsealErr == nil {
		return symKey, false, nil
	}

	// If unsealing failed, we need to create and seal the symmetric key
	// This should only happen on the first run after provisioning or if the TPM was cleared or the keys were messed with

	// Generate random symmetric key
	symKey = make([]byte, symKeySize)
	_, err = rand.Read(symKey)
	if err != nil {
		err = fmt.Errorf("failed to generate symmetric key: %s", err)
		return
	}

	// Seal the symmetric key under the storage primary using CreateKeyWithSensitive
	// Use the loaded storage key handle to ensure it's in the current TPM context
	private, public, _, _, _, err := tpm2.CreateKeyWithSensitive(rwc, storageKey.Handle(), tpm2.PCRSelection{}, "", "", sealedTemplate, symKey)
	if err != nil {
		err = fmt.Errorf("failed to seal symmetric key: %s", err)
		return
	}

	// Load the sealed key object into TPM memory
	// Note: Exponential backoff is less necessary now that storage key is in same context
	var handle tpmutil.Handle
	maxRetries := 4
	for i := 0; i < maxRetries; i++ {
		// Wait before retry: exponential backoff (10ms, 100ms, 1s)
		if i > 0 {
			time.Sleep(time.Duration(math.Pow10(i)) * 10 * time.Millisecond)
		}
		handle, _, err = tpm2.Load(rwc, storageKey.Handle(), "", public, private)
		if err == nil {
			break
		}

		// If it's the last retry, return the error
		if i == maxRetries-1 {
			err = fmt.Errorf("failed to load sealed object after %d retries: %s", maxRetries, err)
			return
		}
	}
	defer tpm2.FlushContext(rwc, handle)

	// Persist the sealed object to the persistent handle
	err = tpm2.EvictControl(rwc, "", tpm2.HandleOwner, handle, tpmSealedSymKeyHandle)
	if err != nil {
		err = fmt.Errorf("failed to persist sealed key: %s", err)
		return
	}

	return symKey, true, nil
}

// TpmEncryptSymmetric encrypts the given bytes using the m2cp symmetric storage key.
func TpmEncryptSymmetric(plaintext []byte) ([]byte, error) {
	// Get or create the sealed symmetric key
	key, _, err := newCashedSymKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get symmetric key: %s", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %s", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %s", err)
	}

	// Generate IV
	nonce := make([]byte, gcm.NonceSize()) // Typically 12 bytes for GCM
	_, err = rand.Read(nonce)              // Generate random nonce
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %s", err)
	}

	// Encrypt and prepend IV to the ciphertext
	var ciphertext []byte
	ciphertext = gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// custom error for decrypting with a new key
type NewKeyError struct {
	Message string
}

func (e *NewKeyError) Error() string {
	return e.Message
}

func TpmDecryptSymmetric(encrypted []byte) ([]byte, error) {
	// Get or create the sealed symmetric key
	key, new, err := newCashedSymKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get symmetric key: %s", err)
	}
	if new {
		return nil, &NewKeyError{Message: "a new symmetric key was created, cannot decrypt data encrypted with a different key"}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %s", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %s", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext (nonce is prepended )
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %s", err)
	}

	return plaintext, nil
}

func seedToSerial(seed string) string {
	var deviceUUID uuid.UUID
	namespace := uuid.NewSHA1(uuid.NameSpaceOID, []byte("m2cp-device-serial"))
	deviceUUID = uuid.NewSHA1(namespace, []byte(seed))
	return deviceUUID.String()
}

func DeterministicDeviceSerial() (string, error) {
	if HasTpm() {
		return TpmDeterministicDeviceSerial()
	}
	return NoTpmDeterministicDeviceSerial()
}

func NoTpmDeterministicDeviceSerial() (string, error) {
	for _, device := range []string{"/dev/mmcblk0", "/dev/mmcblk1", "/dev/mmcblk2"} {
		if serial, err := getDriveSerial(device); err == nil {
			return seedToSerial(serial), nil
		}
	}
	return "", fmt.Errorf("failed building deterministic serial - no seeding sources found")
}

// TpmDeterministicDeviceSerial generates a deterministic serial number for the device from the m2cp key (derived from EK).
// If TPM is not available, it generates a deterministic serial number using other immutable device properties.
func TpmDeterministicDeviceSerial() (string, error) {

	var seed string = ""
	if err := withTpm(tpmSigningKeyHandle, tpmSigningKeyTemplate, func(key *client.Key) error {
		keyId := rsaPublicKey(key.PublicKey().(*rsa.PublicKey)).ID()
		seed = keyId
		return nil
	}); err != nil {
		return "", err
	}

	return seedToSerial(seed), nil
}

var counter = 0

func TpmTest() {
	message := []byte(fmt.Sprintf("hello world #%d", counter))
	counter++
	fmt.Printf("Signing '%s'..\n", string(message))

	signature, err := TpmSignBytes(message)
	if err != nil {
		fmt.Printf("failed: %s\n", err)
	} else {
		fmt.Printf("\nsuccess\n%s\n", base64.StdEncoding.EncodeToString(signature))
	}
	fmt.Println()
}

func HasTpm() bool {
	_, err := os.Stat(pathTpmDevice)
	return err == nil
}

func withTpm(handle tpmutil.Handle, template tpm2.Public, f func(key *client.Key) error) error {
	const maxTries = 3
	const waitBetweenTriesS = 1

	for try := 0; try < maxTries; try++ {
		err := withTpmTry(handle, template, f)
		if err != nil {
			time.Sleep(time.Duration(waitBetweenTriesS) * time.Second)
			continue
		} else {
			return nil
		}
	}

	return fmt.Errorf("failed to perform TPM operation after %d tries", maxTries)
}

func withTpmTry(handle tpmutil.Handle, template tpm2.Public, f func(key *client.Key) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	tpmLock.Lock()
	defer tpmLock.Unlock()

	rwc, err := openTpm()
	if err != nil {
		return fmt.Errorf("failed to open TPM: %s", err)
	}
	defer func() {
		_ = rwc.Close()
	}()

	key, err := client.NewCachedKey(rwc, tpm2.HandleEndorsement, template, handle)
	if err != nil {
		return fmt.Errorf("failed to create key: %s", err)
	}
	defer key.Close()

	err = f(key)
	return err
}

func tpmVerifyEkSignature(pubKey crypto.PublicKey, message, signature []byte) bool {
	message = tpmHashBytes(message)

	hashAlgo := crypto.SHA256

	hash := hashAlgo.New()
	hash.Write(message)
	digest := hash.Sum(nil)

	return rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), hashAlgo, digest, signature) == nil

}

func tpmHashBytes(toHash []byte) []byte {
	hash := crypto.SHA3_384.New()
	hash.Write(toHash)
	return hash.Sum(nil)
}

func getDriveSerial(drive string) (serial string, err error) {
	cmd := exec.Command("udevadm", "info", "--query=all", "--name="+drive)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err = cmd.Run(); err != nil {
		return
	}

	model := ""

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ID_SERIAL=") {
			serial = strings.Split(line, "=")[1]
		} else if strings.Contains(line, "ID_NAME=") {
			model = strings.Split(line, "=")[1]
		}
	}

	// this is a list of known devices that can be used for TPM ... more can be added as needed
	if model != "Q2J55L" && model != "DG4008" && model != "JB1Q5" {
		return "", fmt.Errorf("tpm:getDriveSerial() ID_NAME -> unknown block device model '%s'", model)
	}

	if serial == "" {
		return "", fmt.Errorf("tpm:getDriveSerial() ID_SERIAL not found")
	}

	return serial, nil
}

func getMacAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var macAddresses []string
	for _, i := range interfaces {
		if i.HardwareAddr != nil {
			macAddresses = append(macAddresses, i.HardwareAddr.String())
		}
	}

	return macAddresses, nil
}
