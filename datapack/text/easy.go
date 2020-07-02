package text

import "bytes"

func Marshal(v interface{}) ([]byte, error) {
	out := new(bytes.Buffer)
	err := NewEncoder(out).Encode(v)
	return out.Bytes(), err
}

func Unmarshal(b []byte, v interface{}) error {
	in := bytes.NewReader(b)
	err := NewDecoder(in).Decode(v)
	return err
}
