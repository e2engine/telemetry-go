package model

import (
	"fmt"
	"strings"
	"testing"
)

func TestInstallationValidation(t *testing.T) {
	id := strings.Repeat("A", 22)
	valid := Installation{ID: id, Token: "et_" + id + "." + strings.Repeat("A", 43)}
	if !valid.IsValid() {
		t.Fatal("rejected valid installation")
	}

	for _, installation := range []Installation{
		{},
		{ID: id, Token: "invalid"},
		{ID: "wrong", Token: valid.Token},
		{ID: id, Token: "et_" + id + "." + strings.Repeat("A", 42)},
		{ID: id, Token: "et_" + strings.Repeat("A", 21) + "." + strings.Repeat("A", 43)},
	} {
		if installation.IsValid() {
			t.Fatalf("accepted invalid installation: %#v", installation)
		}
	}
}

func TestInstallationFormattingRedactsToken(t *testing.T) {
	id := strings.Repeat("A", 22)
	installation := Installation{ID: id, Token: "et_" + id + "." + strings.Repeat("A", 43)}

	for _, diagnostic := range []string{
		fmt.Sprint(installation),
		fmt.Sprintf("%+v", installation),
		fmt.Sprintf("%#v", installation),
	} {
		if strings.Contains(diagnostic, installation.Token) {
			t.Fatal("diagnostic exposed installation token")
		}
	}
}
