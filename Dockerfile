FROM busybox:stable

# Set the working go module directory for the build
ARG GOMOD

# Expose required ports
EXPOSE 9090

# Expose required volumes
VOLUME /etc/echoctl

# Add application binary and use it as default entrypoint
COPY echoctl /bin/echoctl

# Add config file
COPY config.yaml /root

# Add source code; required to produce complete stack traces.
# The path used MUST match the one used when building
# the application binary: `go env GOMOD`.
RUN mkdir -p ${GOMOD} || true
COPY . ${GOMOD}
RUN rm ${GOMOD}/echoctl || true

# Will be used by the `errors` package to simplify the
# paths reported on stack traces.
ENV GOPATH=${GOMOD}

# Set the default entrypoint
ENTRYPOINT ["/bin/echoctl"]
