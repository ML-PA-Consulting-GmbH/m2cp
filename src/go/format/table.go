package format

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Table struct {
	Fields map[string]string
	Rows   []map[string]string
}

func NewTable(fields map[string]string) *Table {
	return &Table{
		Fields: fields,
		Rows:   make([]map[string]string, 0),
	}
}

func (t *Table) AddRow(row map[string]string) {
	t.Rows = append(t.Rows, row)
}

func (t *Table) String() string {
	fields := make([]string, 0, len(t.Fields))
	for field := range t.Fields {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return t.StringSelect(fields)
}

func (t *Table) Filter(field, pattern string) *Table {
	if _, ok := t.Fields[field]; !ok {
		return t // invalid field
	}

	// compile the pattern
	r, err := regexp.Compile(strings.Replace(pattern, `*`, `.*`, -1))
	if err != nil {
		return t // invalid pattern
	}

	// filter
	filtered := NewTable(t.Fields)
	for _, row := range t.Rows {
		if r.MatchString(row[field]) {
			filtered.AddRow(row)
		}
	}
	return filtered
}

func (t *Table) Sort(sort string) *Table {
	return t.Sorts([]string{sort})
}

// Sorts sorts the table by the given fields.
// Sorting order can be specified by adding ":asc", ":nasc", ":desc" or ":ndesc" to the field name.
// Orders starting with "n" are numeric.
// For example, "name:desc" sorts by name in descending order.
func (t *Table) Sorts(sorts []string) *Table {
	// sort expects field names, not field captions.. let's translate them, if necessary
	for i, sort := range sorts {
		parts := strings.Split(sort, ":")
		fieldName := parts[0]
		if _, ok := t.Fields[fieldName]; ok {
			continue
		}
		// try to find the field by caption
		for field, caption := range t.Fields {
			if caption == fieldName {
				if len(parts) > 1 {
					sorts[i] = field + ":" + parts[1]
				} else {
					sorts[i] = field
				}
				break
			}
		}
	}
	sortedRows := make([]map[string]string, len(t.Rows))
	copy(sortedRows, t.Rows)
	sort.SliceStable(sortedRows, func(i, j int) bool {
		for _, sort := range sorts {
			parts := strings.Split(sort, ":")
			field := parts[0]
			direction := "asc"
			if len(parts) > 1 {
				direction = parts[1]
			}
			if direction == "desc" || direction == "ndesc" {
				i, j = j, i
			}
			if direction[0:1] == "n" {
				iVal, _ := strconv.ParseFloat(sortedRows[i][field], 64)
				jVal, _ := strconv.ParseFloat(sortedRows[j][field], 64)
				if iVal < jVal {
					return true
				} else if iVal > jVal {
					return false
				}
			} else {
				if sortedRows[i][field] < sortedRows[j][field] {
					return true
				} else if sortedRows[i][field] > sortedRows[j][field] {
					return false
				}
			}
		}
		return false
	})
	return &Table{
		Fields: t.Fields,
		Rows:   sortedRows,
	}
}

func (t *Table) StringSelect(fields []string) string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 4, ' ', 0)
	header := ""
	for _, v := range fields {
		header += fmt.Sprintf("%s\t", strings.Trim(t.Fields[v], "\n\t "))
	}
	_, _ = fmt.Fprintln(w, header)
	if len(t.Rows) == 0 {
		_ = w.Flush()
		return fmt.Sprintf("%s[empty]\n", b.String())
	}
	for _, row := range t.Rows {
		for _, field := range fields {
			_, _ = fmt.Fprintf(w, "%s\t", strings.Trim(row[field], "\n\t "))
		}
		_, _ = fmt.Fprintln(w, "")
	}
	_ = w.Flush()
	return b.String()
}
