package auth

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"os"
)

// Credentials represents the stored credentials that seaway uses to authenticate
// with it's dependent resources.  This is super temporary and will be replaced
// with a more secure solution in the future.
type Credentials struct {
	AccessKey         string `json:"access_key"`
	SecretKey         string `json:"secret_key"`
	MinioRootUser     string `json:"minioRootUser"`
	MinioRootPassword string `json:"minioRootPassword"`
}

// NewCredentials creates a new set of credentials for seaway to use.
func NewCredentials(filename string) (*Credentials, error) {
	secretKey, err := generateRandomString(24)
	if err != nil {
		return nil, err
	}

	rootPassword, err := generateRandomString(24)
	if err != nil {
		return nil, err
	}

	creds := &Credentials{
		AccessKey:         "seaway",
		SecretKey:         secretKey,
		MinioRootUser:     "minio",
		MinioRootPassword: rootPassword,
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return nil, err
	}

	return creds, nil
}

// LoadCredentials loads the credentials from the specified file.
func LoadCredentials(filename string) (*Credentials, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	creds := &Credentials{}
	if err := json.Unmarshal(data, creds); err != nil {
		return nil, err
	}

	return creds, nil
}

func (c *Credentials) GetAccessKey() string {
	return c.AccessKey
}

func (c *Credentials) GetSecretKey() string {
	return c.SecretKey
}

func (c *Credentials) GetMinioRootUser() string {
	return c.MinioRootUser
}

func (c *Credentials) GetMinioRootPassword() string {
	return c.MinioRootPassword
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
