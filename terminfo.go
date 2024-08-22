package tea

import (
	"strings"

	"github.com/xo/terminfo"
)

func buildTerminfoKeys(flags int, term string) map[string]key {
	table := make(map[string]key)
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
func defaultTerminfoKeys(flags int) map[string]key {
	keys := map[string]key{
		"kcuu1": {typ: KeyUp},
		"kUP":   {typ: KeyUp, mod: ModShift},
		"kUP3":  {typ: KeyUp, mod: ModAlt},
		"kUP4":  {typ: KeyUp, mod: ModShift | ModAlt},
		"kUP5":  {typ: KeyUp, mod: ModCtrl},
		"kUP6":  {typ: KeyUp, mod: ModShift | ModCtrl},
		"kUP7":  {typ: KeyUp, mod: ModAlt | ModCtrl},
		"kUP8":  {typ: KeyUp, mod: ModShift | ModAlt | ModCtrl},
		"kcud1": {typ: KeyDown},
		"kDN":   {typ: KeyDown, mod: ModShift},
		"kDN3":  {typ: KeyDown, mod: ModAlt},
		"kDN4":  {typ: KeyDown, mod: ModShift | ModAlt},
		"kDN5":  {typ: KeyDown, mod: ModCtrl},
		"kDN7":  {typ: KeyDown, mod: ModAlt | ModCtrl},
		"kDN6":  {typ: KeyDown, mod: ModShift | ModCtrl},
		"kDN8":  {typ: KeyDown, mod: ModShift | ModAlt | ModCtrl},
		"kcub1": {typ: KeyLeft},
		"kLFT":  {typ: KeyLeft, mod: ModShift},
		"kLFT3": {typ: KeyLeft, mod: ModAlt},
		"kLFT4": {typ: KeyLeft, mod: ModShift | ModAlt},
		"kLFT5": {typ: KeyLeft, mod: ModCtrl},
		"kLFT6": {typ: KeyLeft, mod: ModShift | ModCtrl},
		"kLFT7": {typ: KeyLeft, mod: ModAlt | ModCtrl},
		"kLFT8": {typ: KeyLeft, mod: ModShift | ModAlt | ModCtrl},
		"kcuf1": {typ: KeyRight},
		"kRIT":  {typ: KeyRight, mod: ModShift},
		"kRIT3": {typ: KeyRight, mod: ModAlt},
		"kRIT4": {typ: KeyRight, mod: ModShift | ModAlt},
		"kRIT5": {typ: KeyRight, mod: ModCtrl},
		"kRIT6": {typ: KeyRight, mod: ModShift | ModCtrl},
		"kRIT7": {typ: KeyRight, mod: ModAlt | ModCtrl},
		"kRIT8": {typ: KeyRight, mod: ModShift | ModAlt | ModCtrl},
		"kich1": {typ: KeyInsert},
		"kIC":   {typ: KeyInsert, mod: ModShift},
		"kIC3":  {typ: KeyInsert, mod: ModAlt},
		"kIC4":  {typ: KeyInsert, mod: ModShift | ModAlt},
		"kIC5":  {typ: KeyInsert, mod: ModCtrl},
		"kIC6":  {typ: KeyInsert, mod: ModShift | ModCtrl},
		"kIC7":  {typ: KeyInsert, mod: ModAlt | ModCtrl},
		"kIC8":  {typ: KeyInsert, mod: ModShift | ModAlt | ModCtrl},
		"kdch1": {typ: KeyDelete},
		"kDC":   {typ: KeyDelete, mod: ModShift},
		"kDC3":  {typ: KeyDelete, mod: ModAlt},
		"kDC4":  {typ: KeyDelete, mod: ModShift | ModAlt},
		"kDC5":  {typ: KeyDelete, mod: ModCtrl},
		"kDC6":  {typ: KeyDelete, mod: ModShift | ModCtrl},
		"kDC7":  {typ: KeyDelete, mod: ModAlt | ModCtrl},
		"kDC8":  {typ: KeyDelete, mod: ModShift | ModAlt | ModCtrl},
		"khome": {typ: KeyHome},
		"kHOM":  {typ: KeyHome, mod: ModShift},
		"kHOM3": {typ: KeyHome, mod: ModAlt},
		"kHOM4": {typ: KeyHome, mod: ModShift | ModAlt},
		"kHOM5": {typ: KeyHome, mod: ModCtrl},
		"kHOM6": {typ: KeyHome, mod: ModShift | ModCtrl},
		"kHOM7": {typ: KeyHome, mod: ModAlt | ModCtrl},
		"kHOM8": {typ: KeyHome, mod: ModShift | ModAlt | ModCtrl},
		"kend":  {typ: KeyEnd},
		"kEND":  {typ: KeyEnd, mod: ModShift},
		"kEND3": {typ: KeyEnd, mod: ModAlt},
		"kEND4": {typ: KeyEnd, mod: ModShift | ModAlt},
		"kEND5": {typ: KeyEnd, mod: ModCtrl},
		"kEND6": {typ: KeyEnd, mod: ModShift | ModCtrl},
		"kEND7": {typ: KeyEnd, mod: ModAlt | ModCtrl},
		"kEND8": {typ: KeyEnd, mod: ModShift | ModAlt | ModCtrl},
		"kpp":   {typ: KeyPgUp},
		"kprv":  {typ: KeyPgUp},
		"kPRV":  {typ: KeyPgUp, mod: ModShift},
		"kPRV3": {typ: KeyPgUp, mod: ModAlt},
		"kPRV4": {typ: KeyPgUp, mod: ModShift | ModAlt},
		"kPRV5": {typ: KeyPgUp, mod: ModCtrl},
		"kPRV6": {typ: KeyPgUp, mod: ModShift | ModCtrl},
		"kPRV7": {typ: KeyPgUp, mod: ModAlt | ModCtrl},
		"kPRV8": {typ: KeyPgUp, mod: ModShift | ModAlt | ModCtrl},
		"knp":   {typ: KeyPgDown},
		"knxt":  {typ: KeyPgDown},
		"kNXT":  {typ: KeyPgDown, mod: ModShift},
		"kNXT3": {typ: KeyPgDown, mod: ModAlt},
		"kNXT4": {typ: KeyPgDown, mod: ModShift | ModAlt},
		"kNXT5": {typ: KeyPgDown, mod: ModCtrl},
		"kNXT6": {typ: KeyPgDown, mod: ModShift | ModCtrl},
		"kNXT7": {typ: KeyPgDown, mod: ModAlt | ModCtrl},
		"kNXT8": {typ: KeyPgDown, mod: ModShift | ModAlt | ModCtrl},

		"kbs":  {typ: KeyBackspace},
		"kcbt": {typ: KeyTab, mod: ModShift},

		// Function keys
		// This only includes the first 12 function keys. The rest are treated
		// as modifiers of the first 12.
		// Take a look at XTerm modifyFunctionKeys
		//
		// XXX: To use unambiguous function keys, use fixterms or kitty clipboard.
		//
		// See https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyFunctionKeys
		// See https://invisible-island.net/xterm/terminfo.html

		"kf1":  {typ: KeyF1},
		"kf2":  {typ: KeyF2},
		"kf3":  {typ: KeyF3},
		"kf4":  {typ: KeyF4},
		"kf5":  {typ: KeyF5},
		"kf6":  {typ: KeyF6},
		"kf7":  {typ: KeyF7},
		"kf8":  {typ: KeyF8},
		"kf9":  {typ: KeyF9},
		"kf10": {typ: KeyF10},
		"kf11": {typ: KeyF11},
		"kf12": {typ: KeyF12},
		"kf13": {typ: KeyF1, mod: ModShift},
		"kf14": {typ: KeyF2, mod: ModShift},
		"kf15": {typ: KeyF3, mod: ModShift},
		"kf16": {typ: KeyF4, mod: ModShift},
		"kf17": {typ: KeyF5, mod: ModShift},
		"kf18": {typ: KeyF6, mod: ModShift},
		"kf19": {typ: KeyF7, mod: ModShift},
		"kf20": {typ: KeyF8, mod: ModShift},
		"kf21": {typ: KeyF9, mod: ModShift},
		"kf22": {typ: KeyF10, mod: ModShift},
		"kf23": {typ: KeyF11, mod: ModShift},
		"kf24": {typ: KeyF12, mod: ModShift},
		"kf25": {typ: KeyF1, mod: ModCtrl},
		"kf26": {typ: KeyF2, mod: ModCtrl},
		"kf27": {typ: KeyF3, mod: ModCtrl},
		"kf28": {typ: KeyF4, mod: ModCtrl},
		"kf29": {typ: KeyF5, mod: ModCtrl},
		"kf30": {typ: KeyF6, mod: ModCtrl},
		"kf31": {typ: KeyF7, mod: ModCtrl},
		"kf32": {typ: KeyF8, mod: ModCtrl},
		"kf33": {typ: KeyF9, mod: ModCtrl},
		"kf34": {typ: KeyF10, mod: ModCtrl},
		"kf35": {typ: KeyF11, mod: ModCtrl},
		"kf36": {typ: KeyF12, mod: ModCtrl},
		"kf37": {typ: KeyF1, mod: ModShift | ModCtrl},
		"kf38": {typ: KeyF2, mod: ModShift | ModCtrl},
		"kf39": {typ: KeyF3, mod: ModShift | ModCtrl},
		"kf40": {typ: KeyF4, mod: ModShift | ModCtrl},
		"kf41": {typ: KeyF5, mod: ModShift | ModCtrl},
		"kf42": {typ: KeyF6, mod: ModShift | ModCtrl},
		"kf43": {typ: KeyF7, mod: ModShift | ModCtrl},
		"kf44": {typ: KeyF8, mod: ModShift | ModCtrl},
		"kf45": {typ: KeyF9, mod: ModShift | ModCtrl},
		"kf46": {typ: KeyF10, mod: ModShift | ModCtrl},
		"kf47": {typ: KeyF11, mod: ModShift | ModCtrl},
		"kf48": {typ: KeyF12, mod: ModShift | ModCtrl},
		"kf49": {typ: KeyF1, mod: ModAlt},
		"kf50": {typ: KeyF2, mod: ModAlt},
		"kf51": {typ: KeyF3, mod: ModAlt},
		"kf52": {typ: KeyF4, mod: ModAlt},
		"kf53": {typ: KeyF5, mod: ModAlt},
		"kf54": {typ: KeyF6, mod: ModAlt},
		"kf55": {typ: KeyF7, mod: ModAlt},
		"kf56": {typ: KeyF8, mod: ModAlt},
		"kf57": {typ: KeyF9, mod: ModAlt},
		"kf58": {typ: KeyF10, mod: ModAlt},
		"kf59": {typ: KeyF11, mod: ModAlt},
		"kf60": {typ: KeyF12, mod: ModAlt},
		"kf61": {typ: KeyF1, mod: ModShift | ModAlt},
		"kf62": {typ: KeyF2, mod: ModShift | ModAlt},
		"kf63": {typ: KeyF3, mod: ModShift | ModAlt},
	}

	// Preserve F keys from F13 to F63 instead of using them for F-keys
	// modifiers.
	if flags&_FlagFKeys != 0 {
		keys["kf13"] = key{typ: KeyF13}
		keys["kf14"] = key{typ: KeyF14}
		keys["kf15"] = key{typ: KeyF15}
		keys["kf16"] = key{typ: KeyF16}
		keys["kf17"] = key{typ: KeyF17}
		keys["kf18"] = key{typ: KeyF18}
		keys["kf19"] = key{typ: KeyF19}
		keys["kf20"] = key{typ: KeyF20}
		keys["kf21"] = key{typ: KeyF21}
		keys["kf22"] = key{typ: KeyF22}
		keys["kf23"] = key{typ: KeyF23}
		keys["kf24"] = key{typ: KeyF24}
		keys["kf25"] = key{typ: KeyF25}
		keys["kf26"] = key{typ: KeyF26}
		keys["kf27"] = key{typ: KeyF27}
		keys["kf28"] = key{typ: KeyF28}
		keys["kf29"] = key{typ: KeyF29}
		keys["kf30"] = key{typ: KeyF30}
		keys["kf31"] = key{typ: KeyF31}
		keys["kf32"] = key{typ: KeyF32}
		keys["kf33"] = key{typ: KeyF33}
		keys["kf34"] = key{typ: KeyF34}
		keys["kf35"] = key{typ: KeyF35}
		keys["kf36"] = key{typ: KeyF36}
		keys["kf37"] = key{typ: KeyF37}
		keys["kf38"] = key{typ: KeyF38}
		keys["kf39"] = key{typ: KeyF39}
		keys["kf40"] = key{typ: KeyF40}
		keys["kf41"] = key{typ: KeyF41}
		keys["kf42"] = key{typ: KeyF42}
		keys["kf43"] = key{typ: KeyF43}
		keys["kf44"] = key{typ: KeyF44}
		keys["kf45"] = key{typ: KeyF45}
		keys["kf46"] = key{typ: KeyF46}
		keys["kf47"] = key{typ: KeyF47}
		keys["kf48"] = key{typ: KeyF48}
		keys["kf49"] = key{typ: KeyF49}
		keys["kf50"] = key{typ: KeyF50}
		keys["kf51"] = key{typ: KeyF51}
		keys["kf52"] = key{typ: KeyF52}
		keys["kf53"] = key{typ: KeyF53}
		keys["kf54"] = key{typ: KeyF54}
		keys["kf55"] = key{typ: KeyF55}
		keys["kf56"] = key{typ: KeyF56}
		keys["kf57"] = key{typ: KeyF57}
		keys["kf58"] = key{typ: KeyF58}
		keys["kf59"] = key{typ: KeyF59}
		keys["kf60"] = key{typ: KeyF60}
		keys["kf61"] = key{typ: KeyF61}
		keys["kf62"] = key{typ: KeyF62}
		keys["kf63"] = key{typ: KeyF63}
	}

	return keys
}
