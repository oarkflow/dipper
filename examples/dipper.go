package main

import (
	"fmt"
	"github.com/oarkflow/dipper"
)

func main() {
	data := map[string]any{
		"patient": map[string]any{
			"dob": "2003-04-10",
		},
		"coding": []any{
			map[string]any{
				"dos": "2021-01-01",
				"details": map[string]any{
					"em": map[string]any{
						"code": "123",
					},
					"cpt": []any{
						map[string]any{"code": "OBS011"},
						map[string]any{"code": "OBS011"},
						map[string]any{"code": "SU002"},
					},
				},
			},
			map[string]any{
				"dos": "2021-01-02",
				"details": map[string]any{
					"em": map[string]any{
						"code": "123",
					},
					"cpt": []any{
						map[string]any{"code": "1OBS011"},
						map[string]any{"code": "1OBS011"},
						map[string]any{"code": "1SU002"},
					},
				},
			},
		},
	}

	// Use the custom Extract function
	fmt.Println(dipper.Get(data, "coding.#.details.cpt.#.code"))
}
