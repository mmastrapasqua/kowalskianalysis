// Date 2 coordinate (latitudine, longitudine) di partenza e arrivo,
// lo scraper chiede a Moovit di calcolare in tempo reale il miglior percorso
// coi mezzi pubblici per percorrere la tratta

package main

import (
	"../lib/moovit"
	"../lib/geoloc"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("usage:\n\t moovit <fromLat> <fromLon> <toLat> <toLon>")
		return
	}
	from := geoloc.Location{os.Args[1], os.Args[2], "unknown"}
	to := geoloc.Location{os.Args[3], os.Args[4], "unkown"}

	_ = moovit.GetRealTimeRoutes(from, to)
}
