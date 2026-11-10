---
name: Pokémon Slowdown
description: Korean-first, keyboard-driven Pokémon Showdown terminal client with a terminal-inherited default.
colors:
  canvas: "#111923"
  foreground: "#e6edf3"
  muted: "#a0afbd"
  dim: "#8fa2b4"
  primary: "#69d4c5"
  accent: "#f2bb77"
  danger: "#ff8585"
  success: "#a9d87c"
  warning: "#f2ca78"
  selection: "#24464b"
  type-ink-dark: "#1b1b1b"
  type-ink-light: "#ffffff"
typography:
  title:
    fontFamily: "terminal emulator default monospace"
    fontSize: "terminal default (one cell)"
    fontWeight: 700
    lineHeight: "1 terminal row"
  body:
    fontFamily: "terminal emulator default monospace"
    fontSize: "terminal default (one cell)"
    fontWeight: 400
    lineHeight: "1 terminal row"
  label:
    fontFamily: "terminal emulator default monospace"
    fontSize: "terminal default (one cell)"
    fontWeight: 400
    lineHeight: "1 terminal row"
  key:
    fontFamily: "terminal emulator default monospace"
    fontSize: "terminal default (one cell)"
    fontWeight: 700
    lineHeight: "1 terminal row"
rounded:
  rounded: "lipgloss.RoundedBorder()"
  normal: "lipgloss.NormalBorder()"
  thick: "lipgloss.ThickBorder()"
spacing:
  cell: "1 terminal cell"
  badge-inset: "0 1 terminal cell"
  box-inset: "0 2 terminal cells"
  blank-row: "1 terminal row"
components:
  title:
    textColor: "{colors.primary}"
    typography: "{typography.title}"
  selected-row:
    backgroundColor: "{colors.selection}"
    textColor: "{colors.foreground}"
    padding: "{spacing.badge-inset}"
  focus-box:
    textColor: "{colors.foreground}"
    rounded: "{rounded.rounded}"
    padding: "{spacing.box-inset}"
  status-footer:
    textColor: "{colors.dim}"
    typography: "{typography.label}"
    height: "1 terminal row"
  toast-line:
    textColor: "{colors.accent}"
    typography: "{typography.label}"
    height: "1 terminal row"
  move-row:
    typography: "{typography.body}"
    padding: "{spacing.box-inset}"
  settings-row:
    typography: "{typography.body}"
    padding: "{spacing.badge-inset}"
  hp-full:
    textColor: "{colors.success}"
    typography: "{typography.body}"
  hp-half:
    textColor: "{colors.warning}"
    typography: "{typography.body}"
  hp-low:
    textColor: "{colors.danger}"
    typography: "{typography.body}"
---

# Design System: Pokémon Slowdown

## Overview

**Creative North Star: "The Quiet Battle Console"**

Pokémon Slowdown is a terminal-first battle console: calm enough to read between turns, direct enough to operate from the keyboard. The `terminal` theme is the default: it inherits the terminal foreground and background and uses markers, bold and underline for emphasis. The midnight palette remains opt-in. The interface is a Go/Bubble Tea terminal UI run locally on Windows, macOS, or Linux, so its geometry is measured in cells and rows rather than browser pixels.

The system keeps battle state legible through text, weight, and position as well as color. Korean and English use the same layout rules; Korean names, labels, logs, and truncation are measured by terminal display width, while player nicknames, chat, authentication values, team exports, and outbound commands remain in their original form. The other shipped themes remain selectable and keep their own palettes.

**Key Characteristics:**

- Cell-measured, keyboard-led composition with a one-row contextual footer.
- Midnight canvas with teal orientation, amber action, and explicit status colors.
- Tonal layering, selection fills, and borders provide depth; there are no shadows.
- Sprite graphics are optional and responsive; text battle information remains usable without them.

## Colors

The default theme specifies no foreground or background colors. `terminal` is the default; `midnight`, `zen`, `dark`, `light`, `gameboy`, `mono`, and `contrast` are existing user-selectable themes and must not be flattened into this palette.

### Primary

- **Midnight canvas** (`{colors.canvas}`): Full-screen terminal background and the base foreground for every screen.
- **Cool white** (`{colors.foreground}`): Default readable copy, names, and selected-row text.
- **Clear teal** (`{colors.primary}`): Screen titles, battle title signal, active mechanics, focused borders, the input cursor, and positive orientation.

### Secondary

- **Warm amber** (`{colors.accent}`): Action keys, search activity, selection cursors, and attention that is not an error.
- **Leaf green** (`{colors.success}`): Online/remembered identity, healthy HP, revealed living party members, and successful outcomes.
- **Soft gold** (`{colors.warning}`): Warnings, reconnecting or offline notices, low PP, half HP, and ties.
- **Coral danger** (`{colors.danger}`): Errors, fainted or low-HP states, failed outcomes, and urgent status effects.

### Neutral

- **Muted slate** (`{colors.muted}`): Secondary labels, counts, HP text, field conditions, and explanatory copy.
- **Dim steel** (`{colors.dim}`): Hints, hidden party members, disabled detail, and low-priority footer copy.
- **Selection sea** (`{colors.selection}`): Full-width selected rows and the settings cursor background.
- **Type badge ink** (`{colors.type-ink-dark}` / `{colors.type-ink-light}`): Black or white badge text selected per Pokémon type for contrast.

Type badges use the canonical per-type background map in `internal/tui/theme.go`; move names use the lighter dark-surface type foreground map. These are functional battle labels, not a replacement for the semantic application palette.

**The Text Carries State Rule.** Gameplay meaning must remain present in a word, number, marker, or layout change; color reinforces status but never carries it alone.

**The One Clear Accent Rule.** Use teal for orientation and focus, amber for an actionable or transient signal, and the semantic status colors only for their stated states. Do not turn every line into an accent.

## Typography

**Display Font:** terminal emulator default monospace; the application does not select a display face.
**Body Font:** terminal emulator default monospace.
**Label/Mono Font:** the same terminal monospace; there is no separate web or display family.

**Character:** One terminal cell is the base typographic unit. Bold weight creates the title and action hierarchy; muted and faint styles lower priority without changing the cell grid. There is no browser-style display scale or letter-spacing token.

### Hierarchy

- **Title** (bold, terminal default size, one row): Screen names, the Pokémon Slowdown header, battle title signal, and overlay titles.
- **Body** (regular, terminal default size, one row): Names, logs, move detail, settings values, and normal interaction copy.
- **Label** (regular, terminal default size, one row): Footer hints, counts, field conditions, and secondary labels.
- **Key** (bold, terminal default size, one row): Move numbers, selection markers, active mechanics, and other keyboard affordances.

Every line is constrained by terminal columns and measured with `lipgloss.Width`. Korean strings must remain valid UTF-8 and must be truncated by display cells, not bytes.

**The Cell Grid Rule.** Never introduce a font, size, line-height, or tracking assumption that cannot be represented by the user's terminal. Preserve one-cell key alignment and the existing CJK width checks.

## Layout

The UI is a full-screen alternate-screen terminal view. The usable floor is 32 columns by 12 rows; below that the app shows a resize message. The body is rendered first, then optional toast and connection-notice rows, then one contextual status row. Body height always subtracts the rows that are actually present so the footer remains visible.

Width selects a semantic layout: `compact` below 46 columns, `sidecar` at 46–69, `standard` at 70–99, and `cinematic` at 100 or more. A pane under 14 rows forces `compact`; a pane under 26 rows downgrades a sprite layout to `sidecar`. Compact battle layouts omit sprites and field conditions and keep active Pokémon on one line. Sidecar reserves 9×5 cells for a sprite, standard 12×6, and cinematic 18×9. Wide layouts place the sprite beside two lines of Pokémon detail.

The lobby uses a list-only composition in compact mode. At wider sizes it gives roughly three-fifths of the width to the format list, a two-cell gap, and the remainder to status and connection context. Battle screens keep the header, active field, conditions, current-turn log tail, and move list in priority order; the log yields space before the move list does.

Focused palette, prompt, and battle overlays use a bordered box with two-cell horizontal padding. Battle overlay frames clamp to a 20-column minimum and a 78-column maximum, with an inner width six cells smaller; battle overlays are placed near the top third and clamped to remain visible. Selected rows use a leading marker and a full-width selection fill.

**The Cell-Fit Rule.** Every screen must fit its measured width and height, including Korean text, optional sprites, overlays, notices, and the footer. Reduce lower-priority detail before hiding an action or connection state.

## Elevation & Depth

The default terminal theme leaves Canvas and type-badge colors unset, with no inverse or filled selection background. In opt-in colored themes, Canvas supplies the selected palette background; depth comes from the muted or primary border on focus boxes, the selection background on the active row, and foreground weight. The app has no CSS or terminal shadow vocabulary. `midnight` and the default dark-family themes use rounded Lipgloss borders; `light`, `gameboy`, and `mono` use normal borders; `contrast` uses a thick border.

### Named Rules

**The Flat Terminal Rule.** Use background tone, border, and text hierarchy to separate regions. Do not add a web-style shadow, blur, gradient, or glass surface to a terminal component.

## Shapes

The default form is a rounded Lipgloss border around a focused command palette, prompt, or battle overlay. The box has two terminal cells of horizontal interior padding. Selected list and settings rows are rectangular full-width fills with a leading `▸`; they are not pills. Pokémon type badges use one terminal cell of horizontal padding and a fixed-width badge column where move rows need alignment. HP bars use the existing `█`/`░` cell bar, while party state uses textual `●`/`○` markers alongside counts.

## Components

### Titles and headers

- **Character:** Quiet orientation at the top of each screen.
- **Color and type:** Bold primary teal for the title; muted slate for the format or help suffix; regular foreground for turn counters.
- **Behavior:** Battle headers retain the title and turn/timer information first, dropping the format name when width is tight.

### Selected rows

- **Shape:** Full-width rectangular fill with a leading `▸` marker.
- **Color:** Selection sea background with cool-white text; amber marks action cursors where the row is not itself a full selected row.
- **States:** Unselected rows use foreground text; disabled or finished detail uses dim/faint text.

### Focus boxes

- **Shape:** Rounded primary border in the default theme, two-cell horizontal padding, width capped by the overlay rules.
- **Uses:** Command palette, sign-in and import prompts, and battle switch/inspect/log/chat/target overlays.
- **States:** The current input receives the primary cursor `▌`; completed fields dim while secret values stay masked; overlay content trims with an explicit remaining-lines marker.

### Battle field and party tracker

- **Character:** Information-dense but ordered around the active Pokémon and current turn.
- **Content:** Name, type badge, HP bar and value, status, boosts, mechanic, item when appropriate, then revealed-party dots and counts.
- **States:** Healthy, half, and low HP use success, warning, and danger respectively; fainted, hidden, revealed, and active members also carry text or count context. Sprites may be on or off without removing the textual field.

### Move rows

- **Shape:** One row per move with a fixed name column, type badge, optional class/base-power detail at 64 or more columns, and PP aligned to the right.
- **Color:** Bold amber move number; move name uses the type foreground map; type badge uses the canonical type background; PP shifts muted → warning → danger as it becomes scarce or empty.
- **States:** Disabled moves and submitted choices use the disabled/faint treatment; a submitted choice adds a check and an explicit waiting message.

### Status footer and transient lines

- **Status footer:** One row with the most specific keyboard hint that fits on the left and connection state on the right. At 70 or more columns it may include the username; at 100 or more it may include search activity.
- **Toast:** A separate accent-colored foreground line that lives above the footer and clears after six seconds.
- **Connection notice:** A separate warning line above the footer when present; connection state is still repeated in the footer.
- **Compact behavior:** At narrow widths the footer keeps move, switch, settings/menu, and connection affordances while dropping longer hints and usernames.

### Settings rows

- **Shape:** Four text rows for language, sprites, animation, and theme; labels are padded to a 12-cell column.
- **States:** The selected row uses the selection fill; values are plain text (`켜짐`/`꺼짐` or `On`/`Off`), and each change saves immediately. Sprite changes rebuild the active renderer and apply to open battles.

## Do's and Don'ts

### Do:

- **Do** inherit terminal colors by default and keep all seven colored themes available through the existing theme cycle.
- **Do** reserve teal for orientation/focus and amber for action or transient attention; pair every status color with readable text, a number, or a marker.
- **Do** measure Korean and English by terminal display cells, preserve UTF-8 boundaries, and leave nicknames, chat, credentials, exports, and protocol commands untouched.
- **Do** keep battle actions usable when sprites are disabled or the pane falls into compact mode; preserve the connection state in the footer.
- **Do** keep ANSI output and cell widths portable across local Windows, macOS, and Linux terminals.

### Don't:

- **Don't** turn the terminal into a web surface with CSS fonts, browser pixel spacing, cards, gradients, shadows, blur, or decorative glass.
- **Don't** use color as the only carrier of battle state or replace the canonical Pokémon type vocabulary with arbitrary theme colors.
- **Don't** remove the existing zen, dark, light, gameboy, mono, or contrast themes to enforce midnight.
- **Don't** treat the seven recorded review frames as proof of an unproxied human battle or of every Windows/macOS terminal-client path; the server's public IP is proxy-locked and those checks remain open in `PRODUCT.md` and `finish-review.md`.
