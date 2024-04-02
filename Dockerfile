FROM golang:1.22

# Set destination for COPY
WORKDIR /app

# Copy the source code.
COPY . .

ARG TARGET=wasip1 # js

# Build
RUN GOOS="$TARGET" GOARCH="wasm" CGO_ENABLED=0 go build -ldflags="-s -w" -tags "gowaku_no_rln,wasm" -o build/bin/libwallet.wasm ./wasm/wallet/
RUN ls -la build/bin/libwallet.wasm 
RUN echo "Target is " $TARGET

# Run
CMD [goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))']
