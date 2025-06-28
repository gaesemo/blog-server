package cursor

import (
	"encoding/base32"
	"strconv"

	typesv1 "github.com/gaesemo/blog-api/go/types/v1"
)

func FromInt64(i64 int64) *typesv1.Cursor {
	raw := strconv.AppendInt(nil, i64, 36)
	cursor := make([]byte, base32.HexEncoding.EncodedLen(len(raw)))
	base32.HexEncoding.Encode(cursor, raw)
	return &typesv1.Cursor{Opaque: cursor}
}

func MustParseInt64(cursor *typesv1.Cursor) int64 {
	if cursor == nil {
		return 0
	}
	if len(cursor.Opaque) == 0 {
		return 0
	}

	raw, err := base32.HexEncoding.DecodeString(string(cursor.Opaque))
	if err != nil {
		return 0
	}
	i64, err := strconv.ParseInt(string(raw), 36, 64)
	if err != nil {
		return 0
	}
	return i64
}
