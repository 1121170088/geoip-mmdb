package coordinate

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type Node struct {
	Adcode int
	Name   string
	Center []float64
	Parent struct {
		Adcode int
	}
	Children []*Node
}

var Nodes map[string]*Node

func Load(fname string) {
	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(bs, &Nodes)
	if err != nil {
		log.Panic(err)
	}
}

var NotFound = errors.New("")

func FindCoor(province, city, area string) (error, []float64) {
	adcode := "100000"
	contry := Nodes[adcode]
	if contry == nil {
		return NotFound, nil
	}
	if len(contry.Children) == 0 {
		return NotFound, nil
	}
	var coor []float64
	if province == "" {
		return NotFound, nil
	}
	for _, prov := range contry.Children {
		if strings.Contains(prov.Name, province) {
			coor = prov.Center
			adcode = strconv.Itoa(prov.Adcode)
			provNode, ok := Nodes[adcode]
			if !ok {
				goto here
			}
			if len(provNode.Children) == 0 {
				goto here
			}
			if city == "" {
				goto here
			}
			for _, cit := range provNode.Children {
				if strings.Contains(cit.Name, city) || strings.Contains(city, cit.Name) {
					coor = cit.Center
					adcode := strconv.Itoa(cit.Adcode)
					cityNode, ok := Nodes[adcode]
					if !ok {
						goto here
					}
					if len(cityNode.Children) == 0 {
						goto here
					}
					if area == "" {
						goto here
					}
					for _, are := range cityNode.Children {
						if strings.Contains(are.Name, area) || strings.Contains(area, are.Name) {
							coor = are.Center
							goto here
						}
					}
				}
			}
		}
	}
here:
	if coor == nil {
		return NotFound, nil
	}
	return nil, coor
}
