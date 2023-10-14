package portfolio

import (
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID                uint                 `gorm:"primaryKey" json:"id"`
	CommodityType     config.CommodityType `json:"commodity_type"`
	ParentCommodityID string               `json:"parent_commodity_id"`
	SecurityID        string               `json:"security_id"`
	SecurityName      string               `json:"security_name"`
	SecurityType      string               `json:"security_type"`
	SecurityRating    string               `json:"security_rating"`
	SecurityIndustry  string               `json:"security_industry"`
	Percentage        decimal.Decimal      `json:"percentage"`
}

func UpsertAll(db *gorm.DB, commodityType config.CommodityType, parentCommodityID string, portfolios []*Portfolio) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Delete(&Portfolio{}, "commodity_type = ? and parent_commodity_id = ?", commodityType, parentCommodityID).Error
		if err != nil {
			return err
		}
		for _, portfolio := range portfolios {
			err := tx.Create(portfolio).Error
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func GetPortfolios(db *gorm.DB, parentCommodityID string) []Portfolio {
	var portfolios []Portfolio
	result := db.Model(&Portfolio{}).Where("parent_commodity_id = ?", parentCommodityID).Find(&portfolios)

	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return portfolios
}

func GetAllParentCommodityIDs(db *gorm.DB) []string {
	var parentCommodityIDs []string
	result := db.Model(&Portfolio{}).Distinct().Pluck("parent_commodity_id", &parentCommodityIDs)

	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return parentCommodityIDs
}
