package seed

import (
	"github.com/suiteserve/suiteserve/repo"
)

var cases = []repo.UnsavedCase{
	{
		Name:        "My Cool Case",
		Description: "This test case does amazing things.",
		Tags:        []string{"all", "header"},
		Links: []repo.CaseLink{
			{
				Type: repo.CaseLinkTypeIssue,
				Name: "XYZ-5",
				Url:  "https://example.com/issues/zyx-5",
			},
			{
				Type: repo.CaseLinkTypeOther,
				Name: "example.org",
				Url:  "https://example.org",
			},
		},
		Args: []repo.CaseArg{
			{"x", 3},
			{"y", "9"},
			{"bool", true},
		},
		Status: repo.CaseStatusCreated,
	},
	{
		Name:        "Cras A Lorem",
		Description: "Duis facilisis ex et purus viverra.",
		Links: []repo.CaseLink{
			{
				Type: repo.CaseLinkTypeIssue,
				Name: "PURUS-4998",
				Url:  "https://issues.example.com/purus-4998",
			},
		},
		Args: []repo.CaseArg{
			{"arr", []string{"x", "y", "z"}},
		},
		Status: repo.CaseStatusRunning,
	},
	{
		Name:        "In Hac Habitasse",
		Description: "Sed commodo elit ex in ex.",
		Tags:        []string{"dictumst"},
		Status:      repo.CaseStatusDisabled,
	},
	{
		Name:        "Libero Sit Amet",
		Description: "Non ultricies elit magna!",
		Tags:        []string{"semper", "lacinia"},
		Args: []repo.CaseArg{
			{"hello", "world"},
			{"ultrices", map[string]interface{}{
				"urna":      "posuere massa",
				"tristique": 630086,
			}},
			{"viverra", []interface{}{"rutrum", 627721297, true, false, nil}},
		},
		Status: repo.CaseStatusPassed,
	},
	{
		Name:        "Lorem Ipsum Dolor",
		Description: "Phasellus eu sapien et justo...",
		Links: []repo.CaseLink{
			{
				Type: repo.CaseLinkTypeOther,
				Name: "Sed pretium a enim",
				Url:  "https://example.com/sed-pretium-a-enim",
			},
		},
		Status: repo.CaseStatusFailed,
	},
	{
		Name:        "Dui Nunc Imperdiet",
		Description: "Cras a lorem nec erat.",
		Status:      repo.CaseStatusErrored,
	},
	{
		Name:        "Felis A Auctor",
		Description: "In hac habitasse platea!",
		Status:      repo.CaseStatusRunning,
	},
}
