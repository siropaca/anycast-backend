package optional

import "encoding/json"

// Field は JSON の「未送信」「null」「値あり」を区別できる型
//
// IsSet == false: フィールドが JSON に含まれていない（変更なし）
// IsSet == true, Value == nil: 明示的に null が送信された（削除）
// IsSet == true, Value != nil: 値が送信された（更新）
type Field[T any] struct {
	Value *T
	IsSet bool
}

// MarshalJSON は JSON エンコード時に呼ばれ、値または null を出力する
func (f Field[T]) MarshalJSON() ([]byte, error) {
	if f.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*f.Value)
}

// UnmarshalJSON は JSON デコード時に呼ばれ、フィールドの存在を記録する
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.IsSet = true
	if string(data) == "null" {
		f.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	f.Value = &v
	return nil
}
