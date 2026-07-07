package querybuilder

import (
	"fmt"
	"strings"
)

type UpdateBuilder struct {
	set  []string
	args []interface{}
	i    int
}

func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{i: 1}
}

func (b *UpdateBuilder) AddField(column string, value interface{}) {
	b.set = append(b.set, fmt.Sprintf("%s = $%d", column, b.i))
	b.args = append(b.args, value)
	b.i++
}

func (b *UpdateBuilder) AddConditionalField(column string, value interface{}, condition bool) {
	if condition {
		b.AddField(column, value)
	}
}

func (b *UpdateBuilder) IsEmpty() bool {
	return len(b.set) == 0
}

func (b *UpdateBuilder) Build(table string, idColumn string, idValue interface{}) (string, []interface{}) {
	query := fmt.Sprintf("UPDATE %s SET %s, updated_at = NOW() WHERE %s = $%d",
		table, strings.Join(b.set, ", "), idColumn, b.i)
	b.args = append(b.args, idValue)
	return query, b.args
}
