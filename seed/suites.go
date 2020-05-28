package seed

import (
	"github.com/tmazeika/testpass/repo"
)

var suites = []repo.UnsavedSuite{
	{
		Name: "My Automation Test",
		FailureTypes: []repo.SuiteFailureType{
			{"IO", "Input/output exception"},
		},
		Tags: []string{"smoke", "footer"},
		EnvVars: []repo.SuiteEnvVar{
			{"BROWSER", "chrome"},
			{"DEBUG", true},
		},
		PlannedCases: 3,
		Status:       repo.SuiteStatusRunning,
	},
	{
		Name: "Volutpat Purus",
		PlannedCases: 0,
		Status:       repo.SuiteStatusRunning,
	},
	{
		Name: "Donec Eu Ex Nec",
		FailureTypes: []repo.SuiteFailureType{
			{"Est", "Mi tristique consectetur"},
			{"Tincidunt", "Suspendisse imperdiet leo leo"},
		},
		Tags: []string{"sapien", "tellus", "leo"},
		EnvVars: []repo.SuiteEnvVar{
			{"EGET", "sollicitudin/etiam"},
			{"ID", 371535},
		},
		PlannedCases: 7,
		Status:       repo.SuiteStatusPassed,
	},
	{
		Name:         "Sed Non Enim",
		PlannedCases: 2,
		Status:       repo.SuiteStatusFailed,
	},
	{
		Name: "Nisi Sed Luctus",
		Tags: []string{"vehicula"},
		EnvVars: []repo.SuiteEnvVar{
			{"SAGITTIS", true},
			{"MAURIS", nil},
			{"TELLUS", []string{"tellus", "cras", "euismod"}},
		},
		PlannedCases: 8,
		Status:       repo.SuiteStatusRunning,
	},
	{
		Name: "Mattis Vitae Non",
		FailureTypes: []repo.SuiteFailureType{
			{"Facilisis", "In hac habitasse platea"},
		},
		EnvVars: []repo.SuiteEnvVar{
			{"FELIS", map[string]interface{}{
				"laoreet": "nulla",
				"ornare":  false},
			},
		},
		PlannedCases: 14,
		Status:       repo.SuiteStatusDisconnected,
	},
}
