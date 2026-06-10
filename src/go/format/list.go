package format

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

type List struct {
	Items []ListItem
}

type ListItem struct {
	Name  string
	Value string
}

func NewList() *List {
	return &List{
		Items: make([]ListItem, 0),
	}
}

func (l *List) Add(name, value string) {
	l.Items = append(l.Items, ListItem{
		Name:  name,
		Value: value,
	})
}

func (l *List) String() string {
	var b bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.TabIndent)
	for _, item := range l.Items {
		_, _ = fmt.Fprintf(w, "%s:\t%s\n", item.Name, item.Value)
	}
	_ = w.Flush()
	return b.String()
}
