package main

import (
	"./lib/scraper/moovit"
	"./lib/scraper/openstreetmap"
	"./lib/scraper/waze"
	"./lib/scraper/car2go"
	"./lib/scraper/enjoy"
	"./lib/scraper/sharengo"
	"./lib/trip"
	"./lib/util"

	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	DEBUGGING = false
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
	BigResult        trip.BigJson
}

func main() {
	// Arguments
	if len(os.Args) < 4 {
		fmt.Println("usage:\n\t scrapemaster [richieste.json] [scraped_data_dir] [output_dir]")
		os.Exit(0)
	}
	jsonRequestsFilename := os.Args[1]
	scrapedDataDir := os.Args[2]
	outputDir := os.Args[3]

	// Logfile
	logfile := createFile(outputDir, "scraper_error.log")
	log.SetOutput(logfile)
	defer logfile.Close()

	// Read requests
	requests := readJsonRequests(jsonRequestsFilename)
	for i := 0; true; i++ {
		resultFilename := "scrapemaster_" + util.FormattedData(time.Now()) + ".json"
		resultFile := createFile(outputDir, resultFilename)
		var resultFileStruct ResultFile
		resultFileStruct.Id = i
		resultFileStruct.Date = time.Now().Format("2006-01-02 15:04:05")

		for _, request := range requests {
			var resultStruct Result
			resultStruct.FromLat = request.From[0]
			resultStruct.FromLon = request.From[1]
			resultStruct.ToLat = request.To[0]
			resultStruct.ToLon = request.To[1]
			resultStruct.BigResult = getRoutesFromAllServices(
				resultStruct.FromLat, resultStruct.FromLon,
				resultStruct.ToLat, resultStruct.ToLon,
				scrapedDataDir)

			resultFileStruct.Results = append(resultFileStruct.Results, resultStruct)

			time.Sleep(30 * time.Second)
		}

		var data []byte
		if DEBUGGING { // debugging ? indent : don't indent, save space
			data, _ = json.MarshalIndent(resultFileStruct.Results, "", "\t")
			resultFileStruct.ResultsSha256Sum = fmt.Sprintf("%x", sha256.Sum256(data))
			data, _ = json.MarshalIndent(resultFileStruct, "", "\t")
		} else {
			data, _ = json.Marshal(resultFileStruct.Results)
			resultFileStruct.ResultsSha256Sum = fmt.Sprintf("%x", sha256.Sum256(data))
			data, _ = json.Marshal(resultFileStruct)
		}

		if _, err := resultFile.Write(data); err != nil {
			log.Println("scrapemaster: can't write to file" + outputDir + "/" + resultFilename, err)
		}

		if err := resultFile.Close(); err != nil {
			log.Println("scrapemaster: can't close file" + outputDir + "/" + resultFilename, err)
		}

		if err := compressFile(outputDir, resultFilename); err != nil {
			log.Println("scrapemaster: can't compress file" + outputDir + "/" + resultFilename, err)
		} else {
			if err := os.Remove(outputDir + "/" + resultFilename); err != nil {
				log.Println("scrapemaster: can't delete file" + outputDir + "/" + resultFilename, err)
			}
		}

		// emp, _ := json.MarshalIndent(results, "", "  ")
		// fmt.Println(string(emp))
		time.Sleep(10 * time.Minute)
	}
}

func getRoutesFromAllServices(fromLat, fromLon, toLat, toLon string, scrapedDataDir string) trip.BigJson {
	moovitResult := moovit.GetRoutes(fromLat, fromLon, toLat, toLon)
	openstreetmapBikeResult := openstreetmap.GetBikeRoutes(fromLat, fromLon, toLat, toLon)
	openstreetmapFootResult := openstreetmap.GetFootRoutes(fromLat, fromLon, toLat, toLon)
	wazeResult := waze.GetRoutes(fromLat, fromLon, toLat, toLon)
	car2goResult := car2go.GetRoutes(fromLat, fromLon, toLat, toLon, scrapedDataDir)
	sharengoResult := sharengo.GetRoutes(fromLat, fromLon, toLat, toLon, scrapedDataDir)
	enjoyResult := enjoy.GetRoutes(fromLat, fromLon, toLat, toLon, scrapedDataDir)

	var results trip.BigJson
	results.MoovitRoutes = moovitResult
	results.OpenStreetMapBikeRoutes = openstreetmapBikeResult
	results.OpenStreetMapFootRoutes = openstreetmapFootResult
	results.WazeRoutes = wazeResult
	results.Car2GoRoutes = car2goResult
	results.EnjoyRoutes = enjoyResult
	results.SharengoRoutes = sharengoResult
	return results
}

func compressFile(dir, filename string) error {
	// pwd, cd
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Println("scrapemaster: cd failed", err)
		return err
	}

	// exec, wait
	cmd := exec.Command("tar", "-cJf", filename + ".tar.xz", filename)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	// cd ..
	if err := os.Chdir(currentDir); err != nil {
		fmt.Println("scrapemaster: pwd failed", err)
		return err
	}
	return nil
}

func readJsonRequests(jsonFilename string) JsonRequestsFile {
	// read file
	requestsData, err := ioutil.ReadFile(jsonFilename)
	if err != nil {
		fmt.Println("scrapemaster: can't open input file " + jsonFilename + ":", err)
		os.Exit(1)
	}

	// unmarshal
	var requests JsonRequestsFile
	if err := json.Unmarshal([]byte(requestsData), &requests); err != nil {
		fmt.Println("scrapemaster: can't unmarshal json file " + jsonFilename + ":", err)
		os.Exit(1)
	}
	return requests
}

func createFile(dir, filename string) *os.File {
	// pwd, cd
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("scrapemaster: pwd failed:", err)
		os.Exit(1)
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Println("scrapemaster: chdir failed:", err)
		os.Exit(1)
	}

	// open append / create append
	file, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("scrapemaster: can't open/create file" + dir + "/" + filename + ":", err)
		os.Exit(1)
	}

	// cd ..
	if err := os.Chdir(currentDir); err != nil {
		fmt.Println("scrapemaster: chdir failed:", err)
		os.Exit(1)
	}
	return file
}