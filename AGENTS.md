# Pokémon Slowdown personal fork

- Keep the existing Go/Bubble Tea architecture and Showdown protocol IDs.
- Korean UI copy lives in `internal/tui/i18n.go`; names are embedded from a pinned PokéAPI snapshot.
- Never translate player nicknames, chat, authentication values, or outbound commands.
- Preference changes must apply to open battles and survive restarting the app.
- Keep credentials in the OS keychain; tests must use the existing keyring stub and temporary config directories.
- Validate with `go test ./... -timeout 300s`, `go vet ./...`, and `go build ./...`.
- Test Korean terminal cell widths and UTF-8 boundaries when changing text or layouts.
- Windows and macOS binaries require separate native smoke checks; cross-compilation alone does not prove runtime behavior.
- Preserve upstream MIT and PokéAPI BSD notices in source and binary distributions.
