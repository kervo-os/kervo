package i18n

import "testing"

// Every language must define every English key natively — a missing
// translation silently falls back to English and ships unnoticed. The
// reverse holds too: a key present only in ko/ja is a typo hiding a
// fallback (the English table is the completeness reference).
func TestTablesComplete(t *testing.T) {
	for _, l := range Supported() {
		for _, k := range Keys() {
			if !Has(l, k) {
				t.Errorf("%s is missing key %q", l, k)
			}
		}
		for k := range tables[l] {
			if !Has(EN, k) {
				t.Errorf("%s has orphan key %q (not in the English reference)", l, k)
			}
		}
	}
}

func TestParse(t *testing.T) {
	for in, want := range map[string]Lang{"": EN, "en": EN, "ko": KO, "ja": JA} {
		if got, err := Parse(in); err != nil || got != want {
			t.Errorf("Parse(%q) = %v/%v, want %v", in, got, err, want)
		}
	}
	if _, err := Parse("de"); err == nil {
		t.Error("unsupported language must error")
	}
}
