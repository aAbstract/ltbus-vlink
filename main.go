package main

import (
	"LTBus_VLink/lib"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	lib.LTBus_VLink_Net_Init(&wg)
	lib.LTBus_VLink_Device_Init(&wg)
	lib.LTBus_VLink_Shell_Init(&wg)
	wg.Wait()
}
