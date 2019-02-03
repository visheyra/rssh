package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	// "crypto/sha512"

	"encoding/base64"
	"encoding/json"
	"encoding/pem"

	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"go.etcd.io/etcd/client"
	"golang.org/x/crypto/ssh"
)

// AgentCredentials represent the
type AgentCredentials struct {
	ID uuid.UUID
	// Agent public SSH Key
	Identity []byte
	// Optional agent secret
	Secret []byte
}

func (a *AgentCredentials) MarshalJSON() ([]byte, error) {
	if a.Identity == nil {
		return nil, errors.New("missing identity to generate agent id")
	}
	return json.Marshal(&struct {
		ID         string `json:"aid"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}{
		ID:         a.ID.String(),
		PublicKey:  base64.StdEncoding.EncodeToString(a.Identity),
		PrivateKey: base64.StdEncoding.EncodeToString(a.Secret),
	})
}

func (a *AgentCredentials) DropSecrets() {
	a.Secret = nil
}

func generatePrivateKey() (*rsa.PrivateKey, error) {
	bitLength := 2048
	r := rand.Reader

	privateKey, err := rsa.GenerateKey(r, bitLength)
	if err != nil {
		return nil, err
	}
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func generateSSHPublicKey(privateKey *rsa.PublicKey) ([]byte, error) {
	pub, err := ssh.NewPublicKey(privateKey)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(pub), nil
}

func serializePrivateKey(privateKey *rsa.PrivateKey) []byte {
	privateKeySerialized := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPemBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: map[string]string{"foo": "oof", "bar": "rab"},
		Bytes:   privateKeySerialized,
	}
	privateKeyPem := pem.EncodeToMemory(&privateKeyPemBlock)
	return privateKeyPem
}

func generateAgentKeys() (pub []byte, priv []byte, err error) {
	privateKey, err := generatePrivateKey()
	if err != nil {
		return nil, nil, err
	}
	priv = serializePrivateKey(privateKey)

	pub, err = generateSSHPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	return pub, priv, nil
}

func GenerateAgentCredentials(domain string) (*AgentCredentials, error) {
	log.Debug().Str("domain", domain).Msg("Generating new agent credentials.")
	agentID := uuid.NewV4()

	pub, priv, err := generateAgentKeys()
	if err != nil {
		return nil, err
	}

	credentials := &AgentCredentials{
		ID:       agentID,
		Identity: pub,
		Secret:   priv,
	}
	log.Debug().
		Str("domain", domain).
		Str("Identity", agentID.String()).
		Msg("Generated account credentials.")
	return credentials, nil
}

func PersistAgentCredentials(etcd client.KeysAPI, creds AgentCredentials) error {
	log.Debug().
		Str("agent", creds.ID.String())
	_, err := etcd.Set(
		context.Background(),
		fmt.Sprintf("/agents/%s", creds.ID.String()),
		"{}",
		nil,
	)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not persist agent in etcd.")
		return err
	}
	return nil
}
