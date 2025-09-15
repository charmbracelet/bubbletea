package main

import (
	"math/rand"
	"strings"
	"sync"
)

const uncapitalized = " of a an and ’n’ "

var (
	adjectives = []string{
		"a hot", "a cute", "a fresh", "a nice", "a lovely",
		"an eager", "a soft", "an expensive", "a new", "an old", "a happy",
		"a messy", "a good", "a bad", "a cheesy", "a friendly", "a free",
		"a cold", "a gorgeous", "a glamorous", "a handsome", "an exquisite",
		"a tantalizing", "a suspicious", "an american", "a wooden", "a golden",
		"a dirty", "a hairy", "a lukewarm", "a burning hot", "a shiny",
		"a rogue", "a green", "a late night", "a mass produced", "a handmade",
		"a wild", "a clean", "a rugged", "the #1", "the best", "the worst",
		"a famous", "an infamous", "a clever", "a microwaved", "a 3D printed",
		"your favorite", "your least favorite", "someone’s", "a precious",
		"a fake", "a genuine", "a bejeweled", "a good-smelling",
	}

	nouns = []string{
		"pear", "banana", "bowl of ramen", "currywurst", "quince",
		"pie", "cake", "burrito", "sushi", "basket of fish ’n’ chips", "burger",
		"kohlrabi", "pineapple", "cantaloupe", "sausage roll", "yuzu",
		"grapefruit", "espresso shot", "sandwich", "bowl of chow mein", "lemon",
		"cup of coffee", "bottle of hot sauce", "can of beer", "glass of wine",
		"muffin", "bagel", "glass of champagne", "bottle of rosé", "pengu",
		"badger", "mango", "okonomiyaki", "meatball", "box of wine",
		"artichoke", "TUI", "linux distro", "dotfile", "weißwurst", "computer",
	}

	shuffle     sync.Once
	nextWordMtx sync.Mutex
)

func nextRandomWord() string {
	shuffle.Do(shuffleWords)

	nextWordMtx.Lock()
	defer nextWordMtx.Unlock()

	adjectives = cycle(adjectives)
	nouns = cycle(nouns)

	return capitalize(adjectives[0] + " " + nouns[0])
}

func shuffleWords() {
	shuf := func(x []string) {
		rand.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
	}
	shuf(adjectives)
	shuf(nouns)
}

func capitalize(s string) string {
	words := strings.Fields(s)

	for i, w := range words {
		if i > 0 && strings.Contains(uncapitalized, " "+w+" ") {
			words[i] = w
		} else {
			words[i] = strings.Title(w)
		}
	}

	return strings.Join(words, " ")
}

func cycle(stack []string) []string {
	return append(stack[1:], stack[0])
}
