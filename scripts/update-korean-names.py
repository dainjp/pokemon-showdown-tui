#!/usr/bin/env python3
"""Regenerate the embedded Korean catalogue from a pinned PokéAPI snapshot."""

import csv
import io
import json
from pathlib import Path
import re
from urllib.request import urlopen

SNAPSHOT = "5cbeb3f3eb55a9b86da52c4869336f9d61fdd6d7"
BASE = f"https://raw.githubusercontent.com/PokeAPI/pokeapi/{SNAPSHOT}"
ROOT = Path(__file__).resolve().parents[1]
FILES = {
    "pokemon": "pokemon_species_names.csv",
    "move": "move_names.csv",
    "ability": "ability_names.csv",
    "item": "item_names.csv",
    "type": "type_names.csv",
}


def download(path):
    with urlopen(f"{BASE}/{path}", timeout=30) as response:
        return response.read().decode("utf-8")


def main():
    catalogue = {}
    for category, filename in FILES.items():
        rows = list(csv.DictReader(io.StringIO(download(f"data/v2/csv/{filename}"))))
        identifier = next(iter(rows[0]))
        english = {row[identifier]: row["name"] for row in rows if row["local_language_id"] == "9"}
        entries = {}
        for row in rows:
            if row["local_language_id"] != "3" or row[identifier] not in english:
                continue
            name = english[row[identifier]]
            if category == "pokemon" and row[identifier] in {"29", "32"}:
                name = "Nidoran-F" if row[identifier] == "29" else "Nidoran-M"
            key = re.sub(r"[^a-z0-9]", "", name.lower())
            if key:
                entries[key] = row["name"]
        catalogue[category] = entries
        print(f"{category}: {len(entries)} Korean names")
    catalogue["type"]["stellar"] = "스텔라"
    target = ROOT / "internal/tui/names_ko.json"
    target.write_text(json.dumps(catalogue, ensure_ascii=False, sort_keys=True, indent=2) + "\n", encoding="utf-8")
    (ROOT / "docs/POKEAPI-LICENSE.txt").write_text(download("LICENSE.md"), encoding="utf-8")


if __name__ == "__main__":
    main()
