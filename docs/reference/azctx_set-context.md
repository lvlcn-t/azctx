## azctx set-context

Set a context entry in azctx config

### Synopsis

Set a context entry in azctx config. The context points to tenant and credential entries in the same merged azctx config.

```
azctx set-context NAME [flags]
```

### Examples

```
  azctx set-context prod --tenant corp --credential ci-sp --subscription 00000000-0000-0000-0000-000000000000
```

### Options

```
      --credential string     Credential name for the context
  -h, --help                  help for set-context
      --subscription string   Optional subscription ID for the context
      --tenant string         Tenant name for the context
```

### SEE ALSO

* [azctx](azctx.md)	 - A CLI tool for managing Azure contexts.

