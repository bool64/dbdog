package dbdog_test

import (
	"testing"
	"time"

	"github.com/bool64/dbdog"
	"github.com/stretchr/testify/assert"
	"github.com/swaggest/form/v5"
)

func TestMapper_SliceFromTable(t *testing.T) {
	type Emb struct {
		B   string         `db:"b"`
		Map map[string]int `db:"m"`
	}

	type item struct {
		A int `db:"a"`
		Emb
	}

	data := [][]string{
		{"a", "b"},
		{"1", "b1"},
		{"2", "b2"},
	}

	m := &dbdog.TableMapper{
		Decoder: form.NewDecoder(),
	}
	m.Decoder.SetTagName("db")
	res, err := m.SliceFromTable(data, new(item))
	assert.NoError(t, err)

	result, ok := res.([]item)
	assert.True(t, ok)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].A)
	assert.Equal(t, "b1", result[0].B)
	assert.Equal(t, 2, result[1].A)
	assert.Equal(t, "b2", result[1].B)
}

func TestTableMapper_Encode(t *testing.T) {
	tm := dbdog.TableMapper{}

	for _, tc := range []struct {
		v interface{}
		s string
	}{
		{"abc", "abc"},
		{123, "123"},
		{123.45, "123.45"},
		{nil, "NULL"},
		{(*time.Time)(nil), "NULL"},
		{time.Time{}, "0001-01-01T00:00:00Z"},
		{&time.Time{}, "0001-01-01T00:00:00Z"},
		{new(int), "0"},
	} {
		s, err := tm.Encode(tc.v)
		assert.NoError(t, err)
		assert.Equal(t, tc.s, s)
	}
}
