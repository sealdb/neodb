/*
 * NeoDB
 *
 * Copyright 2019 The Radon Authors.
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package shift

import (
	"fmt"

	"github.com/sealdb/neodb/tools/shift/xbase"

	"github.com/sealdb/go-mysql/canal"
	"github.com/sealdb/go-mysql/schema"
)

func (h *EventHandler) ParseValue(e *canal.RowsEvent, idx int, v interface{}) string {
	if v == nil {
		return fmt.Sprintf("NULL")
	}

	if _, ok := v.([]byte); ok {
		return fmt.Sprintf("%q", v)
	} else {
		switch {
		case e.Table.Columns[idx].Type == schema.TYPE_NUMBER:
			return fmt.Sprintf("%d", v)
		case e.Table.Columns[idx].Type == schema.TYPE_BIT:
			// Here we should add prefix "0x" for hex
			return fmt.Sprintf("0x%x", v)
		default:
			switch e.Table.Columns[idx].RawType {
			case "tinyblob", "blob", "mediumblob", "longblob":
				// Here we should add prefix "0x" for hex
				str := fmt.Sprintf("0x%x", v)
				// If str is empty, we`ll got "0x"
				if str == "0x" {
					return "\"\""
				}
				return str
			default:
				s := fmt.Sprintf("%v", v)
				return fmt.Sprintf("\"%s\"", xbase.EscapeBytes(xbase.StringToBytes(s)))
			}
		}
	}
}
