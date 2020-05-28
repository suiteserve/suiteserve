package seed

import (
	"context"
	"github.com/tmazeika/testpass/repo"
)

const attachmentCount = 30
const suiteCount = 28

func Seed(repos repo.Repos) error {
	go func() {
		for {
			if _, ok := <-repos.Changes(); !ok {
				break
			}
			// ignore changes
		}
	}()

	seededAttachments := make([]*repo.AttachmentInfo, 0)
	for i := 0; i < attachmentCount; i++ {
		seed := attachments[i%len(attachments)]
		a := seed.a
		src := seed.srcFn()

		file, err := repos.Attachments().Save(context.Background(), a, src)
		if err != nil {
			return err
		}
		seededAttachments = append(seededAttachments, file.Info())

		if i%20 == 0 {
			deletedAt := 1590625822618 + int64(i) * 105000
			err := repos.Attachments().Delete(context.Background(),
				file.Info().Id, deletedAt)
			if err != nil {
				return err
			}
			file.Info().Deleted = true
			file.Info().DeletedAt = deletedAt
		}
	}

	for i := 0; i < suiteCount; i++ {
		seedSuite := suites[i%len(suites)]

		seedSuite.StartedAt = 1590627102982 + int64(i) * 100230
		if i%26 == 0 {
			seedSuite.Deleted = true
			seedSuite.DeletedAt = 1590626812841 + int64(i) * 315300
		}
		if i%6 > 0 {
			seedSuite.Attachments = append(seedSuite.Attachments,
				seededAttachments[i%len(seededAttachments)].Id)
		}
		if i%7 > 4 {
			seedSuite.Attachments = append(seedSuite.Attachments,
				seededAttachments[(i*2)%len(seededAttachments)].Id)
		}

		suiteId, err := repos.Suites().Save(context.Background(), seedSuite)
		if err != nil {
			return err
		}

		for j := 0; j < int(seedSuite.PlannedCases); j++ {
			seedCase := cases[j%len(cases)]

			seedCase.Suite = suiteId
			seedCase.Num = int64(j) % 12
			if i%8 > 0 {
				seedCase.Attachments = append(seedCase.Attachments,
					seededAttachments[i%len(seededAttachments)].Id)
			}
			if i%10 > 3 {
				seedCase.Attachments = append(seedCase.Attachments,
					seededAttachments[(i*3)%len(seededAttachments)].Id)
			}

			seedCase.CreatedAt = 1590628126100 + int64(j) * 307640
			if seedCase.Status != repo.CaseStatusCreated {
				seedCase.StartedAt = 1590628204350 + int64(j) * 1465000
			}
			if seedCase.Status != repo.CaseStatusCreated &&
				seedCase.Status != repo.CaseStatusRunning {
				seedCase.FinishedAt = 1590628243183 + int64(j) * 2133000
			}

			caseId, err := repos.Cases().Save(context.Background(), seedCase)
			if err != nil {
				return err
			}

			for k := 0; k < j; k++ {
				seedLogEntry := logEntries[k%len(logEntries)]

				seedLogEntry.Case = caseId
				seedLogEntry.Index = int64(k)
				seedLogEntry.Timestamp = 1590629158629 + int64(k) * 1511300

				_, err := repos.Logs().Save(context.Background(), seedLogEntry)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
