package tea

import (
	"strings"

	"github.com/xo/terminfo"
)

func buildTerminfoKeys(flags int, term string) map[string]Key {
	table := make(map[string]Key)
	ti, _ := terminfo.Load(term)
	if ti == nil {
		return table
	}

	tiTable := defaultTerminfoKeys(flags)

	// Default keys
	for name, seq := range ti.StringCapsShort() {
		if !strings.HasPrefix(name, "k") || len(seq) == 0 {
			continue
		}

		if k, ok := tiTable[name]; ok {
			table[string(seq)] = k
		}
	}

	// Extended keys
	for name, seq := range ti.ExtStringCapsShort() {
		if !strings.HasPrefix(name, "k") || len(seq) == 0 {
			continue
		}

		if k, ok := tiTable[name]; ok {
			table[string(seq)] = k
		}
	}

	return table
}

// This returns a map of terminfo keys to key events. It's a mix of ncurses
// terminfo default and user-defined key capabilities.
// Upper-case caps that are defined in the default terminfo database are
//   - kNXT
//   - kPRV
//   - kHOM
//   - kEND
//   - kDC
//   - kIC
//   - kLFT
//   - kRIT
//
// See https://man7.org/linux/man-pages/man5/terminfo.5.html
// See https://github.com/mirror/ncurses/blob/master/include/Caps-ncurses
func defaultTerminfoKeys(flags int) map[string]Key {
	keys := map[string]Key{
		"kcuu1": {Sym: KeyUp},
		"kUP":   {Sym: KeyUp, Mod: ModShift},
		"kUP3":  {Sym: KeyUp, Mod: ModAlt},
		"kUP4":  {Sym: KeyUp, Mod: ModShift | ModAlt},
		"kUP5":  {Sym: KeyUp, Mod: ModCtrl},
		"kUP6":  {Sym: KeyUp, Mod: ModShift | ModCtrl},
		"kUP7":  {Sym: KeyUp, Mod: ModAlt | ModCtrl},
		"kUP8":  {Sym: KeyUp, Mod: ModShift | ModAlt | ModCtrl},
		"kcud1": {Sym: KeyDown},
		"kDN":   {Sym: KeyDown, Mod: ModShift},
		"kDN3":  {Sym: KeyDown, Mod: ModAlt},
		"kDN4":  {Sym: KeyDown, Mod: ModShift | ModAlt},
		"kDN5":  {Sym: KeyDown, Mod: ModCtrl},
		"kDN7":  {Sym: KeyDown, Mod: ModAlt | ModCtrl},
		"kDN6":  {Sym: KeyDown, Mod: ModShift | ModCtrl},
		"kDN8":  {Sym: KeyDown, Mod: ModShift | ModAlt | ModCtrl},
		"kcub1": {Sym: KeyLeft},
		"kLFT":  {Sym: KeyLeft, Mod: ModShift},
		"kLFT3": {Sym: KeyLeft, Mod: ModAlt},
		"kLFT4": {Sym: KeyLeft, Mod: ModShift | ModAlt},
		"kLFT5": {Sym: KeyLeft, Mod: ModCtrl},
		"kLFT6": {Sym: KeyLeft, Mod: ModShift | ModCtrl},
		"kLFT7": {Sym: KeyLeft, Mod: ModAlt | ModCtrl},
		"kLFT8": {Sym: KeyLeft, Mod: ModShift | ModAlt | ModCtrl},
		"kcuf1": {Sym: KeyRight},
		"kRIT":  {Sym: KeyRight, Mod: ModShift},
		"kRIT3": {Sym: KeyRight, Mod: ModAlt},
		"kRIT4": {Sym: KeyRight, Mod: ModShift | ModAlt},
		"kRIT5": {Sym: KeyRight, Mod: ModCtrl},
		"kRIT6": {Sym: KeyRight, Mod: ModShift | ModCtrl},
		"kRIT7": {Sym: KeyRight, Mod: ModAlt | ModCtrl},
		"kRIT8": {Sym: KeyRight, Mod: ModShift | ModAlt | ModCtrl},
		"kich1": {Sym: KeyInsert},
		"kIC":   {Sym: KeyInsert, Mod: ModShift},
		"kIC3":  {Sym: KeyInsert, Mod: ModAlt},
		"kIC4":  {Sym: KeyInsert, Mod: ModShift | ModAlt},
		"kIC5":  {Sym: KeyInsert, Mod: ModCtrl},
		"kIC6":  {Sym: KeyInsert, Mod: ModShift | ModCtrl},
		"kIC7":  {Sym: KeyInsert, Mod: ModAlt | ModCtrl},
		"kIC8":  {Sym: KeyInsert, Mod: ModShift | ModAlt | ModCtrl},
		"kdch1": {Sym: KeyDelete},
		"kDC":   {Sym: KeyDelete, Mod: ModShift},
		"kDC3":  {Sym: KeyDelete, Mod: ModAlt},
		"kDC4":  {Sym: KeyDelete, Mod: ModShift | ModAlt},
		"kDC5":  {Sym: KeyDelete, Mod: ModCtrl},
		"kDC6":  {Sym: KeyDelete, Mod: ModShift | ModCtrl},
		"kDC7":  {Sym: KeyDelete, Mod: ModAlt | ModCtrl},
		"kDC8":  {Sym: KeyDelete, Mod: ModShift | ModAlt | ModCtrl},
		"khome": {Sym: KeyHome},
		"kHOM":  {Sym: KeyHome, Mod: ModShift},
		"kHOM3": {Sym: KeyHome, Mod: ModAlt},
		"kHOM4": {Sym: KeyHome, Mod: ModShift | ModAlt},
		"kHOM5": {Sym: KeyHome, Mod: ModCtrl},
		"kHOM6": {Sym: KeyHome, Mod: ModShift | ModCtrl},
		"kHOM7": {Sym: KeyHome, Mod: ModAlt | ModCtrl},
		"kHOM8": {Sym: KeyHome, Mod: ModShift | ModAlt | ModCtrl},
		"kend":  {Sym: KeyEnd},
		"kEND":  {Sym: KeyEnd, Mod: ModShift},
		"kEND3": {Sym: KeyEnd, Mod: ModAlt},
		"kEND4": {Sym: KeyEnd, Mod: ModShift | ModAlt},
		"kEND5": {Sym: KeyEnd, Mod: ModCtrl},
		"kEND6": {Sym: KeyEnd, Mod: ModShift | ModCtrl},
		"kEND7": {Sym: KeyEnd, Mod: ModAlt | ModCtrl},
		"kEND8": {Sym: KeyEnd, Mod: ModShift | ModAlt | ModCtrl},
		"kpp":   {Sym: KeyPgUp},
		"kprv":  {Sym: KeyPgUp},
		"kPRV":  {Sym: KeyPgUp, Mod: ModShift},
		"kPRV3": {Sym: KeyPgUp, Mod: ModAlt},
		"kPRV4": {Sym: KeyPgUp, Mod: ModShift | ModAlt},
		"kPRV5": {Sym: KeyPgUp, Mod: ModCtrl},
		"kPRV6": {Sym: KeyPgUp, Mod: ModShift | ModCtrl},
		"kPRV7": {Sym: KeyPgUp, Mod: ModAlt | ModCtrl},
		"kPRV8": {Sym: KeyPgUp, Mod: ModShift | ModAlt | ModCtrl},
		"knp":   {Sym: KeyPgDown},
		"knxt":  {Sym: KeyPgDown},
		"kNXT":  {Sym: KeyPgDown, Mod: ModShift},
		"kNXT3": {Sym: KeyPgDown, Mod: ModAlt},
		"kNXT4": {Sym: KeyPgDown, Mod: ModShift | ModAlt},
		"kNXT5": {Sym: KeyPgDown, Mod: ModCtrl},
		"kNXT6": {Sym: KeyPgDown, Mod: ModShift | ModCtrl},
		"kNXT7": {Sym: KeyPgDown, Mod: ModAlt | ModCtrl},
		"kNXT8": {Sym: KeyPgDown, Mod: ModShift | ModAlt | ModCtrl},

		"kbs":  {Sym: KeyBackspace},
		"kcbt": {Sym: KeyTab, Mod: ModShift},

		// Function keys
		// This only includes the first 12 function keys. The rest are treated
		// as modifiers of the first 12.
		// Take a look at XTerm modifyFunctionKeys
		//
		// XXX: To use unambiguous function keys, use fixterms or kitty clipboard.
		//
		// See https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyFunctionKeys
		// See https://invisible-island.net/xterm/terminfo.html

		"kf1":  {Sym: KeyF1},
		"kf2":  {Sym: KeyF2},
		"kf3":  {Sym: KeyF3},
		"kf4":  {Sym: KeyF4},
		"kf5":  {Sym: KeyF5},
		"kf6":  {Sym: KeyF6},
		"kf7":  {Sym: KeyF7},
		"kf8":  {Sym: KeyF8},
		"kf9":  {Sym: KeyF9},
		"kf10": {Sym: KeyF10},
		"kf11": {Sym: KeyF11},
		"kf12": {Sym: KeyF12},
		"kf13": {Sym: KeyF1, Mod: ModShift},
		"kf14": {Sym: KeyF2, Mod: ModShift},
		"kf15": {Sym: KeyF3, Mod: ModShift},
		"kf16": {Sym: KeyF4, Mod: ModShift},
		"kf17": {Sym: KeyF5, Mod: ModShift},
		"kf18": {Sym: KeyF6, Mod: ModShift},
		"kf19": {Sym: KeyF7, Mod: ModShift},
		"kf20": {Sym: KeyF8, Mod: ModShift},
		"kf21": {Sym: KeyF9, Mod: ModShift},
		"kf22": {Sym: KeyF10, Mod: ModShift},
		"kf23": {Sym: KeyF11, Mod: ModShift},
		"kf24": {Sym: KeyF12, Mod: ModShift},
		"kf25": {Sym: KeyF1, Mod: ModCtrl},
		"kf26": {Sym: KeyF2, Mod: ModCtrl},
		"kf27": {Sym: KeyF3, Mod: ModCtrl},
		"kf28": {Sym: KeyF4, Mod: ModCtrl},
		"kf29": {Sym: KeyF5, Mod: ModCtrl},
		"kf30": {Sym: KeyF6, Mod: ModCtrl},
		"kf31": {Sym: KeyF7, Mod: ModCtrl},
		"kf32": {Sym: KeyF8, Mod: ModCtrl},
		"kf33": {Sym: KeyF9, Mod: ModCtrl},
		"kf34": {Sym: KeyF10, Mod: ModCtrl},
		"kf35": {Sym: KeyF11, Mod: ModCtrl},
		"kf36": {Sym: KeyF12, Mod: ModCtrl},
		"kf37": {Sym: KeyF1, Mod: ModShift | ModCtrl},
		"kf38": {Sym: KeyF2, Mod: ModShift | ModCtrl},
		"kf39": {Sym: KeyF3, Mod: ModShift | ModCtrl},
		"kf40": {Sym: KeyF4, Mod: ModShift | ModCtrl},
		"kf41": {Sym: KeyF5, Mod: ModShift | ModCtrl},
		"kf42": {Sym: KeyF6, Mod: ModShift | ModCtrl},
		"kf43": {Sym: KeyF7, Mod: ModShift | ModCtrl},
		"kf44": {Sym: KeyF8, Mod: ModShift | ModCtrl},
		"kf45": {Sym: KeyF9, Mod: ModShift | ModCtrl},
		"kf46": {Sym: KeyF10, Mod: ModShift | ModCtrl},
		"kf47": {Sym: KeyF11, Mod: ModShift | ModCtrl},
		"kf48": {Sym: KeyF12, Mod: ModShift | ModCtrl},
		"kf49": {Sym: KeyF1, Mod: ModAlt},
		"kf50": {Sym: KeyF2, Mod: ModAlt},
		"kf51": {Sym: KeyF3, Mod: ModAlt},
		"kf52": {Sym: KeyF4, Mod: ModAlt},
		"kf53": {Sym: KeyF5, Mod: ModAlt},
		"kf54": {Sym: KeyF6, Mod: ModAlt},
		"kf55": {Sym: KeyF7, Mod: ModAlt},
		"kf56": {Sym: KeyF8, Mod: ModAlt},
		"kf57": {Sym: KeyF9, Mod: ModAlt},
		"kf58": {Sym: KeyF10, Mod: ModAlt},
		"kf59": {Sym: KeyF11, Mod: ModAlt},
		"kf60": {Sym: KeyF12, Mod: ModAlt},
		"kf61": {Sym: KeyF1, Mod: ModShift | ModAlt},
		"kf62": {Sym: KeyF2, Mod: ModShift | ModAlt},
		"kf63": {Sym: KeyF3, Mod: ModShift | ModAlt},
	}

	// Preserve F keys from F13 to F63 instead of using them for F-keys
	// modifiers.
	if flags&_FlagFKeys != 0 {
		keys["kf13"] = Key{Sym: KeyF13}
		keys["kf14"] = Key{Sym: KeyF14}
		keys["kf15"] = Key{Sym: KeyF15}
		keys["kf16"] = Key{Sym: KeyF16}
		keys["kf17"] = Key{Sym: KeyF17}
		keys["kf18"] = Key{Sym: KeyF18}
		keys["kf19"] = Key{Sym: KeyF19}
		keys["kf20"] = Key{Sym: KeyF20}
		keys["kf21"] = Key{Sym: KeyF21}
		keys["kf22"] = Key{Sym: KeyF22}
		keys["kf23"] = Key{Sym: KeyF23}
		keys["kf24"] = Key{Sym: KeyF24}
		keys["kf25"] = Key{Sym: KeyF25}
		keys["kf26"] = Key{Sym: KeyF26}
		keys["kf27"] = Key{Sym: KeyF27}
		keys["kf28"] = Key{Sym: KeyF28}
		keys["kf29"] = Key{Sym: KeyF29}
		keys["kf30"] = Key{Sym: KeyF30}
		keys["kf31"] = Key{Sym: KeyF31}
		keys["kf32"] = Key{Sym: KeyF32}
		keys["kf33"] = Key{Sym: KeyF33}
		keys["kf34"] = Key{Sym: KeyF34}
		keys["kf35"] = Key{Sym: KeyF35}
		keys["kf36"] = Key{Sym: KeyF36}
		keys["kf37"] = Key{Sym: KeyF37}
		keys["kf38"] = Key{Sym: KeyF38}
		keys["kf39"] = Key{Sym: KeyF39}
		keys["kf40"] = Key{Sym: KeyF40}
		keys["kf41"] = Key{Sym: KeyF41}
		keys["kf42"] = Key{Sym: KeyF42}
		keys["kf43"] = Key{Sym: KeyF43}
		keys["kf44"] = Key{Sym: KeyF44}
		keys["kf45"] = Key{Sym: KeyF45}
		keys["kf46"] = Key{Sym: KeyF46}
		keys["kf47"] = Key{Sym: KeyF47}
		keys["kf48"] = Key{Sym: KeyF48}
		keys["kf49"] = Key{Sym: KeyF49}
		keys["kf50"] = Key{Sym: KeyF50}
		keys["kf51"] = Key{Sym: KeyF51}
		keys["kf52"] = Key{Sym: KeyF52}
		keys["kf53"] = Key{Sym: KeyF53}
		keys["kf54"] = Key{Sym: KeyF54}
		keys["kf55"] = Key{Sym: KeyF55}
		keys["kf56"] = Key{Sym: KeyF56}
		keys["kf57"] = Key{Sym: KeyF57}
		keys["kf58"] = Key{Sym: KeyF58}
		keys["kf59"] = Key{Sym: KeyF59}
		keys["kf60"] = Key{Sym: KeyF60}
		keys["kf61"] = Key{Sym: KeyF61}
		keys["kf62"] = Key{Sym: KeyF62}
		keys["kf63"] = Key{Sym: KeyF63}
	}

	return keys
}
