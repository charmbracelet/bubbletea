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
		"kcuu1": {Type: KeyUp},
		"kUP":   {Type: KeyUp, Mod: ModShift},
		"kUP3":  {Type: KeyUp, Mod: ModAlt},
		"kUP4":  {Type: KeyUp, Mod: ModShift | ModAlt},
		"kUP5":  {Type: KeyUp, Mod: ModCtrl},
		"kUP6":  {Type: KeyUp, Mod: ModShift | ModCtrl},
		"kUP7":  {Type: KeyUp, Mod: ModAlt | ModCtrl},
		"kUP8":  {Type: KeyUp, Mod: ModShift | ModAlt | ModCtrl},
		"kcud1": {Type: KeyDown},
		"kDN":   {Type: KeyDown, Mod: ModShift},
		"kDN3":  {Type: KeyDown, Mod: ModAlt},
		"kDN4":  {Type: KeyDown, Mod: ModShift | ModAlt},
		"kDN5":  {Type: KeyDown, Mod: ModCtrl},
		"kDN7":  {Type: KeyDown, Mod: ModAlt | ModCtrl},
		"kDN6":  {Type: KeyDown, Mod: ModShift | ModCtrl},
		"kDN8":  {Type: KeyDown, Mod: ModShift | ModAlt | ModCtrl},
		"kcub1": {Type: KeyLeft},
		"kLFT":  {Type: KeyLeft, Mod: ModShift},
		"kLFT3": {Type: KeyLeft, Mod: ModAlt},
		"kLFT4": {Type: KeyLeft, Mod: ModShift | ModAlt},
		"kLFT5": {Type: KeyLeft, Mod: ModCtrl},
		"kLFT6": {Type: KeyLeft, Mod: ModShift | ModCtrl},
		"kLFT7": {Type: KeyLeft, Mod: ModAlt | ModCtrl},
		"kLFT8": {Type: KeyLeft, Mod: ModShift | ModAlt | ModCtrl},
		"kcuf1": {Type: KeyRight},
		"kRIT":  {Type: KeyRight, Mod: ModShift},
		"kRIT3": {Type: KeyRight, Mod: ModAlt},
		"kRIT4": {Type: KeyRight, Mod: ModShift | ModAlt},
		"kRIT5": {Type: KeyRight, Mod: ModCtrl},
		"kRIT6": {Type: KeyRight, Mod: ModShift | ModCtrl},
		"kRIT7": {Type: KeyRight, Mod: ModAlt | ModCtrl},
		"kRIT8": {Type: KeyRight, Mod: ModShift | ModAlt | ModCtrl},
		"kich1": {Type: KeyInsert},
		"kIC":   {Type: KeyInsert, Mod: ModShift},
		"kIC3":  {Type: KeyInsert, Mod: ModAlt},
		"kIC4":  {Type: KeyInsert, Mod: ModShift | ModAlt},
		"kIC5":  {Type: KeyInsert, Mod: ModCtrl},
		"kIC6":  {Type: KeyInsert, Mod: ModShift | ModCtrl},
		"kIC7":  {Type: KeyInsert, Mod: ModAlt | ModCtrl},
		"kIC8":  {Type: KeyInsert, Mod: ModShift | ModAlt | ModCtrl},
		"kdch1": {Type: KeyDelete},
		"kDC":   {Type: KeyDelete, Mod: ModShift},
		"kDC3":  {Type: KeyDelete, Mod: ModAlt},
		"kDC4":  {Type: KeyDelete, Mod: ModShift | ModAlt},
		"kDC5":  {Type: KeyDelete, Mod: ModCtrl},
		"kDC6":  {Type: KeyDelete, Mod: ModShift | ModCtrl},
		"kDC7":  {Type: KeyDelete, Mod: ModAlt | ModCtrl},
		"kDC8":  {Type: KeyDelete, Mod: ModShift | ModAlt | ModCtrl},
		"khome": {Type: KeyHome},
		"kHOM":  {Type: KeyHome, Mod: ModShift},
		"kHOM3": {Type: KeyHome, Mod: ModAlt},
		"kHOM4": {Type: KeyHome, Mod: ModShift | ModAlt},
		"kHOM5": {Type: KeyHome, Mod: ModCtrl},
		"kHOM6": {Type: KeyHome, Mod: ModShift | ModCtrl},
		"kHOM7": {Type: KeyHome, Mod: ModAlt | ModCtrl},
		"kHOM8": {Type: KeyHome, Mod: ModShift | ModAlt | ModCtrl},
		"kend":  {Type: KeyEnd},
		"kEND":  {Type: KeyEnd, Mod: ModShift},
		"kEND3": {Type: KeyEnd, Mod: ModAlt},
		"kEND4": {Type: KeyEnd, Mod: ModShift | ModAlt},
		"kEND5": {Type: KeyEnd, Mod: ModCtrl},
		"kEND6": {Type: KeyEnd, Mod: ModShift | ModCtrl},
		"kEND7": {Type: KeyEnd, Mod: ModAlt | ModCtrl},
		"kEND8": {Type: KeyEnd, Mod: ModShift | ModAlt | ModCtrl},
		"kpp":   {Type: KeyPgUp},
		"kprv":  {Type: KeyPgUp},
		"kPRV":  {Type: KeyPgUp, Mod: ModShift},
		"kPRV3": {Type: KeyPgUp, Mod: ModAlt},
		"kPRV4": {Type: KeyPgUp, Mod: ModShift | ModAlt},
		"kPRV5": {Type: KeyPgUp, Mod: ModCtrl},
		"kPRV6": {Type: KeyPgUp, Mod: ModShift | ModCtrl},
		"kPRV7": {Type: KeyPgUp, Mod: ModAlt | ModCtrl},
		"kPRV8": {Type: KeyPgUp, Mod: ModShift | ModAlt | ModCtrl},
		"knp":   {Type: KeyPgDown},
		"knxt":  {Type: KeyPgDown},
		"kNXT":  {Type: KeyPgDown, Mod: ModShift},
		"kNXT3": {Type: KeyPgDown, Mod: ModAlt},
		"kNXT4": {Type: KeyPgDown, Mod: ModShift | ModAlt},
		"kNXT5": {Type: KeyPgDown, Mod: ModCtrl},
		"kNXT6": {Type: KeyPgDown, Mod: ModShift | ModCtrl},
		"kNXT7": {Type: KeyPgDown, Mod: ModAlt | ModCtrl},
		"kNXT8": {Type: KeyPgDown, Mod: ModShift | ModAlt | ModCtrl},

		"kbs":  {Type: KeyBackspace},
		"kcbt": {Type: KeyTab, Mod: ModShift},

		// Function keys
		// This only includes the first 12 function keys. The rest are treated
		// as modifiers of the first 12.
		// Take a look at XTerm modifyFunctionKeys
		//
		// XXX: To use unambiguous function keys, use fixterms or kitty clipboard.
		//
		// See https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyFunctionKeys
		// See https://invisible-island.net/xterm/terminfo.html

		"kf1":  {Type: KeyF1},
		"kf2":  {Type: KeyF2},
		"kf3":  {Type: KeyF3},
		"kf4":  {Type: KeyF4},
		"kf5":  {Type: KeyF5},
		"kf6":  {Type: KeyF6},
		"kf7":  {Type: KeyF7},
		"kf8":  {Type: KeyF8},
		"kf9":  {Type: KeyF9},
		"kf10": {Type: KeyF10},
		"kf11": {Type: KeyF11},
		"kf12": {Type: KeyF12},
		"kf13": {Type: KeyF1, Mod: ModShift},
		"kf14": {Type: KeyF2, Mod: ModShift},
		"kf15": {Type: KeyF3, Mod: ModShift},
		"kf16": {Type: KeyF4, Mod: ModShift},
		"kf17": {Type: KeyF5, Mod: ModShift},
		"kf18": {Type: KeyF6, Mod: ModShift},
		"kf19": {Type: KeyF7, Mod: ModShift},
		"kf20": {Type: KeyF8, Mod: ModShift},
		"kf21": {Type: KeyF9, Mod: ModShift},
		"kf22": {Type: KeyF10, Mod: ModShift},
		"kf23": {Type: KeyF11, Mod: ModShift},
		"kf24": {Type: KeyF12, Mod: ModShift},
		"kf25": {Type: KeyF1, Mod: ModCtrl},
		"kf26": {Type: KeyF2, Mod: ModCtrl},
		"kf27": {Type: KeyF3, Mod: ModCtrl},
		"kf28": {Type: KeyF4, Mod: ModCtrl},
		"kf29": {Type: KeyF5, Mod: ModCtrl},
		"kf30": {Type: KeyF6, Mod: ModCtrl},
		"kf31": {Type: KeyF7, Mod: ModCtrl},
		"kf32": {Type: KeyF8, Mod: ModCtrl},
		"kf33": {Type: KeyF9, Mod: ModCtrl},
		"kf34": {Type: KeyF10, Mod: ModCtrl},
		"kf35": {Type: KeyF11, Mod: ModCtrl},
		"kf36": {Type: KeyF12, Mod: ModCtrl},
		"kf37": {Type: KeyF1, Mod: ModShift | ModCtrl},
		"kf38": {Type: KeyF2, Mod: ModShift | ModCtrl},
		"kf39": {Type: KeyF3, Mod: ModShift | ModCtrl},
		"kf40": {Type: KeyF4, Mod: ModShift | ModCtrl},
		"kf41": {Type: KeyF5, Mod: ModShift | ModCtrl},
		"kf42": {Type: KeyF6, Mod: ModShift | ModCtrl},
		"kf43": {Type: KeyF7, Mod: ModShift | ModCtrl},
		"kf44": {Type: KeyF8, Mod: ModShift | ModCtrl},
		"kf45": {Type: KeyF9, Mod: ModShift | ModCtrl},
		"kf46": {Type: KeyF10, Mod: ModShift | ModCtrl},
		"kf47": {Type: KeyF11, Mod: ModShift | ModCtrl},
		"kf48": {Type: KeyF12, Mod: ModShift | ModCtrl},
		"kf49": {Type: KeyF1, Mod: ModAlt},
		"kf50": {Type: KeyF2, Mod: ModAlt},
		"kf51": {Type: KeyF3, Mod: ModAlt},
		"kf52": {Type: KeyF4, Mod: ModAlt},
		"kf53": {Type: KeyF5, Mod: ModAlt},
		"kf54": {Type: KeyF6, Mod: ModAlt},
		"kf55": {Type: KeyF7, Mod: ModAlt},
		"kf56": {Type: KeyF8, Mod: ModAlt},
		"kf57": {Type: KeyF9, Mod: ModAlt},
		"kf58": {Type: KeyF10, Mod: ModAlt},
		"kf59": {Type: KeyF11, Mod: ModAlt},
		"kf60": {Type: KeyF12, Mod: ModAlt},
		"kf61": {Type: KeyF1, Mod: ModShift | ModAlt},
		"kf62": {Type: KeyF2, Mod: ModShift | ModAlt},
		"kf63": {Type: KeyF3, Mod: ModShift | ModAlt},
	}

	// Preserve F keys from F13 to F63 instead of using them for F-keys
	// modifiers.
	if flags&_FlagFKeys != 0 {
		keys["kf13"] = Key{Type: KeyF13}
		keys["kf14"] = Key{Type: KeyF14}
		keys["kf15"] = Key{Type: KeyF15}
		keys["kf16"] = Key{Type: KeyF16}
		keys["kf17"] = Key{Type: KeyF17}
		keys["kf18"] = Key{Type: KeyF18}
		keys["kf19"] = Key{Type: KeyF19}
		keys["kf20"] = Key{Type: KeyF20}
		keys["kf21"] = Key{Type: KeyF21}
		keys["kf22"] = Key{Type: KeyF22}
		keys["kf23"] = Key{Type: KeyF23}
		keys["kf24"] = Key{Type: KeyF24}
		keys["kf25"] = Key{Type: KeyF25}
		keys["kf26"] = Key{Type: KeyF26}
		keys["kf27"] = Key{Type: KeyF27}
		keys["kf28"] = Key{Type: KeyF28}
		keys["kf29"] = Key{Type: KeyF29}
		keys["kf30"] = Key{Type: KeyF30}
		keys["kf31"] = Key{Type: KeyF31}
		keys["kf32"] = Key{Type: KeyF32}
		keys["kf33"] = Key{Type: KeyF33}
		keys["kf34"] = Key{Type: KeyF34}
		keys["kf35"] = Key{Type: KeyF35}
		keys["kf36"] = Key{Type: KeyF36}
		keys["kf37"] = Key{Type: KeyF37}
		keys["kf38"] = Key{Type: KeyF38}
		keys["kf39"] = Key{Type: KeyF39}
		keys["kf40"] = Key{Type: KeyF40}
		keys["kf41"] = Key{Type: KeyF41}
		keys["kf42"] = Key{Type: KeyF42}
		keys["kf43"] = Key{Type: KeyF43}
		keys["kf44"] = Key{Type: KeyF44}
		keys["kf45"] = Key{Type: KeyF45}
		keys["kf46"] = Key{Type: KeyF46}
		keys["kf47"] = Key{Type: KeyF47}
		keys["kf48"] = Key{Type: KeyF48}
		keys["kf49"] = Key{Type: KeyF49}
		keys["kf50"] = Key{Type: KeyF50}
		keys["kf51"] = Key{Type: KeyF51}
		keys["kf52"] = Key{Type: KeyF52}
		keys["kf53"] = Key{Type: KeyF53}
		keys["kf54"] = Key{Type: KeyF54}
		keys["kf55"] = Key{Type: KeyF55}
		keys["kf56"] = Key{Type: KeyF56}
		keys["kf57"] = Key{Type: KeyF57}
		keys["kf58"] = Key{Type: KeyF58}
		keys["kf59"] = Key{Type: KeyF59}
		keys["kf60"] = Key{Type: KeyF60}
		keys["kf61"] = Key{Type: KeyF61}
		keys["kf62"] = Key{Type: KeyF62}
		keys["kf63"] = Key{Type: KeyF63}
	}

	return keys
}
