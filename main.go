package main

import (
	_ "huobicagent/config"
	"huobicagent/domainservice/crawlerservice"
)

func main() {
	crawlerservice.HuobiChromedp()
}
