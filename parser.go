package toml

import (
	"fmt"
	"os"
	"strings"

	"github.com/kamalshkeir/kstrct"
)

func ParseFileAndFill[T any](filePath string, ptrStrct *T) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	m, err := Parse(data)
	if err != nil {
		return fmt.Errorf("error parsing toml file: %v", err)
	}
	return Fill(ptrStrct, m)
}

func ParseAndFill[T any](fileBytes []byte, ptrStrct *T) error {
	m, err := Parse(fileBytes)
	if err != nil {
		return fmt.Errorf("error parsing toml file: %v", err)
	}
	return Fill(ptrStrct, m)
}

func Parse(fileBytes []byte) (map[string]any, error) {
	config := make(map[string]any)
	lines := strings.Split(string(fileBytes), "\n")
	state := ""
	stateKey := ""
	stateContent := ""
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "#") || line == "" {
			if state != "multi" {
				state = ""
				stateKey = ""
				stateContent = ""
			}
			continue
		}
		switch state {
		case "multi":
			if !strings.Contains(line, `"""`) {
				stateContent += `\n` + line
			} else {
				stateContent += `\n` + strings.TrimSuffix(line, `"""`)
				config[stateKey] = stateContent
				state = ""
				stateKey = ""
				stateContent = ""
			}
			continue
		case "table":
			if strings.Contains(line, "=") {
				sp := strings.Split(line, "=")
				key := sp[0]
				value := sp[1]
				if sl, ok := config[stateKey].(map[string]any); ok {
					sl[key] = value
				} else {
					fmt.Println(stateKey, "is not map")
				}
				continue
			}
		case "nested":
			if strings.Contains(line, "=") {
				sp := strings.Split(line, "=")
				key := sp[0]
				value := sp[1]
				if sl, ok := config[stateKey].(map[string]any); ok {
					if nn, ok := sl[stateContent].(map[string]any); ok {
						nn[key] = value
					} else {
						sl[stateContent].(map[string]any)[key] = map[string]any{
							key: value,
						}
					}
				} else {
					fmt.Println(stateKey, "is not map")
				}
				continue
			}
		case "slice":
			if strings.Contains(line, "=") {
				sp := strings.Split(line, "=")
				key := sp[0]
				value := sp[1]
				if sl, ok := config[stateKey].([]map[string]any); ok {
					if len(sl) == 0 {
						config[stateKey] = append(sl, map[string]any{
							key: value,
						})
					} else {
						m := sl[len(sl)-1]
						m[key] = value
					}
				} else {
					fmt.Println(stateKey, "is not slice map")
				}
				continue
			}
		}

		if strings.Contains(line, "=") {
			// kv values or multi lines
			sp := strings.Split(line, "=")
			key := strings.TrimSpace(sp[0])
			value := strings.TrimSpace(sp[1])

			// check multi
			if strings.Contains(value, `"""`) {
				state = "multi"
				stateKey = key
				stateContent = strings.ReplaceAll(value, `"""`, "")
				continue
			}

			if strings.Contains(value, "[") {
				value = strings.Trim(value, "[]")
			} else {
				value = strings.Trim(value, "\"")
			}
			config[key] = value
		} else {
			// table or slice
			if line[0] == '[' {
				if line[1] == '[' {
					// slice
					key := strings.TrimSpace(strings.Trim(line, "[]"))
					state = "slice"
					stateKey = key
					if vvv, ok := config[key].([]map[string]any); !ok {
						config[key] = make([]map[string]any, 0)
					} else {
						config[key] = append(vvv, map[string]any{})
					}
				} else {
					// table
					key := strings.TrimSpace(strings.Trim(line, "[]"))
					sp := strings.Split(key, ".") //database.monitor
					if len(sp) > 1 {
						// nested struct
						state = "nested"
						stateKey = sp[0]
						stateContent = sp[1]
						if v, ok := config[stateKey].(map[string]any); ok {
							if _, ok := v[stateContent]; !ok {
								v[stateContent] = map[string]any{}
								config[stateKey] = v
							}
						}
						continue
					}
					config[key] = map[string]any{}
					state = "table"
					stateKey = key
				}
				continue
			}
			fmt.Println("line not handled:", line)
		}
	}

	return config, nil
}

func Fill[T any](ptrStrct *T, mapData map[string]any) error {
	err := kstrct.FillM(ptrStrct, mapData, true)
	if err != nil {
		return err
	}
	return nil
}
