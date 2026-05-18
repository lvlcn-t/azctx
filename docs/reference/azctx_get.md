## azctx get

Get information about a context

### Synopsis

Get information about one context. When NAME is omitted, it returns the current context.

```
azctx get [name] [flags]
```

### Examples

```
  azctx get
  azctx get prod -o json
  azctx get prod -o table
```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of: text|table|json (default "text")
```

### SEE ALSO

* [azctx](azctx.md)	 - A CLI tool for managing Azure contexts.

