package seed

import (
	"github.com/tmazeika/testpass/repo"
	"strconv"
)

func Seed(repos repo.Repos) error {
	go func() {
		for {
			// ignore changes
			<-repos.Changes()
		}
	}()

	for i := 0; i < 30; i++ {
		suiteId, err := repos.Suites().Save(repo.Suite{
			Name: "Suite " + strconv.Itoa(i),
			FailureTypes: []repo.SuiteFailureType{
				{
					Name:        "My Error",
					Description: "A custom error",
				},
			},
			Tags: []string{"smoke"},
			EnvVars: []repo.SuiteEnvVar{
				{
					Key:   "BROWSER",
					Value: "chrome",
				},
			},
			PlannedCases: int64(i % 10),
			Status:       repo.SuiteStatusRunning,
			StartedAt:    1589947257188,
		})
		if err != nil {
			return err
		}

		for j := 0; j < 30; j++ {
			caseId, err := repos.Cases().Save(repo.Case{
				Suite:       suiteId,
				Name:        "Case " + strconv.Itoa(j),
				Description: "This is my test case.",
				Tags:        []string{"fast"},
				Num:         int64(j % 28),
				Links: []repo.CaseLink{
					{
						Type: repo.CaseLinkTypeIssue,
						Name: "ISSUE-5",
						Url:  "https://example.com/issues/ISSUE-5",
					},
				},
				Args: []repo.CaseArg{
					{Key: "x", Value: 3},
					{Key: "y", Value: "five"},
				},
				Status:    repo.CaseStatusCreated,
				CreatedAt: 1589949229585,
			})
			if err != nil {
				return err
			}

			for k := 0; k < 30; k++ {
				_, err := repos.Logs().Save(repo.LogEntry{
					Case:      caseId,
					Seq:       int64(k),
					Level:     repo.LogLevelTypeInfo,
					Trace:     "",
					Message:   "This is a log message!",
					Timestamp: 1589950443438,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
