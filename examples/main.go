package main

import (
	"fmt"

	"github.com/oarkflow/dipper"
)

type Request struct {
	Em  Em
	Cpt []Cpt
}

type Em struct {
	Code             string
	EncounterUid     int
	WorkItemUid      int
	BillingProvider  string
	ResidentProvider string
}

type Cpt struct {
	Code             string
	EncounterUid     int
	WorkItemUid      int
	BillingProvider  string
	ResidentProvider string
}

func main() {
	data := Request{
		Em: Em{
			Code:             "001",
			EncounterUid:     1,
			WorkItemUid:      2,
			BillingProvider:  "Test provider",
			ResidentProvider: "Test Resident Provider",
		},
		Cpt: []Cpt{
			{
				Code:             "001",
				EncounterUid:     1,
				WorkItemUid:      2,
				BillingProvider:  "Test provider",
				ResidentProvider: "Test Resident Provider",
			},
			{
				Code:             "OBS01",
				EncounterUid:     1,
				WorkItemUid:      2,
				BillingProvider:  "Test provider",
				ResidentProvider: "Test Resident Provider",
			},
			{
				Code:             "SU002",
				EncounterUid:     1,
				WorkItemUid:      2,
				BillingProvider:  "Test provider",
				ResidentProvider: "Test Resident Provider",
			},
		},
	}
	data2 := []Cpt{
		{
			Code:             "001",
			EncounterUid:     1,
			WorkItemUid:      2,
			BillingProvider:  "Test provider",
			ResidentProvider: "Test Resident Provider",
		},
		{
			Code:             "OBS01",
			EncounterUid:     2,
			WorkItemUid:      2,
			BillingProvider:  "Test provider",
			ResidentProvider: "Test Resident Provider",
		},
		{
			Code:             "SU002",
			EncounterUid:     3,
			WorkItemUid:      2,
			BillingProvider:  "Test provider",
			ResidentProvider: "Test Resident Provider",
		},
	}

	fmt.Println(dipper.Get(data, "Cpt.#.Code", "Cpt.EncounterUid"))
	fmt.Println(dipper.Get(data2, "#.Code"))
}
