package c

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	//"github.com/labstack/gommon/log"
	"encoding/csv"
	"github.com/lpisces/postcode/cmd/serve"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Ret struct {
	Postcode string
	Addr     []string
	Status   int
}

func GetHome(c echo.Context) error {

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

	raw, err := ioutil.ReadFile(serve.Conf.Source)
	if err != nil {
		return err
	}

	node := Node{}
	json.Unmarshal(raw, &node)
	//log.Info(node.Children[0].Children[0].Children[0])

	code := c.QueryParam("code")
	r := &Ret{
		Postcode: code,
		Status:   1,
		Addr:     []string{},
	}

	if code == "" {
		return c.JSON(http.StatusOK, r)
	}

	if len(code) != 6 {
		return c.JSON(http.StatusOK, r)
	}

	for _, v := range node.Children {
		for _, vv := range v.Children {
			if strings.Contains(vv.Postcode, code) {
				r.Status = 0
				r.Addr = append(r.Addr, fmt.Sprintf("%s,%s", v.Name, vv.Name))
			}
			for _, vvv := range vv.Children {
				if strings.Contains(vvv.Postcode, code) {
					r.Status = 0
					r.Addr = append(r.Addr, fmt.Sprintf("%s,%s,%s", v.Name, vv.Name, vvv.Name))
				}
			}
		}
	}

	if len(r.Addr) == 0 && strings.HasSuffix(code, "000") {
		codeBytes := []byte(code)
		codeBytes[5] = '1'
		code = string(codeBytes)

		for _, v := range node.Children {
			for _, vv := range v.Children {
				if strings.Contains(vv.Postcode, code) {
					r.Status = 0
					r.Addr = append(r.Addr, fmt.Sprintf("%s,%s", v.Name, vv.Name))
				}
				for _, vvv := range vv.Children {
					if strings.Contains(vvv.Postcode, code) {
						r.Status = 0
						r.Addr = append(r.Addr, fmt.Sprintf("%s,%s,%s", v.Name, vv.Name, vvv.Name))
					}
				}
			}
		}
	}

	return c.JSON(http.StatusOK, r)
}

func GetCSV(c echo.Context) error {
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

	raw, err := ioutil.ReadFile(serve.Conf.Source)
	if err != nil {
		return err
	}

	node := Node{}
	json.Unmarshal(raw, &node)
	//log.Info(node.Children[0].Children[0].Children[0])

	line := []string{}
	for _, v := range node.Children {
		for _, vv := range v.Children {
			if len(vv.Children) == 0 {
				line = append(line, fmt.Sprintf("%s,%s", v.Name, vv.Name))
				continue
			}
			for _, vvv := range vv.Children {
				line = append(line, fmt.Sprintf("%s,%s,%s", v.Name, vv.Name, vvv.Name))
			}
		}
	}

	f, err := os.Create("pcd.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	w.Write([]string{"省", "市", "区"})
	for _, v := range line {
		arr := strings.Split(v, ",")
		w.Write(arr)
	}
	w.Flush()

	//return c.Stream(http.StatusOK, "text/csv", f)
	//return c.File("pcd.csv")
	return c.Attachment("pcd.csv", "pcd.csv")
}
