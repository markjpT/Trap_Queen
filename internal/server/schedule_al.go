package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ScheduleALSection struct {
	Code    string `json:"code"`
	Section string `json:"section"`
	Details string `json:"details"`
}

var Sections []ScheduleALSection

func init() {
	Sections = []ScheduleALSection{
		{Code: "immovable", Section: "A (1)", Details: "Immovable Assets"},
		{Code: "metal", Section: "B (1) (i)", Details: "Jewellery, bullion etc"},
		{Code: "art", Section: "B (1) (ii)", Details: "Archaeological collections, drawings, painting, sculpture or any work of art"},
		{Code: "vehicle", Section: "B (1) (ii)", Details: "Vehicles, yachts, boats and aircrafts"},
		{Code: "bank", Section: "B (1) (iv) (a)", Details: "Financial assets: Bank (including all deposits)"},
		{Code: "share", Section: "B (1) (iv) (b)", Details: "Financial assets: Shares and securities"},
		{Code: "insurance", Section: "B (1) (iv) (c)", Details: "Financial assets: Insurance policies"},
		{Code: "loan", Section: "B (1) (iv) (d)", Details: "Financial assets: Loans and advances given"},
		{Code: "cash", Section: "B (1) (iv) (e)", Details: "Financial assets: Cash in hand"},
		{Code: "liability", Section: "C (1)", Details: "Liabilities"},
	}
}

type ScheduleALEntry struct {
	Section ScheduleALSection `json:"section"`
	Amount  decimal.Decimal   `json:"amount"`
}

type ScheduleAL struct {
	Entries []ScheduleALEntry `json:"entries"`
	Date    time.Time         `json:"date"`
}

func GetScheduleAL(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%", "Liabilities:%").All()
	var scheduleALs map[string]ScheduleAL = make(map[string]ScheduleAL)

	start := utils.Now().AddDate(1, 0, 0)

	for {
		start = utils.BeginningOfFinancialYear(start)
		postings = lo.Filter(postings, func(p posting.Posting, _ int) bool { return p.Date.Before(start) })
		if len(postings) == 0 {
			break
		}

		start = start.AddDate(0, 0, -1)
		scheduleALs[utils.FYHuman(start)] = ScheduleAL{Entries: computeScheduleAL(postings), Date: start}
	}

	return gin.H{"schedule_als": scheduleALs}
}

func computeScheduleAL(postings []posting.Posting) []ScheduleALEntry {
	scheduleALConfigs := config.GetConfig().ScheduleALs

	return lo.Map(Sections, func(section ScheduleALSection, _ int) ScheduleALEntry {
		config, found := lo.Find(scheduleALConfigs, func(scheduleALConfig config.ScheduleAL) bool {
			return scheduleALConfig.Code == section.Code
		})

		var amount decimal.Decimal

		if found {
			ps := accounting.FilterByGlob(postings, config.Accounts)
			if section.Code == "liability" {
				ps = lo.Map(ps, func(p posting.Posting, _ int) posting.Posting {
					return p.Negate()
				})
			}
			amount = accounting.CostBalance(ps)
		} else {
			amount = decimal.Zero
		}

		return ScheduleALEntry{
			Section: section,
			Amount:  amount,
		}
	})

}
