package dbdog

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	"github.com/swaggest/form/v5"
)

const null = "NULL"

// TableMapper maps data from Go value to string and back.
type TableMapper struct {
	Decoder *form.Decoder
	Encoder *form.Encoder
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && rv.IsZero() {
		return true
	}

	return false
}

// Encode converts Go value to string.
func (m *TableMapper) Encode(v interface{}) (string, error) {
	if m.Encoder == nil {
		m.Encoder = form.NewEncoder()
	}

	if isNil(v) {
		return null, nil
	}

	vv, err := m.Encoder.Encode(v)
	if err != nil {
		return "", fmt.Errorf("failed to stringify variable value of type %T: %w", v, err)
	}

	return vv[""][0], nil
}

// SliceFromTable creates a slice from gherkin table, item type is used as slice element type.
func (m *TableMapper) SliceFromTable(data *godog.Table, item interface{}) (interface{}, error) {
	itemType := reflect.TypeOf(item)
	if itemType == nil {
		return nil, errNilItemStruct
	}

	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	result := reflect.MakeSlice(reflect.SliceOf(itemType), len(data.Rows)-1, len(data.Rows)-1)

	err := m.IterateTable(IterateConfig{
		Data: data, Item: item,
		ReceiveRow: func(index int, row interface{}, colNames []string, rawValues []string) error {
			result.Index(index).Set(reflect.Indirect(reflect.ValueOf(row)))

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return result.Interface(), nil
}

// ColNames returns a slice of column names.
func ColNames(cells []*messages.PickleStepArgument_PickleTable_PickleTableRow_PickleTableCell) []string {
	colNames := make([]string, len(cells))

	for i, col := range cells {
		colNames[i] = col.Value
	}

	return colNames
}

// IterateConfig controls behavior of TableMapper.IterateTable.
type IterateConfig struct {
	Data       *godog.Table
	SkipDecode func(column, value string) bool
	Item       interface{}
	Replaces   map[string]string
	ReceiveRow func(index int, row interface{}, colNames []string, rawValues []string) error
}

var (
	errNilItemStruct = errors.New("nil item struct received")
	errRowRequired   = errors.New("header and at least one row required in table")
)

func itemType(v interface{}) (reflect.Type, error) {
	itemType := reflect.TypeOf(v)
	if itemType == nil {
		return nil, errNilItemStruct
	}

	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	return itemType, nil
}

// IterateTable walks gherkin table calling row receiver with mapped row.
// If receiver returns error iteration stops and error is propagated.
func (m *TableMapper) IterateTable(c IterateConfig) error {
	if m.Decoder == nil {
		m.Decoder = form.NewDecoder()
	}

	if len(c.Data.Rows) < 2 {
		return errRowRequired
	}

	colNames := ColNames(c.Data.Rows[0].Cells)

	itemType, err := itemType(c.Item)
	if err != nil {
		return err
	}

	values := make(map[string][]string, len(colNames))

	for rowIndex, row := range c.Data.Rows[1:] {
		itemBuf := reflect.New(itemType)
		raw := make([]string, 0, len(colNames))

		for i, cell := range row.Cells {
			raw = append(raw, cell.Value)

			if c.SkipDecode != nil && c.SkipDecode(colNames[i], cell.Value) {
				continue
			}

			if strings.HasSuffix(cell.Value, "::string") {
				cell.Value = strings.TrimSuffix(cell.Value, "::string")
			}

			if v, found := c.Replaces[cell.Value]; found {
				cell.Value = v
			}

			if cell.Value != null {
				values[colNames[i]] = []string{cell.Value}
			} else {
				delete(values, colNames[i])
			}
		}

		val := itemBuf.Interface()

		err := m.Decoder.Decode(val, values)
		if err != nil {
			return err
		}

		err = c.ReceiveRow(rowIndex, itemBuf.Interface(), colNames, raw)
		if err != nil {
			return err
		}
	}

	return nil
}
