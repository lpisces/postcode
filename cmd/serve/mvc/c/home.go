package c

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	//"github.com/labstack/gommon/log"
	"github.com/lpisces/postcode/cmd/serve"
	"io/ioutil"
	"net/http"
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
	return c.JSON(http.StatusOK, r)
}
