FROM golang:1.19

RUN go install github.com/shurcooL/goexec@latest

# Set destination for COPY
WORKDIR /app

# Copy the source code.
COPY . .

# Build
#RUN GOOS=js GOARCH=wasm go build -o build/bin/libstatus.wasm ./build/bin/statusgo-lib
RUN make statusgo-library-wasm

RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# Run
CMD [goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))']
