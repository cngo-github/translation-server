// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NOTE: This package is still under development. Parts of it are not yet implemented,
// and the API is subject to change.
//
// The language package provides a type to represent language tags based on BCP 47.
// It supports various canonicalizations defined in CLDR.
// See http://tools.ietf.org/html/bcp47 for more details.
package language

import "strings"

var (
	// Und represents the undertermined language. It is also the root language tag.
	Und   Tag = und
	En    Tag = en    // Default language tag for English.
	En_US Tag = en_US // Default language tag for American English.
	De    Tag = de    // Default language tag for German.
	// TODO: list of most common language tags.
)

var (
	Supported Set // All supported language indetifiers.
	Common    Set // A selection of common language indetifiers.
)

var (
	de    = Tag{lang: lang_de}
	en    = Tag{lang: lang_en}
	en_US = Tag{lang: lang_en, region: regUS}
	und   = Tag{}
)

// Tag represents a BCP 47 language tag. It is used to specifify
// an instance of a specific language or locale.
// All language tag values are guaranteed to be well-formed.
type Tag struct {
	// In most cases, just lang, region and script will be needed.  In such cases
	// str may be nil.
	lang     langID
	region   regionID
	script   scriptID
	pVariant byte   // offset in str, includes preceding '-'
	pExt     uint16 // offset of first extension, includes preceding '-'
	str      *string
}

// Make calls Parse and Canonicalize and returns the resulting Tag.
// Any errors are ignored and a sensible default is returned.
// In most cases, language tags should be created using this method.
func Make(id string) Tag {
	loc, _ := Parse(id)
	loc, _ = loc.Canonicalize(Default)
	return loc
}

// equalTags compares language, script and region subtags only.
func (t Tag) equalTags(a Tag) bool {
	return t.lang == a.lang && t.script == a.script && t.region == a.region
}

// IsRoot returns true if t is equal to language "und".
func (t Tag) IsRoot() bool {
	if t.str != nil {
		n := len(*t.str)
		if int(t.pVariant) < n {
			return false
		}
		t.str = nil
	}
	return t.equalTags(und)
}

// private reports whether the Tag consists solely of a private use tag.
func (t Tag) private() bool {
	return t.str != nil && t.pVariant == 0
}

// CanonType can be used to enable or disable various types of canonicalization.
type CanonType int

const (
	// Replace deprecated values with their preferred ones.
	Deprecated CanonType = 1 << iota
	// Remove redundant scripts.
	SuppressScript
	// Normalize legacy encodings, as defined by CLDR.
	Legacy
	// Map the dominant language of a macro language group to the macro language subtag.
	// For example cmn -> zh.
	Macro
	// The CLDR flag should be used if full compatibility with CLDR is required.  There are
	// a few cases where language.Tag may differ from CLDR.
	CLDR
	// All canonicalizations prescribed by BCP 47.
	BCP47   = Deprecated | SuppressScript
	All     = BCP47 | Legacy | Macro
	Default = All

	// TODO: LikelyScript, LikelyRegion: supress similar to ICU.
)

// canonicalize returns the canonicalized equivalent of the tag and
// whether there was any change.
func (t Tag) canonicalize(c CanonType) (Tag, bool) {
	changed := false
	if c&SuppressScript != 0 {
		if t.lang < langNoIndexOffset && uint8(t.script) == suppressScript[t.lang] {
			t.script = 0
			changed = true
		}
	}
	if c&Legacy != 0 {
		// We hard code this set as it is very small, unlikely to change and requires some
		// handling that does not fit elsewhere.
		switch t.lang {
		case lang_no:
			if c&CLDR != 0 {
				t.lang = lang_nb
				changed = true
			}
		case lang_tl:
			t.lang = lang_fil
			changed = true
		case lang_sh:
			if t.script == 0 {
				t.script = scrLatn
			}
			t.lang = lang_sr
			changed = true
		}
	}
	if c&Deprecated != 0 {
		l := normLang(langOldMap[:], t.lang)
		if l != t.lang {
			// CLDR maps "mo" to "ro". This mapping loses the piece of information
			// that "mo" very likely implies the region "MD". This may be important
			// for applications that insist on making a difference between these
			// two language codes.
			if t.lang == lang_mo && t.region == 0 && c&CLDR == 0 {
				t.region = regMD
			}
			changed = true
			t.lang = l
		}
	}
	if c&Macro != 0 {
		// We deviate here from CLDR. The mapping "nb" -> "no" qualifies as a typical
		// Macro language mapping.  However, for legacy reasons, CLDR maps "no,
		// the macro language code for Norwegian, to the dominant variant "nb.
		// This change is currently under consideration for CLDR as well.
		// See http://unicode.org/cldr/trac/ticket/2698 and also
		// http://unicode.org/cldr/trac/ticket/1790 for some of the practical
		// implications.
		// TODO: this check could be removed if CLDR adopts this change.
		if c&CLDR == 0 || t.lang != lang_nb {
			l := normLang(langMacroMap[:], t.lang)
			if l != t.lang {
				changed = true
				t.lang = l
			}
		}
	}
	return t, changed
}

// Canonicalize returns the canonicalized equivalent of the tag.
func (t Tag) Canonicalize(c CanonType) (Tag, error) {
	t, changed := t.canonicalize(c)
	if changed && t.str != nil {
		t.remakeString()
	}
	return t, nil
}

// Confidence indicates the level of certainty for a given return value.
// For example, Serbian may be written in cyrillic or latin script.
// The confidence level indicates whether a value was explicitly specified,
// whether it is typically the only possible value, or whether there is
// an ambiguity.
type Confidence int

const (
	No    Confidence = iota // full confidence that there was no match
	Low                     // most likely value picked out of a set of alternatives
	High                    // value is generally assumed to be the correct match
	Exact                   // exact match or explicitly specified value
)

var confName = []string{"No", "Low", "High", "Exact"}

func (c Confidence) String() string {
	return confName[c]
}

// remakeString is used to update t.str in case lang, script or region changed.
// It is assumed that pExt and pVariant still point to the start of the
// respective parts, if applicable.
// remakeString can also be used to compute the string for Tag for which str
// is not defined.
func (t *Tag) remakeString() {
	extra := ""
	if t.str != nil && int(t.pVariant) < len(*t.str) {
		extra = (*t.str)[t.pVariant:]
		if t.pVariant > 0 {
			extra = extra[1:]
		}
	}
	buf := [128]byte{}
	isUnd := t.lang == 0
	n := t.lang.stringToBuf(buf[:])
	if t.script != 0 {
		n += copy(buf[n:], "-")
		n += copy(buf[n:], t.script.String())
		isUnd = false
	}
	if t.region != 0 {
		n += copy(buf[n:], "-")
		n += copy(buf[n:], t.region.String())
		isUnd = false
	}
	b := buf[:n]
	if extra != "" {
		if isUnd && strings.HasPrefix(extra, "x-") {
			t.str = &extra
			t.pVariant = 0
			t.pExt = 0
			return
		} else {
			diff := uint8(n) - t.pVariant
			b = append(b, '-')
			b = append(b, extra...)
			t.pVariant += diff
			t.pExt += uint16(diff)
		}
	} else {
		t.pVariant = uint8(len(b))
		t.pExt = uint16(len(b))
	}
	s := string(b)
	t.str = &s
}

// String returns the canonical string representation of the language tag.
func (t Tag) String() string {
	if t.str == nil {
		t.remakeString()
	}
	return *t.str
}

// Base returns the base language of the language tag. If the base language is
// unspecified, an attempt will be made to infer it from the context.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (t Tag) Base() (Base, Confidence) {
	if t.lang != 0 {
		return Base{t.lang}, Exact
	}
	c := High
	if t.script == 0 && !(Region{t.region}).IsCountry() {
		c = Low
	}
	if tag, err := addTags(t); err == nil && tag.lang != 0 {
		return Base{tag.lang}, c
	}
	return Base{0}, No
}

// Script infers the script for the language tag. If it was not explictly given, it will infer
// a most likely candidate.
// If more than one script is commonly used for a language, the most likely one
// is returned with a low confidence indication. For example, it returns (Cyrl, Low)
// for Serbian.
// Note that an inferred script is never guaranteed to be the correct one. Latin is
// almost exclusively used for Afrikaans, but Arabic has been used for some texts
// in the past.  Also, the script that is commonly used may change over time.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (t Tag) Script() (Script, Confidence) {
	if t.script != 0 {
		return Script{t.script}, Exact
	}
	if t.lang < langNoIndexOffset {
		if sc := suppressScript[t.lang]; sc != 0 {
			return Script{scriptID(sc)}, High
		}
	}
	sc, c := Script{scrZyyy}, No
	if tag, err := addTags(t); err == nil {
		sc, c = Script{tag.script}, Low
	}
	t, _ = t.Canonicalize(Deprecated | Macro)
	if tag, err := addTags(t); err == nil {
		sc, c = Script{tag.script}, Low
	}
	// Translate srcZzzz (uncoded) to srcZyyy (undetermined).
	if sc == (Script{scrZzzz}) {
		return Script{scrZyyy}, No
	}
	return sc, c
}

// Region returns the region for the language tag. If it was not explicitly given, it will
// infer a most likely candidate from the context.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (t Tag) Region() (Region, Confidence) {
	if t.region != 0 {
		return Region{t.region}, Exact
	}
	if t, err := addTags(t); err == nil {
		return Region{t.region}, Low // TODO: differentiate between high and low.
	}
	t, _ = t.Canonicalize(Deprecated | Macro)
	if tag, err := addTags(t); err == nil {
		return Region{tag.region}, Low
	}
	return Region{regZZ}, No // TODO: return world instead of undetermined?
}

// Variant returns the variants specified explicitly for this language tag.
// or nil if no variant was specified.
func (t Tag) Variant() []Variant {
	// TODO: implement
	return nil
}

// TypeForKey returns the type associated with the given key, where key and type
// are of the allowed values defined for the Unicode locale extension ('u') in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
// TypeForKey will traverse the inheritance chain to get the correct value.
func (t Tag) TypeForKey(key string) string {
	// TODO: implement
	return ""
}

// SetTypeForKey returns a new Tag with the key set to type, where key and type
// are of the allowed values defined for the Unicode locale extension ('u') in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
func (t Tag) SetTypeForKey(key, value string) Tag {
	// TODO: implement
	return Tag{}
}

// Base is an ISO 639 language code, used for encoding the base language
// of a language tag.
type Base struct {
	langID
}

// ParseBase parses a 2- or 3-letter ISO 639 code.
// It returns an error if the given string is not a valid language code.
func ParseBase(s string) (Base, error) {
	if n := len(s); n < 2 || 3 < n {
		return Base{}, errInvalid
	}
	var buf [3]byte
	l, err := getLangID(buf[:copy(buf[:], s)])
	return Base{l}, err
}

// Tag returns a Tag with this base language as its only subtag.
func (b Base) Tag() Tag {
	return Tag{lang: b.langID}
}

// Script is a 4-letter ISO 15924 code for representing scripts.
// It is idiomatically represented in title case.
type Script struct {
	scriptID
}

// ParseScript parses a 4-letter ISO 15924 code.
// It returns an error if the given string is not a valid script code.
func ParseScript(s string) (Script, error) {
	if len(s) != 4 {
		return Script{}, errInvalid
	}
	var buf [4]byte
	sc, err := getScriptID(script, buf[:copy(buf[:], s)])
	return Script{sc}, err
}

// Tag returns a Tag with the undetermined language and this script as its only subtags.
func (s Script) Tag() Tag {
	return Tag{script: s.scriptID}
}

// Region is an ISO 3166-1 or UN M.49 code for representing countries and regions.
type Region struct {
	regionID
}

// EncodeM49 returns the Region for the given UN M.49 code.
// It returns an error if r is not a valid code.
func EncodeM49(r int) (Region, error) {
	rid, err := getRegionM49(r)
	return Region{rid}, err
}

// ParseRegion parses a 2- or 3-letter ISO 3166-1 or a UN M.49 code.
// It returns an error if the given string is not a valid region code.
func ParseRegion(s string) (Region, error) {
	if n := len(s); n < 2 || 3 < n {
		return Region{}, errInvalid
	}
	var buf [3]byte
	r, err := getRegionID(buf[:copy(buf[:], s)])
	return Region{r}, err
}

// Tag returns a Tag with the undetermined language and this region as its only subtags.
func (r Region) Tag() Tag {
	return Tag{region: r.regionID}
}

// IsCountry returns whether this region is a country or autonomous area.
func (r Region) IsCountry() bool {
	if r.regionID < isoRegionOffset || r.IsPrivateUse() {
		return false
	}
	return true
}

// Variant represents a registered variant of a language as defined by BCP 47.
type Variant struct {
	// TODO: implement
	variant string
}

// String returns the string representation of the variant.
func (v Variant) String() string {
	// TODO: implement
	return v.variant
}

// Currency is an ISO 4217 currency designator.
type Currency struct {
	currencyID
}

// ParseCurrency parses a 3-letter ISO 4217 code.
// It returns an error if the given string is not a valid currency code.
func ParseCurrency(s string) (Currency, error) {
	if len(s) != 3 {
		return Currency{}, errInvalid
	}
	var buf [3]byte
	c, err := getCurrencyID(currency, buf[:copy(buf[:], s)])
	return Currency{c}, err
}

// Set provides information about a set of tags.
type Set interface {
	Tags() []Tag
	BaseLanguages() []Base
	Regions() []Region
	Scripts() []Script
	Currencies() []Currency
}
