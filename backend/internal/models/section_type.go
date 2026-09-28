package models

// SectionType identifies the kind of a landing-page section. It mirrors the
// CHECK constraint on sections.type.
type SectionType string

const (
	SectionTypeHero    SectionType = "hero"
	SectionTypeAbout   SectionType = "about"
	SectionTypeGallery SectionType = "gallery"
	SectionTypeContact SectionType = "contact"
	SectionTypeHours   SectionType = "hours"
)

// IsValidSectionType reports whether value is a known section type.
func IsValidSectionType(value string) bool {
	switch SectionType(value) {
	case SectionTypeHero, SectionTypeAbout, SectionTypeGallery, SectionTypeContact, SectionTypeHours:
		return true
	default:
		return false
	}
}
