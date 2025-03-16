package balena

// flattenJSON recursively flattens a nested JSON map.
func flattenJSON(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := flattenJSON(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}

type IDWrapper struct {
	ID int `json:"__id"`
}

type BalenaListResponse struct {
	Data []map[string]interface{} `json:"d"`
}

type BalenaSingleResponse struct {
	Data map[string]interface{} `json:"d"`
}
