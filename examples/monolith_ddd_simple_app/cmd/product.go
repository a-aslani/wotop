package cmd

import (
	"github.com/a-aslani/wotop/examples/monolith_ddd_simple_app/configs"
	"github.com/a-aslani/wotop/logger"
	"github.com/a-aslani/wotop/wotop"
)

type product struct{}

func NewProduct() wotop.wotop[configs.Config] {
	return &product{}
}

func (p product) Run(cfg *configs.Config) error {

	const appName = "product"

	appData := wotop.NewApplicationData(appName)

	_ = appData

	log, err := logger.NewGrayLog(cfg.GraylogAddr, cfg.Stage)
	if err != nil {
		return err
	}

	defer log.Sync()

	return nil

}

func (product) registerUsecase(
	log logger.Logger,
) {

}
