FROM mcr.microsoft.com/azure-cli:2.86.0

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/azctx /usr/local/bin/azctx

USER nonroot

ENTRYPOINT ["/usr/local/bin/azctx"]
