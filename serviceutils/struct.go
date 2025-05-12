package serviceutils

import "encoding/json"

// CopyStructValue copy identical struct value from source to dest
// will return error if source and dest is not identical
func CopyStructValue(source, dest interface{}) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, dest); err != nil {
		return err
	}

	return nil
}
