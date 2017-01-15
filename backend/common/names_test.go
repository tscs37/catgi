package common

import "testing"

import "github.com/stretchr/testify/assert"

func TestEscapeName(t *testing.T) {
	assert := assert.New(t)

	assert.EqualValues("HelloWorld", EscapeName("Hello/World"))
	assert.EqualValues("HelloWorld", EscapeName("Hello\\World"))
	assert.EqualValues("HelloWorld", EscapeName("HelloWorld"))
}

func TestSplitName(t *testing.T) {
	assert := assert.New(t)

	assert.EqualValues("H/e/l/l/o/W/o/r/ld", SplitName("HelloWorld", 1))
	assert.EqualValues("He/ll/oW/or/ld", SplitName("HelloWorld", 2))
	assert.EqualValues("Hel/loW/orld", SplitName("HelloWorld", 3))
	assert.EqualValues("Hell/oWor/ld", SplitName("HelloWorld", 4))
	assert.EqualValues("Hello/World", SplitName("HelloWorld", 5))
	assert.EqualValues("HelloW/orld", SplitName("HelloWorld", 6))
	assert.EqualValues("HelloWo/rld", SplitName("HelloWorld", 7))
	assert.EqualValues("HelloWor/ld", SplitName("HelloWorld", 8))
	assert.EqualValues("HelloWorld", SplitName("HelloWorld", 9))
	assert.EqualValues("HelloWorld", SplitName("HelloWorld", 10))
}

func TestDataName(t *testing.T) {
	assert := assert.New(t)

	dn := DataName("ThisIsATest", 2)
	assert.EqualValues("file/Th/is/Is/AT/est/data.bin", dn)
	assert.True(IsDataFile(dn))
	assert.False(IsDataFile(dn[1:]))
	assert.False(IsDataFile(dn[:len(dn)-1]))
}

func TestMetaName(t *testing.T) {
	assert := assert.New(t)

	form := "json"
	dn := MetaName("ThisIsATest", 2, form)
	assert.EqualValues("file/Th/is/Is/AT/est/meta."+form, dn)
	assert.True(IsMetaFile(dn, form))
	assert.False(IsMetaFile(dn[1:], form))
	assert.False(IsMetaFile(dn[:len(dn)-1], form))
}

func TestPubName(t *testing.T) {
	assert := assert.New(t)

	dn := PubName("ThisIsATest", 2)
	assert.EqualValues("public/Th/is/Is/AT/est", dn)
	assert.True(IsPublicFile(dn))
	assert.False(IsPublicFile(dn[1:]))
}

func TestClearPubName(t *testing.T) {
	assert := assert.New(t)

	dn := ClearPubName("ThisIsATest", 2)
	assert.EqualValues("named/Th/is/Is/AT/est/flakes.json", dn)
	assert.True(IsNamedFile(dn))
	assert.False(IsNamedFile(dn[1:]))
	assert.False(IsNamedFile(dn[:len(dn)-1]))
}
