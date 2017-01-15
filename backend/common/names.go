package common

import "strings"

// EscapeName removes all slashes from a string without
// replacement.
// This makes names like "Hello/World" equivalent to "HelloWorld"
// The intention is to prevent manipulated flakes to escape the root
// fs and inspect the rest of the backend.
func EscapeName(flake string) string {
	return strings.NewReplacer("/", "", "\\", "").Replace(flake)
}

// DataName returns the path used for raw data storage.
// Format: "file/<flake>/public.json"
func DataName(flake string, skipSize int) string {
	return "file/" + SplitName(flake, skipSize) + "/data.bin"
}

// MetaName returns the path that is used to store metainformation for a file
// Format: "file/<flake>/meta.json"
func MetaName(flake string, skipSize int, format string) string {
	return "file/" + SplitName(flake, skipSize) + "/meta." + format
}

// PubName returns the path to a public flake of that name
// skipSize indicates when to split apart the flake name. See: SplitName
// Format: "public/<flake>"
func PubName(flake string, skipSize int) string {
	return "public/" + SplitName(flake, skipSize)
}

// ClearPubName is used to generate the name for a clearname publication
// skipSize indicates when to split apart the flake name. See: SplitName
// Format: "named/<name>/flakes.json"
func ClearPubName(name string, skipSize int) string {
	return "named/" + SplitName(name, skipSize) + "/flakes.json"
}

// IsMetaFile returns true if the filename matches that of a Meta File
func IsMetaFile(file, format string) bool {
	return strings.HasPrefix(file, "file/") && strings.HasSuffix(file, "/meta."+format)
}

// IsDataFile returns true if the filename matches that of a Data File
func IsDataFile(file string) bool {
	return strings.HasPrefix(file, "file/") && strings.HasSuffix(file, "/data.bin")
}

// IsPublicFile returns true if the name matches that of a publication
func IsPublicFile(file string) bool {
	return strings.HasPrefix(file, "public/")
}

// IsNamedFile returns true if the name matches that of a named publication
func IsNamedFile(file string) bool {
	return strings.HasPrefix(file, "named/") && strings.HasSuffix(file, "/flakes.json")
}

// SplitName splits a string according to the following rules:
// 1. Create a slice of strings
// 2. If the remaining size of the string is larger than skipSize plus 1
//      then take the first 2 runes and append them as string to the slice
// 3. If this is not the case, take all remaining runes and append
//      them to the slice
// 4. Join all slice elements with "/" inbetween.
//
// skipSize is by default 2
//
// Example:
//
// HelloWorld       =>      He/ll/oW/or/ld
// HelloInternet    =>      He/ll/oI/nt/er/net
func SplitName(flakeStr string, skipSize int) string {
	flake := []rune(flakeStr)
	var out = []string{}
	for {
		if len(flake) > skipSize+1 {
			out = append(out, string(flake[0:skipSize]))
			flake = flake[skipSize:]
		} else {
			out = append(out, string(flake))
			return strings.Join(out, "/")
		}
	}
}
