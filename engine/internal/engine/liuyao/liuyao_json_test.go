package liuyao

import (
	"encoding/json"
	"fmt"
)

// These JSON projections are used by domain golden tests. The MCP wire layer
// maps the domain enums explicitly, so they intentionally do not participate
// in the production handler path.
func (q LiuQin) MarshalJSON() ([]byte, error) { return json.Marshal(q.String()) }

func (q *LiuQin) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("liu_qin: expected string, got %s", string(data))
	}
	for i, n := range liuQinNames {
		if n == s {
			*q = LiuQin(i)
			return nil
		}
	}
	return fmt.Errorf("liu_qin: unknown %q", s)
}

func (l LiuShou) MarshalJSON() ([]byte, error) { return json.Marshal(l.String()) }

func (l *LiuShou) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("liu shou must be a string, got %s", string(data))
	}
	for i, n := range liuShouNames {
		if n == name {
			*l = LiuShou(i)
			return nil
		}
	}
	return fmt.Errorf("unknown liushou: %q", name)
}
