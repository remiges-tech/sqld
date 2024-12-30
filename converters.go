package sqld

import (
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5/pgtype"
)

func init() {
	// Register bool converter
	RegisterConverter(reflect.TypeOf((*pgtype.Bool)(nil)).Elem(), func(v interface{}) (interface{}, error) {
		switch val := v.(type) {
		case bool:
			return &pgtype.Bool{Bool: val, Valid: true}, nil
		case *bool:
			if val == nil {
				return &pgtype.Bool{Valid: false}, nil
			}
			return &pgtype.Bool{Bool: *val, Valid: true}, nil
		default:
			return nil, fmt.Errorf("value %v is not a bool", v)
		}
	})
}
