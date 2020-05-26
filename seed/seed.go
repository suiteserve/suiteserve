package seed

import (
	"context"
	"github.com/tmazeika/testpass/repo"
	"math/rand"
	"strconv"
	"strings"
)

func Seed(repos repo.Repos) error {
	rnd := rand.New(rand.NewSource(5))

	go func() {
		for {
			if _, ok := <-repos.Changes(); !ok {
				break
			}
			// ignore changes
		}
	}()

	for i := 0; i < 50; i++ {
		_, err := repos.Attachments().Save(context.Background(), repo.UnsavedAttachmentInfo{
			Filename: "attachment" + strconv.Itoa(i),
			ContentType: randStr(rnd, []string{
				"application/json",
				"image/jpeg",
				"image/png",
				"text/plain; charset=utf-8",
			}),
		}, strings.NewReader(randStr(rnd, []string{
			"Hello, world! I'm some plain text.\nI'm on another line...",
			`This is a plain text file.`,
			`{"x": 3, "y": 5}`,
			`I'm a 🐸 lol`,
		})))
		if err != nil {
			return err
		}
	}

	for i := 0; i < 35; i++ {
		seedTags := []string{
			"smoke",
			"integration",
			"all",
			"fast",
		}
		suiteId, err := repos.Suites().Save(context.Background(), repo.UnsavedSuite{
			Name: "A Suite " + strconv.Itoa(i),
			FailureTypes: []repo.SuiteFailureType{
				{
					Name:        "My Error",
					Description: "A custom error.",
				},
				{
					Name:        "Someone Else's Error",
					Description: "Not my error.\nI promise!",
				},
			},
			Tags: randSlice(rnd, seedTags),
			EnvVars: []repo.SuiteEnvVar{
				{
					Key:   "X",
					Value: rnd.Intn(1000000),
				},
				{
					Key:   "BROWSER",
					Value: "chrome",
				},
				{
					Key:   "A_FILE",
					Value: "/some/path/to/file",
				},
			},
			PlannedCases: rnd.Int63n(30),
			Status:       randSuiteStatus(rnd),
			StartedAt:    1589947257188 + int64(i)*2,
		})
		if err != nil {
			return err
		}

		caseCount := int(rnd.NormFloat64()*17 + 26)
		for j := 0; j < caseCount; j++ {
			seedDescriptions := []string{
				"This is my test case.",
				"There are many like it.",
				"But this one is mine!",
			}
			caseId, err := repos.Cases().Save(context.Background(), repo.UnsavedCase{
				Suite:       suiteId,
				Name:        "A Case " + strconv.Itoa(j),
				Description: randStr(rnd, seedDescriptions),
				Tags:        randSlice(rnd, seedTags),
				Num:         int64(j % 20),
				Links: []repo.CaseLink{
					{
						Type: repo.CaseLinkTypeIssue,
						Name: "ISSUE-5",
						Url:  "https://example.com/issues/ISSUE-5",
					}, {
						Type: repo.CaseLinkTypeOther,
						Name: "My Resource",
						Url:  "https://example.com/some/resource",
					},
				},
				Args: []repo.CaseArg{
					{
						Key:   "i",
						Value: rnd.Int63n(1000000),
					},
					{
						Key: "arr",
						Value: []interface{}{
							"hello",
							"world",
							50,
							true,
						},
					},
					{
						Key:   "bool",
						Value: rnd.Intn(2) == 1,
					},
					{
						Key: "obj",
						Value: map[string]interface{}{
							"x": rnd.Intn(1000000),
							"y": rnd.Intn(2) == 1,
							"nested": map[string]interface{}{
								"z": "I'm a string",
							},
						},
					},
					{
						Key: "nada",
					},
					{
						Key: "str",
						Value: randStr(rnd, []string{
							"one",
							"two",
							"three",
							"four",
							"five",
							"I'm a\nmultiline\nstr",
						}),
					},
				},
				Status:    randCaseStatus(rnd),
				CreatedAt: 1589949229585 + int64(j),
			})
			if err != nil {
				return err
			}

			for k := 0; k < rnd.Intn(50); k++ {
				_, err := repos.Logs().Save(context.Background(), repo.UnsavedLogEntry{
					Case:  caseId,
					Index: int64(k) % 40,
					Level: randLogLevelType(rnd),
					Trace: randStr(rnd, []string{
						"",
						"",
						"",
						"",
						"",
						"NullPointerException: something went wrong :(",
						"Stack trace:\nline 1: a loc\nline 2: another loc\ndone",
					}),
					Message: randStr(rnd, []string{
						"This is a log message!",
						"This is a log message!",
						"This is a log message!",
						"I'm a log log",
						"I'm a log log",
						"One entry, two entry, three...",
						"One entry, two entry, three...",
						"This is a sample message",
						"This is a sample message",
						"",
						"I have a\ttab",
						"I have a\nNL in me",
						"I have a\nfew\nnewlines\nin me!",
					}),
					Timestamp: 1589950443438 + int64(k),
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func randStr(rnd *rand.Rand, arr []string) string {
	return arr[rnd.Intn(len(arr))]
}

func randSlice(rnd *rand.Rand, arr []string) []string {
	rnd.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr[:rnd.Intn(len(arr)+1)]
}

func randSuiteStatus(rnd *rand.Rand) repo.SuiteStatus {
	arr := []repo.SuiteStatus{
		repo.SuiteStatusRunning,
		repo.SuiteStatusRunning,
		repo.SuiteStatusRunning,
		repo.SuiteStatusPassed,
		repo.SuiteStatusPassed,
		repo.SuiteStatusPassed,
		repo.SuiteStatusPassed,
		repo.SuiteStatusPassed,
		repo.SuiteStatusPassed,
		repo.SuiteStatusPassed,
		repo.SuiteStatusFailed,
		repo.SuiteStatusFailed,
		repo.SuiteStatusFailed,
		repo.SuiteStatusFailed,
		repo.SuiteStatusDisconnected,
	}
	return arr[rnd.Intn(len(arr))]
}

func randCaseStatus(rnd *rand.Rand) repo.CaseStatus {
	arr := []repo.CaseStatus{
		repo.CaseStatusCreated,
		repo.CaseStatusCreated,
		repo.CaseStatusCreated,
		repo.CaseStatusDisabled,
		repo.CaseStatusRunning,
		repo.CaseStatusRunning,
		repo.CaseStatusRunning,
		repo.CaseStatusRunning,
		repo.CaseStatusPassed,
		repo.CaseStatusPassed,
		repo.CaseStatusPassed,
		repo.CaseStatusPassed,
		repo.CaseStatusPassed,
		repo.CaseStatusPassed,
		repo.CaseStatusFailed,
		repo.CaseStatusFailed,
		repo.CaseStatusErrored,
	}
	return arr[rnd.Intn(len(arr))]
}

func randLogLevelType(rnd *rand.Rand) repo.LogLevelType {
	arr := []repo.LogLevelType{
		repo.LogLevelTypeTrace,
		repo.LogLevelTypeTrace,
		repo.LogLevelTypeDebug,
		repo.LogLevelTypeDebug,
		repo.LogLevelTypeDebug,
		repo.LogLevelTypeDebug,
		repo.LogLevelTypeDebug,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeInfo,
		repo.LogLevelTypeWarn,
		repo.LogLevelTypeWarn,
		repo.LogLevelTypeError,
	}
	return arr[rnd.Intn(len(arr))]
}
