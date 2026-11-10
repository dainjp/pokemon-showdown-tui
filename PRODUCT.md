# Personal Pokémon Showdown terminal client

This is a personal fork of `unnipv/pokemon-slowdown`, connected to the existing
Pokémon Showdown service for human-versus-human battles. The owner requested
Korean support, sprite on/off, visual customization, and an extensible codebase.

## Confirmed needs

- Run on local Windows and macOS PCs for the next test, comparing account login
  against the proxy-locked Ubuntu development server. Provide prebuilt archives
  and Korean installation instructions in `dainjp/pokemon-showdown-tui`.
- Keep the existing keyboard-driven terminal workflow and official server.
- Support Korean alongside English; keep player nicknames, chat, team exports,
  authentication and server commands in their original form.
- Make sprites easy to enable/disable and preferences easy to save.
- Use a calm dark interface with a clear accent. This direction was explicitly
  selected by the owner.

## First version choices

The initial `0.3.0-ko.1` test build defaults to Korean and the `midnight` theme.
F2 opens language, sprite, animation and theme settings, saved on each change.
The interface preserves upstream Go/Bubble Tea structure and existing themes.
Korean reference names come from a pinned PokéAPI CSV snapshot, bundled with its
BSD notice. Common client-generated battle logs are localized; untranslated
server prose, descriptions and rare forms retain an English fallback.

## Successful workflow

Launch the extracted `slowdown` binary in the PC's local terminal with
`--no-sprites`, sign in through the `:` menu, choose a format or challenge another user, and play with the
existing move/switch controls. F2 customization also works while a battle is
open. Small panes must preserve readable actions and connection state.

## Validation boundary

Linux launch, live guest connection, F2 interaction, preference persistence,
recorded-battle rendering, unit tests, race checks and cross-platform builds are
verified in the development environment. Native Windows and macOS launch,
terminal rendering, credential persistence in those OS keyrings, and an actual
human battle still need checks. The development server receives a proxy-lock
notice from Showdown. A separate probe observed a normal connection closure
after authentication; the notice alone does not establish its cause. The client
pauses automatic reauthentication after that sequence. Successful persistent
account login remains unverified.
