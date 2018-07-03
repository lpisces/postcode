package collect

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/mozillazg/go-pinyin"
	"github.com/tidwall/gjson"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BaseUrl   = "http://v.juhe.cn/postcode/"
	Retry     = 5
	MaxSpider = 10
)

type Node struct {
	ID       string  `json:"id"`
	ParentID string  `json:"parentId"`
	Name     string  `json:"value"`
	Type     string  `json:"type"`
	Postcode string  `json:"postnumber"`
	Pinyin   string  `json:"pinyin"`
	IsLeaf   bool    `json:"isleaf"`
	Children []*Node `json:"children"`
}

func Run(c *cli.Context) (err error) {

	// Load default config
	config := DefaultConfig()

	// override default config
	configFilePath := c.String("config")
	if configFilePath != "" {
		if err := config.Load(configFilePath); err != nil {
			log.Fatal(err)
		}
	}

	// flag override ini file config
	env := c.String("env")
	if env != "" {
		config.Mode = env
	}

	appKey := c.String("key")
	if appKey != "" {
		config.AppKey = appKey
	}

	cache := c.String("cache")
	if cache != "" {
		config.Cache = cache
	}
	if _, err := os.Stat(config.Cache); os.IsNotExist(err) {
		if e := os.Mkdir(config.Cache, os.ModeDir); e != nil {
			return e
		}
	}

	output := c.String("output")
	if output != "" {
		config.Output = output
	}

	Conf = config

	// run mode
	if config.Mode != "production" {
		Debug = true
	}

	node, err := Pcd()
	if err != nil {
		log.Fatal(err)
	}
	//log.Info(node.Children[0].Children[0].Children[0])
	//enc := json.NewEncoder(os.Stdout)
	//enc.Encode(node)
	nodeJson, _ := json.Marshal(node)
	err = ioutil.WriteFile(Conf.Output, nodeJson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// Pcd pcd endpoint
func Pcd() (node *Node, err error) {

	endpoint := "pcd"

	u, _ := url.Parse(fmt.Sprintf("%s%s", BaseUrl, endpoint))
	param := url.Values{}
	param.Set("key", Conf.AppKey) //应用APPKEY(应用详细页查询)
	param.Set("dtype", "")        //返回数据的格式,xml或json，默认json

	u.RawQuery = param.Encode()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", u.String(), nil)

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	// status code check
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http code: %d", resp.StatusCode)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	// error code check
	errorCode := gjson.GetBytes(body, "error_code")
	if errorCode.Int() != 0 {
		err = fmt.Errorf("error code: %d", errorCode.Int())
		return
	}

	root := Node{ID: "0", Name: "China", ParentID: "0", Type: "N", Children: []*Node{}}

	a := pinyin.NewArgs()
	for k, _ := range gjson.GetBytes(body, "result").Array() {
		pNode := Node{
			ID:       gjson.GetBytes(body, fmt.Sprintf("result.%d.id", k)).String(),
			Name:     gjson.GetBytes(body, fmt.Sprintf("result.%d.province", k)).String(),
			Type:     "P",
			Children: []*Node{},
			ParentID: "0",
			IsLeaf:   false,
		}
		py := []string{}
		for _, v := range pinyin.Pinyin(pNode.Name, a) {
			py = append(py, v...)
		}
		pNode.Pinyin = strings.Join(py, " ")
		for kk, _ := range gjson.GetBytes(body, fmt.Sprintf("result.%d.city", k)).Array() {
			cNode := Node{
				ID:       gjson.GetBytes(body, fmt.Sprintf("result.%d.city.%d.id", k, kk)).String(),
				Name:     gjson.GetBytes(body, fmt.Sprintf("result.%d.city.%d.city", k, kk)).String(),
				Type:     "C",
				Children: []*Node{},
				ParentID: pNode.ID,
				IsLeaf:   false,
			}
			py := []string{}
			for _, v := range pinyin.Pinyin(cNode.Name, a) {
				py = append(py, v...)
			}
			cNode.Pinyin = strings.Join(py, " ")

			var wg sync.WaitGroup
			limitChan := make(chan int, MaxSpider)

			for kkk, _ := range gjson.GetBytes(body, fmt.Sprintf("result.%d.city.%d.district", k, kk)).Array() {
				dNode := Node{
					ID:       gjson.GetBytes(body, fmt.Sprintf("result.%d.city.%d.district.%d.id", k, kk, kkk)).String(),
					Name:     gjson.GetBytes(body, fmt.Sprintf("result.%d.city.%d.district.%d.district", k, kk, kkk)).String(),
					Type:     "D",
					Children: []*Node{},
					ParentID: cNode.ID,
					IsLeaf:   true,
				}
				py := []string{}
				for _, v := range pinyin.Pinyin(dNode.Name, a) {
					py = append(py, v...)
				}
				dNode.Pinyin = strings.Join(py, " ")
				wg.Add(1)
				limitChan <- 0
				go func() {
					defer func() {
						wg.Done()
						<-limitChan
					}()
					log.Infof("searching code for node %s_%s_%s", pNode.Name, cNode.Name, dNode.Name)
					var i int
					for i = Retry; i > 0; i-- {
						dNode.Postcode, err = Search(pNode.ID, cNode.ID, dNode.ID)
						if err == nil {
							break
						}
						log.Infof("retry %s_%s_%s: %d", pNode.Name, cNode.Name, dNode.Name, Retry-i)
					}
					if i == 0 {
						log.Infof("node %s_%s_%s failed", pNode.Name, cNode.Name, dNode.Name)
					}
					cNode.Children = append(cNode.Children, &dNode)
					log.Infof("get code for node %s_%s_%s succeed", pNode.Name, cNode.Name, dNode.Name)
					log.Infof("code is %s", dNode.Postcode)
				}()
			}
			wg.Wait()
			pNode.Children = append(pNode.Children, &cNode)
		}
		root.Children = append(root.Children, &pNode)
	}

	return &root, nil
}

// Search search endpoint
func Search(pid, cid, did string) (code string, e error) {
	endpoint := "search"

	get := func(page int) (body []byte, totalPage int, err error) {

		cacheFile := fmt.Sprintf("%s/%s_%s_%s_%d", Conf.Cache, pid, cid, did, page)
		if _, err := os.Stat(cacheFile); os.IsNotExist(err) {

			u, _ := url.Parse(fmt.Sprintf("%s%s", BaseUrl, endpoint))
			param := url.Values{}
			param.Set("key", Conf.AppKey)              //应用APPKEY(应用详细页查询)
			param.Set("pid", pid)                      //省份ID
			param.Set("cid", cid)                      //城市ID
			param.Set("did", did)                      //区域ID
			param.Set("pagesize", "50")                //区域ID
			param.Set("page", fmt.Sprintf("%d", page)) //区域ID

			u.RawQuery = param.Encode()

			client := &http.Client{
				Timeout: time.Duration(2 * time.Second),
			}
			req, _ := http.NewRequest("GET", u.String(), nil)

			resp, err := client.Do(req)
			if err != nil {
				return body, totalPage, err
			}

			// status code check
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("http code: %d", resp.StatusCode)
				return body, totalPage, err
			}

			body, _ = ioutil.ReadAll(resp.Body)
			ioutil.WriteFile(cacheFile, body, 0644)
			log.Info("write cache " + cacheFile)
		} else {
			body, _ = ioutil.ReadFile(cacheFile)
			log.Info("read cache " + cacheFile)
		}

		// error code check
		errorCode := gjson.GetBytes(body, "error_code")
		if errorCode.Int() != 0 {
			err = fmt.Errorf("error code: %d", errorCode.Int())
			return
		}

		tp := gjson.GetBytes(body, "result.totalpage").String()
		totalPage, err = strconv.Atoi(tp)
		if err != nil {
			return
		}
		return body, totalPage, nil
	}

	parse := func(body []byte) (code []string) {
		for _, v := range gjson.GetBytes(body, "result.list.#.PostNumber").Array() {
			code = append(code, v.String())
		}
		return
	}

	body, totalPage, e := get(1)
	if e != nil {
		return
	}

	log.Infof("total page is %d", totalPage)
	codeList := []string{}
	for i := 1; i <= totalPage; i++ {
		log.Infof("start parse page %d/%d", i, totalPage)
		if i == 1 {
			codeList = append(codeList, parse(body)...)
			continue
		}
		body, _, e := get(i)
		if e != nil {
			return "", e
		}
		codeList = append(codeList, parse(body)...)
		log.Infof("compelted parse page %d", i)
	}

	codeMap := make(map[string]string)
	codeArr := []string{}
	for _, v := range codeList {
		if _, ok := codeMap[v]; ok {
			continue
		}
		codeMap[v] = v
		codeArr = append(codeArr, v)
	}

	code = strings.Join(codeArr, ",")
	return
}
