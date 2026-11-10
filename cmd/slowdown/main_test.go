package main

import (
	"flag"
	"reflect"
	"testing"
)

func TestInterspersedFlags(t *testing.T) {
	for _, args := range [][]string{
		{"play", "gen9ou", "--lang", "ko", "--no-sprites"},
		{"--lang=ko", "play", "--no-sprites", "gen9ou"},
		{"--no-sprites", "--lang", "ko", "play", "gen9ou"},
	} {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		lang := fs.String("lang", "", "")
		noSprites := fs.Bool("no-sprites", false, "")
		flags, positional, _, err := splitArgs(fs, args)
		if err != nil {
			t.Fatal(err)
		}
		if err := fs.Parse(flags); err != nil {
			t.Fatal(err)
		}
		if *lang != "ko" || !*noSprites || !reflect.DeepEqual(positional, []string{"play", "gen9ou"}) {
			t.Fatalf("%v: %v, %s, %v", args, positional, *lang, *noSprites)
		}
	}
}

func TestDoctorAndInvalidFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("sprites", "", "")
	fs.String("lang", "", "")
	_, pos, probe, err := splitArgs(fs, []string{"doctor", "--sprites"})
	if err != nil || !probe || len(pos) != 1 {
		t.Fatalf("doctor: %v %v %v", pos, probe, err)
	}
	for _, args := range [][]string{{"play", "gen9ou", "--lang"}, {"play", "gen9ou", "--invalid"}} {
		if _, _, _, err := splitArgs(fs, args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
