package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	//<a.....path="images/201807/thumb_img/2223_thumb_G_1531956513301.jpg" title="11111">
	ReImg = `<img[\s\S]+?src="(images[\s\S]+?)"`
	ReImgName = `<a.+?path="(.+?)">`
	ReTitle = `title="(.+)`
	chSem = make(chan int,5)
	downloadWG sync.WaitGroup
	randomMT sync.Mutex
	chImgMaps chan map[string]string
	)

func GetPageImgurls(url string) []string {
	//i := "1"
	//url := "https://www.yuebing.com/category-0-b0-min0-max0-attr0-" + i + "-sort_order-ASC.html"
	html := GetHtml(url)

	re := regexp.MustCompile(ReImgName)
	rets := re.FindAllStringSubmatch(html, -1)

	imgUrls := make([]string, 0)
	for _, ret := range rets {
		imgUrl := "https://www.yuebing.com/"+ret[1]
		imgUrls = append(imgUrls, imgUrl)
	}
	return imgUrls
}

func main() {
	for i:=1;i<=15;i++{
		j := strconv.Itoa(i)
		url := "https://www.yuebing.com/category-0-b0-min0-max0-attr0-" + j + "-sort_order-ASC.html"
		imginfos := GetPageImginfos(url)
		for _,imgInfoMap := range imginfos{
			DownloadImgAsync(imgInfoMap["url"],imgInfoMap["filename"],&downloadWG)
			time.Sleep(500 * time.Millisecond)
		}
	}
	downloadWG.Wait()
}

func DownloadImg(url string,file_name string)  {
	fmt.Println("downloading")
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	filename := `路径` + file_name + ".jpg"
	imgBytes, _ := ioutil.ReadAll(resp.Body)
	err := ioutil.WriteFile(filename, imgBytes, 0644)
	if err == nil{
		fmt.Println(filename+"下载成功！")
	}else{
		fmt.Println(filename+"下载失败！")
	}
}

func DownloadImgAsync(url ,filename string,wg *sync.WaitGroup)  {
	wg.Add(1)
	go func() {
		chSem <- 1
		DownloadImg(url,filename)
		<-chSem
		downloadWG.Done()
	}()

}

func GetHtml(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	html := string(bytes)
	return html
}

func GetRandomInt(start,end int) int {
	randomMT.Lock()
	<- time.After(1 * time.Nanosecond)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := start + r.Intn(end - start)
	randomMT.Unlock()
	return ret

}
/*
生成时间戳_随机数的文件名
 */
func GetRandomName() string {
	timestamp := strconv.Itoa(int(time.Now().UnixNano()))
	randomNum := strconv.Itoa(GetRandomInt(100, 10000))
	return timestamp + "-" + randomNum
}

func GetImgNameTag(imgTag string) string {
	re := regexp.MustCompile(ReTitle)
	rets := re.FindAllStringSubmatch(imgTag, -1)
	//fmt.Println(rets)
	if len(rets) > 0{
		return rets[0][1]
	}else {
		return GetRandomName()
	}
}

func GetPageImginfos(url string) []map[string] string {
	html := GetHtml(url)

	re := regexp.MustCompile(ReImgName)
	rets := re.FindAllStringSubmatch(html, -1)
	imgInfos := make([]map[string] string,0)
	for _,ret := range rets {
		imgInfo := make(map[string] string)
		imgUrl := "https://www.yuebing.com/"+ret[1]
		imgInfo["url"] = imgUrl[0:78]
		imgInfo["filename"]=GetImgNameTag(ret[1])

		//fmt.Println(imgInfo["filename"])

		imgInfos = append(imgInfos, imgInfo)

	}
	return imgInfos
}