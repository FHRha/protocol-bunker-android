# Go Binaries For Android

Place executable files named `server-go` in ABI-specific folders:

- `server-binaries/arm64-v8a/server-go`
- `server-binaries/armeabi-v7a/server-go`
- `server-binaries/x86_64/server-go`
- `server-binaries/x86/server-go`

The app copies the matching binary to:

- `filesDir/server-go/server-go`

on startup when `DEV_MODE` is disabled.
