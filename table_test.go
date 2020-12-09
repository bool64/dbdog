package dbdog_test

import (
	"testing"

	"github.com/bool64/dbdog"
	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"
	"github.com/swaggest/form"
)

type (
	tableRow  = messages.PickleStepArgument_PickleTable_PickleTableRow
	tableCell = messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell
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

	data := &godog.Table{
		Rows: []*tableRow{
			{Cells: []*tableCell{
				{Value: "a"},
				{Value: "b"},
			}},
			{Cells: []*tableCell{
				{Value: "1"},
				{Value: "b1"},
			}},
			{Cells: []*tableCell{
				{Value: "2"},
				{Value: "b2"},
			}},
		},
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
