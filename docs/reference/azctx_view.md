## azctx view

Display merged azctx config

### Synopsis

Display merged azctx config from AZCTX path list or the default config path.

```
azctx view [flags]
```

### Examples

```
  azctx view
  azctx view -o json
  azctx view --raw -o json
```

### Options

```
  -h, --help            help for view
  -o, --output string   Output format. One of: text|table|json (default "text")
      --raw             Print the source file instead of merged output
```

### SEE ALSO

* [azctx](azctx.md)	 - A CLI tool for managing Azure contexts.

