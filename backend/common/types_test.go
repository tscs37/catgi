package common

import (
	"testing"
	"time"

	"errors"

	"github.com/stretchr/testify/assert"
)

func TestNewFNEError(t *testing.T) {
	assert := assert.New(t)

	fneerr := NewErrorFileNotExists("test.file", nil)
	assert.Error(fneerr)
	assert.True(IsFileNotExists(fneerr))
	assert.EqualValues("ErrFileNotExist(test.file)", fneerr.Error())

	fneerr = NewErrorFileNotExists("test.file", errors.New("Test error"))
	assert.Error(fneerr)
	assert.True(IsFileNotExists(fneerr))
	assert.EqualValues("ErrFileNotExist(test.file): Test error", fneerr.Error())
}

func TestDateOnlyTime(t *testing.T) {
	assert := assert.New(t)

	dot, err := FromString("2017-12-12")

	assert.NoError(err, "Parsing ISO Date does not yield error")

	assert.EqualValues(12, dot.Day())
	assert.EqualValues(12, dot.Month())
	assert.EqualValues(2017, dot.Year())

	marsh, err := dot.MarshalJSON()

	assert.NoError(err, "Marshall returns no error")
	assert.Equal("\"2017-12-12\"", string(marsh), "Marshall returns ISO in JSON")
}

func TestDOTWrongFormat(t *testing.T) {
	assert := assert.New(t)

	dot, err := FromString("Not A Date")

	assert.Error(err)
	assert.Nil(dot)
}

func TestDOTTTL(t *testing.T) {
	assert := assert.New(t)

	dot := FromTime(time.Now().AddDate(0, 0, 1))
	assert.EqualValues(1*24*60*60*time.Second, dot.TTL())

	dot = FromTime(time.Now().AddDate(-1, -1, -1))
	assert.EqualValues(0, dot.TTL())
}

func TestDOTMarshal(t *testing.T) {
	assert := assert.New(t)

	dot := FromTime(time.Now())

	err := dot.UnmarshalJSON([]byte("\"\""))

	assert.Error(err)
	assert.EqualValues("Cannot parse empty date", err.Error())

	err = dot.UnmarshalJSON([]byte("\"hello\""))

	assert.Error(err)

	err = dot.UnmarshalJSON([]byte("\"2017-12-12\""))

	assert.NoError(err)
}
