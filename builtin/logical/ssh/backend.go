package ssh

import (
	"strings"

	"github.com/hashicorp/vault/helper/salt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type backend struct {
	*framework.Backend
	view logical.Storage
	salt *salt.Salt
}

func Factory(conf *logical.BackendConfig) (logical.Backend, error) {
	b, err := Backend(conf)
	if err != nil {
		return nil, err
	}
	return b.Setup(conf)
}

func Backend(conf *logical.BackendConfig) (*backend, error) {
	var b backend
	b.view = conf.StorageView
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),

		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"verify",
			},

			LocalStorage: []string{
				"otp/",
			},
		},

		Paths: []*framework.Path{
			pathConfigZeroAddress(&b),
			pathKeys(&b),
			pathListRoles(&b),
			pathRoles(&b),
			pathCredsCreate(&b),
			pathLookup(&b),
			pathVerify(&b),
		},

		Secrets: []*framework.Secret{
			secretDynamicKey(&b),
			secretOTP(&b),
		},

		Init: b.Initialize,
	}
	return &b, nil
}

func (b *backend) Initialize() error {
	salt, err := salt.NewSalt(b.view, &salt.Config{
		HashFunc: salt.SHA256Hash,
	})
	if err != nil {
		return err
	}
	b.salt = salt
	return nil
}

const backendHelp = `
The SSH backend generates credentials allowing clients to establish SSH
connections to remote hosts.

There are two variants of the backend, which generate different types of
credentials: dynamic keys and One-Time Passwords (OTPs). The desired behavior
is role-specific and chosen at role creation time with the 'key_type'
parameter.

Please see the backend documentation for a thorough description of both
types. The Vault team strongly recommends the OTP type.

After mounting this backend, before generating credentials, configure the
backend's lease behavior using the 'config/lease' endpoint and create roles
using the 'roles/' endpoint.
`
