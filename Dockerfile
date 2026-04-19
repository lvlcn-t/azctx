FROM mcr.microsoft.com/azure-cli:2.85.0

COPY azctx /usr/local/bin/azctx

USER nonroot

ENTRYPOINT ["/usr/local/bin/azctx"]
