package scraper

import (
	"./moovit"
	"./openstreetmap"
	"./waze"
	"./car2go"
	"./enjoy"
	"./sharengo"
	"../util"

	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

type JsonRequestsFile []struct {
	From    []string `json:"from"`
	To      []string `json:"to"`
	Comment string   `json:"_comment"`
}

type ResultFile struct {
	Id               int
	Date             string
	ResultsSha256Sum string
	Results          []Result
}

type Result struct {
	FromLat, FromLon string
	ToLat, ToLon     string
	BigResult        BigJson
}

type BigJson struct {
	MoovitRoutes            moovit.Result
	OpenStreetMapBikeRoutes openstreetmap.Result
	OpenStreetMapFootRoutes openstreetmap.Result
	WazeRoutes              waze.Result
	Car2GoRoutes            car2go.Result
	EnjoyRoutes             enjoy.Result
	SharengoRoutes          sharengo.Result
}

func GetRoutesFromAllServices(fromLat, fromLon, toLat, toLon string, carsharingDataDir string) BigJson {
	return BigJson{
		moovit.GetRoutes(fromLat, fromLon, toLat, toLon),
		openstreetmap.GetBikeRoutes(fromLat, fromLon, toLat, toLon),
		openstreetmap.GetFootRoutes(fromLat, fromLon, toLat, toLon),
		waze.GetRoutes(fromLat, fromLon, toLat, toLon),
		car2go.GetRoutes(fromLat, fromLon, toLat, toLon, carsharingDataDir),
		enjoy.GetRoutes(fromLat, fromLon, toLat, toLon, carsharingDataDir),
		sharengo.GetRoutes(fromLat, fromLon, toLat, toLon, carsharingDataDir)}
}

func ReadRequests(f string) JsonRequestsFile {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println("scrapemaster: can't open request file " + f + ":", err)
		os.Exit(-1)
	}

	requests := JsonRequestsFile{}
	if err := json.Unmarshal(data, &requests); err != nil {
		fmt.Println("scrapemaster: can't unmarshal request file " + f + ":", err)
		os.Exit(-1)
	}
	return requests
}

func SaveResult(rf ResultFile, outputDir string) {
	resultFile := createFile(outputDir + "/" + util.FormattedData(time.Now()) + ".json")

	// checksum results
	data, _ := json.Marshal(rf.Results)
	rf.ResultsSha256Sum = fmt.Sprintf("%x", sha256.Sum256(data))

	// write and close
	data, _ = json.Marshal(rf)
	if _, err := resultFile.Write(data); err != nil {
		log.Println("scrapemaster: can't write to file", resultFile.Name(), err)
	}
	if err := resultFile.Close(); err != nil {
		log.Println("scrapemaster: can't close file", resultFile.Name(), err)
	}

	// compress
	compressFile(resultFile.Name())
}

func RefreshSessions() {
	for err := moovit.GetWebPage(); err != nil; err = waze.GetWebPage() {
		log.Println("moovit: failed to get webpage:", err)
		time.Sleep(1 * time.Minute)
	}


	for err := waze.GetWebPage(); err != nil; err = waze.GetWebPage() {
		log.Println("waze: failed to get webpage:", err)
		time.Sleep(1 * time.Minute)
	}

	for err := openstreetmap.GetWebPage(); err != nil; openstreetmap.GetWebPage() {
		log.Println("openstreetmap: failed to get webpage:", err)
		time.Sleep(1 * time.Minute)
	}
}

func createFile(f string) *os.File {
	file, err := os.OpenFile(f, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("scrapemaster: can't open/create file " + f + ":", err)
		os.Exit(1)
	}
	return file
}

func compressFile(f string) {
	// exec and wait
	cmd := exec.Command("tar", "-cJf", f + ".tar.xz", f)
	if err := cmd.Start(); err != nil {
		log.Println("scrapemaster: can't compress file " + f + ":", err)
		os.Exit(-1)
	}
	if err := cmd.Wait(); err != nil {
		log.Println("scrapemaster: command error:", err)
		os.Exit(-1)
	}

	// remove
	if err := os.Remove(f); err != nil {
		log.Println("scrapemaster: can't delete uncompressed file" + f + ":", err)
		os.Exit(-1)
	}
}