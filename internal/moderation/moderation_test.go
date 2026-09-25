package moderation

import "testing"

func TestContainsObjectionable(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		// clean text, including words that contain a blocked term
		{"normal comment", "that twist at 1:23 was insane", false},
		{"profanity is allowed", "holy shit this scene", false},
		{"substring raccoon", "the raccoon cameo lol", false},
		{"substring spicy", "spicy take but I agree", false},
		{"substring cocoon", "Cocoon (1985) vibes", false},
		{"double letters untouched", "what a con artist", false},
		{"rape scene discussion", "that scene was hard to watch", false},
		{"empty", "", false},

		// blocked
		{"plain slur", "you are a faggot", true},
		{"uppercase", "FAGGOT", true},
		{"punctuation around", "...retard!!", true},
		{"leetspeak", "n1gg3r", true},
		{"symbol substitution", "f@g", true},
		{"stretched letters", "faaaggooot", true},
		{"spaced out", "f a g", true},
		{"dotted", "k.y.s", true},
		{"phrase", "just kill yourself", true},
		{"phrase with leet", "k1ll yourself", true},
		{"username underscore", "big_retard_99", true},

		// added from go-away review + additional slurs
		{"slur variant niggah", "niggah", true},
		{"slur variant nigguh", "nigguh", true},
		{"slur variant niggaz", "niggaz", true},
		{"slur niglet", "niglet", true},
		{"slur jigaboo", "jigaboo", true},
		{"slur porchmonkey", "porchmonkey", true},
		{"slur junglebunny", "junglebunny", true},
		{"slur zipperhead", "zipperhead", true},
		{"slur chingchong", "chingchong", true},
		{"slur kyke", "kyke", true},
		{"slur hymie", "hymie", true},
		{"slur fudgepacker", "fudgepacker", true},
		{"slur faggy", "faggy", true},
		{"slur shemale", "shemale", true},
		{"slur retarded", "that is so retarded", true},
		{"slur mongoloid", "mongoloid", true},
		{"phrase porch monkey", "porch monkey", true},
		{"phrase jungle bunny", "jungle bunny", true},
		{"phrase ching chong", "ching chong", true},
		{"phrase kill yourselves", "kill yourselves", true},
		{"phrase neck yourself", "go neck yourself", true},
		{"phrase unalive yourself", "unalive yourself", true},
		{"phrase hope you die", "I hope you die", true},
		{"phrase drink bleach", "go drink bleach", true},

		// movie/TV discussion that must stay allowed
		{"show title queer", "Queer as Folk is a great show", false},
		{"history nazi", "the nazi officer in Inglourious Basterds", false},
		{"character dies", "I hope he dies in the finale", false},
		{"monkey in plot", "the monkey in the jungle scene", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsObjectionable(tt.text); got != tt.want {
				t.Errorf("ContainsObjectionable(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}
