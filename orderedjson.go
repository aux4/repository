package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// kv is an ordered key/value pair whose value is preserved as raw JSON bytes.
// Keeping values as json.RawMessage preserves both key insertion order and
// number-literal formatting (e.g. 45.50, 199.00) byte-for-byte, matching the
// jq-based pipeline this rewrite replaces.
type kv struct {
	Key string
	Val json.RawMessage
}

// parseObject decodes a JSON object into ordered key/value pairs, preserving the
// raw bytes of each value.
func parseObject(b []byte) ([]kv, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected a JSON object")
	}

	var out []kv
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string key")
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		out = append(out, kv{Key: key, Val: raw})
	}

	if _, err := dec.Token(); err != nil { // consume closing '}'
		return nil, err
	}
	return out, nil
}

// setKV updates the value of an existing key in place (preserving its position)
// or appends it when absent, matching jq's `. + {key: value}` semantics.
func setKV(kvs []kv, key string, val json.RawMessage) []kv {
	for i := range kvs {
		if kvs[i].Key == key {
			kvs[i].Val = val
			return kvs
		}
	}
	return append(kvs, kv{Key: key, Val: val})
}

// getKV returns the raw value for key and whether it was present.
func getKV(kvs []kv, key string) (json.RawMessage, bool) {
	for _, e := range kvs {
		if e.Key == key {
			return e.Val, true
		}
	}
	return nil, false
}

// removeKV returns a copy of kvs with the given key removed.
func removeKV(kvs []kv, key string) []kv {
	res := make([]kv, 0, len(kvs))
	for _, e := range kvs {
		if e.Key != key {
			res = append(res, e)
		}
	}
	return res
}

// emitObject serializes ordered pairs into compact JSON, preserving each value's
// raw byte representation.
func emitObject(kvs []kv) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, e := range kvs {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(e.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		var compact bytes.Buffer
		if err := json.Compact(&compact, e.Val); err != nil {
			return nil, err
		}
		buf.Write(compact.Bytes())
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// compactArray joins compact object encodings into a single-line JSON array
// (like `jq -c`): no indentation, no newlines between elements. Each object is
// already byte-preserving (number literals and key order intact), so this simply
// wraps them in brackets without re-serializing.
func compactArray(objs [][]byte) ([]byte, error) {
	var out bytes.Buffer
	out.WriteByte('[')
	for i, o := range objs {
		if i > 0 {
			out.WriteByte(',')
		}
		out.Write(o)
	}
	out.WriteByte(']')
	return out.Bytes(), nil
}
