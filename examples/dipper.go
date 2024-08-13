package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/oarkflow/dipper"
)

func main() {
	var data map[string]any
	content, err := os.ReadFile("data.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		panic(err)
	}

	fmt.Println(dipper.Get(data, "coding.#.details.cpt.#.code"))
	fmt.Println(dipper.Get(data, "coding.#.details.cpt.#.code", "coding.#.dos"))
}
