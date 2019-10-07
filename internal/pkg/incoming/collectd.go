package incoming

import (
	"fmt"
	"log"
	"strconv"

	"collectd.org/cdtime"
	"github.com/json-iterator/go"
)

// Collectd  ...
type Collectd struct {
	Values         []float64   `json:"values"`
	Dstypes        []string    `json:"dstypes"`
	Dsnames        []string    `json:"dsnames"`
	Time           cdtime.Time `json:"time"`
	Interval       float64     `json:"interval"`
	Host           string      `json:"host"`
	Plugin         string      `json:"plugin"`
	PluginInstance string      `json:"plugin_instance"`
	Type           string      `json:"type"`
	TypeInstance   string      `json:"type_instance"`
	new            bool
}

// GetName implement interface
func (c Collectd) GetName() string {
	return c.Plugin
}

// GetKey ...
func (c Collectd) GetKey() string {
	return c.Host
}

//ParseInputByte ...
//TODO(mmagr): probably unify interface with ParseInputJSON
func (c *Collectd) ParseInputByte(data []byte) error {
	cparse := make([]Collectd, 1)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(data, &cparse)
	if err != nil {
		log.Printf("Error parsing InputByte: %s", err)
		return err
	}
	c1 := cparse[0]
	c1.SetNew(true)
	c.SetData(&c1)
	return nil
}

//SetNew ...
func (c *Collectd) SetNew(new bool) {
	c.new = new
}

//GetInterval ...
func (c *Collectd) GetInterval() float64 {
	return c.Interval
}

// ISNew   ..
func (c *Collectd) ISNew() bool {
	return c.new
}

//DSName newName converts one data source of a value list to a string
//representation.
func (c *Collectd) DSName(index int) string {
	if c.Dsnames != nil {
		return c.Dsnames[index]
	} else if len(c.Values) != 1 {
		//TODO(mmagr): verify validity of above conditional later
		return strconv.FormatInt(int64(index), 10)
	}
	return "value"
}

//SetData ...
func (c *Collectd) SetData(data DataTypeInterface) {
	if collectd, ok := data.(*Collectd); ok { // type assert on it
		if c.Host != collectd.Host {
			c.Host = collectd.Host
		}
		if c.Plugin != collectd.Plugin {
			c.Plugin = collectd.Plugin
		}
		c.Interval = collectd.Interval
		c.Values = collectd.Values
		c.Dsnames = collectd.Dsnames
		c.Dstypes = collectd.Dstypes
		c.Time = collectd.Time
		if c.PluginInstance != collectd.PluginInstance {
			c.PluginInstance = collectd.PluginInstance
		}
		if c.Type != collectd.Type {
			c.Type = collectd.Type
		}
		if c.TypeInstance != collectd.TypeInstance {
			c.TypeInstance = collectd.TypeInstance
		}
		c.SetNew(true)
	}
}

//GetLabels ...
func (c Collectd) GetLabels() map[string]string {
	labels := map[string]string{}
	if c.PluginInstance != "" {
		labels[c.Plugin] = c.PluginInstance
	}
	if c.TypeInstance != "" {
		if c.PluginInstance == "" {
			labels[c.Plugin] = c.TypeInstance
		} else {
			labels["type"] = c.TypeInstance
		}
	}
	// Make sure that "type" and c.Plugin labels always
	// exists.  Otherwise, Prometheus checks fail
	//
	if _, typeexist := labels["type"]; !typeexist {
		labels["type"] = "base"
	}
	if _, typeexist := labels[c.Plugin]; !typeexist {
		labels[c.Plugin] = "base"
	}

	labels["instance"] = c.Host

	return labels
}

//GetMetricDesc   newDesc converts one data source of a value list to a Prometheus description.
func (c Collectd) GetMetricDesc(index int) string {
	help := fmt.Sprintf("Service Assurance exporter: '%s' Type: '%s' Dstype: '%s' Dsname: '%s'",
		c.Plugin, c.Type, c.Dstypes[index], c.DSName(index))
	return help
}

//GetMetricName  ...
func (c Collectd) GetMetricName(index int) string {
	name := "sa_collectd_" + c.Plugin + "_" + c.Type
	if c.Plugin == c.Type {
		name = "sa_collectd_" + c.Type
	}

	if dsname := c.DSName(index); dsname != "value" {
		name += "_" + dsname
	}

	switch c.Dstypes[index] {
	case "counter", "derive":
		name += "_total"
	}
	return name
}

//GetItemKey  ...
func (c Collectd) GetItemKey() string {
	name := c.Plugin + "_" + c.Type
	if c.Plugin == c.Type {
		name = c.Type
	}
	if c.PluginInstance != "" {
		name += "_" + c.PluginInstance
	}
	if c.TypeInstance != "" {
		name += "_" + c.TypeInstance
	}
	return name
}

//ParseInputJSON   ...
func (c *Collectd) ParseInputJSON(jsonString string) ([]DataTypeInterface, error) {
	collect := []Collectd{}
	jsonBlob := []byte(jsonString)
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(jsonBlob, &collect)
	if err != nil {
		log.Println("Error parsing json:", err)
		return nil, err
	}
	retDtype := make([]DataTypeInterface, len(collect))
	for index, rt := range collect {
		retDtype[index] = &rt
	}
	return retDtype, nil
}
