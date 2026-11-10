// Command slowdown is a terminal-native Pokémon Showdown client.
//
// Usage:
//
//	slowdown                     open the lobby
//	slowdown rand                queue a random battle immediately
//	slowdown play <format>       queue a specific format
//	slowdown challenge <user>    challenge a user
//	slowdown spectate <battle>   watch a battle
//	slowdown teams               open team management
//	slowdown doctor [--sprites]  print terminal diagnostics
//	slowdown --version
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/unnipv/pokemon-slowdown/internal/app"
	"runtime/debug"
)

// version is set at build time by GoReleaser via -ldflags.
var version = "dev"

// versionString reports the release version. Binaries built by GoReleaser carry
// it in the version variable; binaries installed with `go install` do not get
// ldflags, but the module version is recorded in the build info, so read that
// instead of reporting "dev".
func versionString() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "slowdown:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("slowdown", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		showVersion = fs.Bool("version", false, "print the version and exit")
		debug       = fs.Bool("debug", false, "write sanitized protocol events to a log file")
		noSprites   = fs.Bool("no-sprites", false, "disable sprites entirely")
		spritesMode = fs.String("sprites", "", "sprite backend: auto|kitty|iterm2|sixel|blocks|none")
		theme       = fs.String("theme", "", "colour theme: "+strings.Join(themeNames(), ", "))
		language    = fs.String("lang", "", "interface language: ko or en")
	)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: slowdown [command] [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  (none)              open the lobby")
		fmt.Fprintln(os.Stderr, "  rand                queue a random battle")
		fmt.Fprintln(os.Stderr, "  play <format>       queue a format, e.g. play gen9ou")
		fmt.Fprintln(os.Stderr, "  challenge <user>    choose a format and challenge a user")
		fmt.Fprintln(os.Stderr, "  spectate <battle>   watch a battle id or replay URL")
		fmt.Fprintln(os.Stderr, "  teams               open team management")
		fmt.Fprintln(os.Stderr, "  doctor [--sprites]  print terminal diagnostics")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "flags:")
		fs.PrintDefaults()
	}

	flags, positional, probe, err := splitArgs(fs, args)
	if err != nil {
		return err
	}
	if err := fs.Parse(flags); err != nil {
		return err
	}
	cmd := ""
	if len(positional) > 0 {
		cmd, positional = positional[0], positional[1:]
	}
	if cmd == "doctor" {
		return app.Doctor(os.Stdout, probe)
	}

	if *showVersion {
		fmt.Println("pokemon slowdown", versionString())
		return nil
	}
	if *language != "" && *language != "ko" && *language != "en" {
		return fmt.Errorf("unknown language %q (use ko or en)", *language)
	}

	opts := app.Options{
		Language:    *language,
		Debug:       *debug,
		NoSprites:   *noSprites,
		SpritesMode: *spritesMode,
		Theme:       *theme,
		Version:     version,
	}

	switch cmd {
	case "":
		// plain launch
	case "rand":
		opts.Format = "gen9randombattle"
	case "play":
		if len(positional) < 1 {
			return fmt.Errorf("play needs a format, e.g. slowdown play gen9ou")
		}
		opts.Format = positional[0]
	case "challenge":
		if len(positional) < 1 {
			return fmt.Errorf("challenge needs a username")
		}
		opts.Challenge = positional[0]
	case "spectate":
		if len(positional) < 1 {
			return fmt.Errorf("spectate needs a battle id or replay URL")
		}
		opts.Spectate = positional[0]
	case "teams":
		opts.OpenTeams = true
	case "version":
		fmt.Println("pokemon slowdown", versionString())
		return nil
	default:
		return fmt.Errorf("unknown command %q (try slowdown --help)", cmd)
	}

	return app.Run(opts)
}

func themeNames() []string {
	return []string{"midnight", "zen", "dark", "light", "gameboy", "mono", "contrast"}
}

// splitArgs keeps positional arguments while allowing flags before or after them.
func splitArgs(fs *flag.FlagSet, args []string) (flags, positional []string, probe bool, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positional = append(positional, arg)
			continue
		}
		name, _, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if len(positional) > 0 && positional[0] == "doctor" && name == "sprites" {
			if hasValue {
				return nil, nil, false, fmt.Errorf("doctor --sprites takes no value")
			}
			probe = true
			continue
		}
		f := fs.Lookup(name)
		if f == nil {
			if name == "h" || name == "help" {
				flags = append(flags, arg)
				continue
			}
			return nil, nil, false, fmt.Errorf("unknown flag %s", arg)
		}
		flags = append(flags, arg)
		boolean, ok := f.Value.(interface{ IsBoolFlag() bool })
		if !hasValue && !(ok && boolean.IsBoolFlag()) {
			i++
			if i == len(args) {
				return nil, nil, false, fmt.Errorf("flag %s needs a value", arg)
			}
			flags = append(flags, args[i])
		}
	}
	return
}
