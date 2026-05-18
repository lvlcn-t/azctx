## azctx set-credential

Set a credential entry in azctx config

### Synopsis

Set a credential entry in azctx config.

```
azctx set-credential NAME [flags]
```

### Examples

```
  azctx set-credential ci-sp --type service-principal \
    --client-id 11111111-1111-1111-1111-111111111111 \
    --client-secret super-secret
```

### Options

```
      --client-certificate-path string   Client certificate path
      --client-id string                 Client ID
      --client-secret string             Client secret
      --federated-token-file string      Path to federated token file
  -h, --help                             help for set-credential
      --type string                      Credential type: service-principal|user|managed-identity|oidc
```

### SEE ALSO

* [azctx](azctx.md)	 - A CLI tool for managing Azure contexts.

