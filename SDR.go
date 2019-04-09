package splitdownload
import (
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"fmt"
	"math"
	"strconv"
	"sync"
	"io/ioutil"
)
//SDR - Split Downloader
type SDR struct{
	NoOfParts int
	DownloadLink string
}
type bytesrange struct{
	order int
	start uint64
	end uint64
}

//PartialDownload - Download a part of file from internet
func (downloader SDR)PartialDownload( bytes [2]uint64, saveas string, saveto string){
	downloadLink, _ := url.Parse(downloader.DownloadLink)
	if saveas == ""{
		temp := strings.Split(downloadLink.Path, "/")
		saveas = temp[len(temp) - 1]
	}
	if _, err := os.Stat(saveto); err != nil{
		os.MkdirAll(saveto, os.ModeDir)
	}
	chunkSize := uint64(math.Ceil(float64(bytes[1] - bytes[0]) / float64(downloader.NoOfParts)))
	output := make([][]byte, downloader.NoOfParts)
	var waitgroup sync.WaitGroup
	var jobs = make([]bytesrange, downloader.NoOfParts)
	index := 0
	for start := bytes[0]; start < bytes[1]; start += chunkSize{
		end := start + chunkSize - 1
		if end > bytes[1] + 1{
			end = bytes[1]
		}
		jobs[index] = bytesrange{index, start, end}
		index++
	}
	for _, job := range jobs{
		waitgroup.Add(1)
		go func(job bytesrange){
			defer waitgroup.Done()
			var request *http.Request
			request, _ = http.NewRequest("GET", downloader.DownloadLink, nil)
			request.Header.Set("Range", "bytes="+strconv.FormatUint(job.start, 10)+"-"+strconv.FormatUint(job.end, 10))
			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil{
				fmt.Println("Error Downloading ",request)
			}
			defer response.Body.Close()
			body, _ := ioutil.ReadAll(response.Body)
			output[job.order] = body
		}(job)
	}
	waitgroup.Wait()
	final, err := os.OpenFile(path.Join(saveto,saveas), os.O_APPEND | os.O_CREATE | os.O_TRUNC, 0600)
	if err != nil{
		fmt.Println(err,"ERROR OPENING FILE")
	}
	defer final.Close()
	for i:=0; i<downloader.NoOfParts; i++{
		final.Write(output[i])
	}
}
//CompleteDownload - Download the Whole file from internet
func (downloader SDR)CompleteDownload(saveas string, saveto string){
	response, _ := http.Head(downloader.DownloadLink)
	downloader.PartialDownload([2]uint64{0, uint64(response.ContentLength)}, saveas, saveto)
}
