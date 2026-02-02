package instances

import (
	"jjmc/mods/spiget"
)

const SpigetBaseURL = spiget.SpigetBaseURL

type SpigetClient = spiget.Client
type SpigetResource = spiget.Resource
type SpigetAuthor = spiget.Author
type SpigetVersion = spiget.Version

func NewSpigetClient() *SpigetClient {
	return spiget.New()
}
