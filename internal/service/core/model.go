package core

import (
	"sort"
	"strings"
)

const (
	MaxMsgFieldValueSize = 500
)

type Message struct {
	Tag string

	Level   string
	Ts      string
	Message string
	Error   string
	Caller  string

	V map[string]any
}

func NewMessage(tag string, src map[string]any) *Message {
	m := &Message{}

	srcCopy := make(map[string]any, len(src))
	for k, v := range src {
		k = m.formatKey(k)
		if len(k) == 0 {
			continue
		}
		srcCopy[k] = v
	}
	m.V = srcCopy

	m.normalize()

	if tag != "" {
		m.Tag = tag
	}

	return m
}

func (m *Message) normalize() {
	m.Tag = m.popFieldStr("tag")
	m.Level = m.popFieldStr("level", "lvl")
	m.Ts = m.popFieldStr("ts", "time", "timestamp")
	m.Message = m.popFieldStr("msg", "message")
	m.Error = m.popFieldStr("error", "err")
	m.Caller = m.popFieldStr("caller")

	filteredV := make(map[string]any, len(m.V))
	for k, v := range m.V {
		switch val := v.(type) {
		case string:
			valRune := []rune(val)
			if len(valRune) == 0 {
				continue
			}
			if len(valRune) > MaxMsgFieldValueSize {
				filteredV[k] = string(valRune[:MaxMsgFieldValueSize]) + "..."
			} else {
				filteredV[k] = v
			}
		default:
			filteredV[k] = v
		}
	}
	m.V = filteredV
}

func (m *Message) formatKey(k string) string {
	return strings.ToUpper(strings.TrimSpace(k))
}

func (m *Message) popFieldStr(keyVariants ...string) string {
	var result string
	for _, key := range keyVariants {
		key = m.formatKey(key)
		if v, ok := m.V[key]; ok {
			result = v.(string)
			delete(m.V, key)
			if result != "" {
				return result
			}
		}
	}
	return ""
}

func (m *Message) GetVSortedKeys() []string {
	keys := make([]string, 0, len(m.V))
	for k := range m.V {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
