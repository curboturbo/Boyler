package generator

import (
	"fmt"
	"math/rand"
	"time"
)

var adjectives = []string{
	"happy",
	"clever",
	"brave",
	"calm",
	"focused",
	"kind",
	"sharp",
	"swift",
	"bright",
	"silent",
	"wild",
	"cool",
	"fast",
	"strong",
	"bold",
	"lucky",
	"proud",
	"wise",
	"tiny",
	"giant",
	"red",
	"blue",
	"green",
	"golden",
	"silver",
	"frozen",
	"hidden",
	"ancient",
	"cosmic",
	"electric",
	"magic",
	"rapid",
	"stable",
	"secure",
	"elastic",
	"dynamic",
	"atomic",
	"neon",
	"quantum",
}

var nouns = []string{
	"turing",
	"curie",
	"einstein",
	"tesla",
	"newton",
	"darwin",
	"hopper",
	"lovelace",
	"morse",
	"bell",
	"fermi",
	"hawking",
	"pasteur",
	"bohr",
	"kepler",
	"galileo",
	"archimedes",
	"watson",
	"crick",
	"jobs",
	"wiles",
	"knuth",
	"ritchie",
	"thompson",
	"torvalds",
	"hamilton",
	"ramanujan",
	"curtis",
	"zeppelin",
	"apollo",
	"phoenix",
	"falcon",
	"wolf",
	"lion",
	"tiger",
	"eagle",
	"hawk",
	"otter",
	"fox",
	"bear",
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func NameOrCreate(name string) string {
	if name != "" {
		return name
	}

	return fmt.Sprintf(
		"%s_%s",
		random(adjectives),
		random(nouns),
	)
}

func random(words []string) string {
	return words[rnd.Intn(len(words))]
}