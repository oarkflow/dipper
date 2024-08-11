package main

import (
	"fmt"

	"github.com/oarkflow/json/sjson"

	"github.com/oarkflow/dipper"
)

func main() {
	// test()
	test2()
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
					"code": "123",
					"encounter_uid": 1,
					"work_item_uid": 2,
					"billing_provider": "Test provider",
					"resident_provider": "Test Resident Provider"
				},
				"cpt": [
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

	result = sjson.GetBytes(requestData2, "coding.#.em")
	if result.Exists() {
		em := result.Value()
		fmt.Println(em)
	}

	result = sjson.GetBytes(requestData2, "coding.#.em.code")
	if result.Exists() {
		code := result.Value()
		fmt.Println(code)
	}

	result = sjson.GetBytes(requestData2, "coding.#.cpt.#.code")
	if result.Exists() {
		cpt := result.Value()
		fmt.Println(cpt)
	}
}

func test2() {
	result, err := dipper.Get(requestData2, "coding.#.details.cpt.#.code")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
