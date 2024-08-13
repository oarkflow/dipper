package main

import (
	"encoding/json"
	"fmt"

	"github.com/oarkflow/json/sjson"

	"github.com/oarkflow/dipper"
)

func main() {
	test()
	// test2()
}

var requestData2 = []byte(`
{
	"patient": {
		"dob": "2003-04-10"
	},
	"coding": [
		{
			"dos": "2021-01-01",
			"details": {
				"em": {
					"code": "123"
				},
				"cpt": [
					{
						"code": "OBS011"
					},
					{
						"code": "OBS011"
					},
					{
						"code": "SU002"
					}
				]
			}
		},
		{
			"dos": "2021-01-02",
			"details": {
				"em": {
					"code": "123"
				},
				"cpt": [
					{
						"code": "1OBS011"
					},
					{
						"code": "1OBS011"
					},
					{
						"code": "1SU002"
					}
				]
			}
		}
	]
}
`)

var reqData = `
[
          {
              "code": "OBS011",
              "billing_provider": "Test provider",
              "resident_provider": "Test Resident Provider"
          },
          {
              "code": "OBS011",
              "billing_provider": "Test provider",
              "resident_provider": "Test Resident Provider"
          },
          {
              "code": "SU002",
              "billing_provider": "Test provider",
              "resident_provider": "Test Resident Provider"
          }
      ]
    }
  ]
`

func test() {
	result := sjson.GetBytes(requestData2, "patient.dob")
	if result.Exists() {
		dob := result.Value()
		fmt.Println(dob.(string))
	}

	result = sjson.GetBytes(requestData2, "coding.#.details.cpt.#.code")
	if result.Exists() {
		em := result.Value()
		fmt.Println(em)
	}

	result = sjson.GetBytes(requestData2, "coding.#.dos")
	if result.Exists() {
		em := result.Value()
		fmt.Println(em)
	}
}

func test2() {
	var data map[string]any
	err := json.Unmarshal(requestData2, &data)
	if err != nil {
		panic(err)
	}

	result, err := dipper.Get(data, "coding.#.details.cpt.#.code")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
