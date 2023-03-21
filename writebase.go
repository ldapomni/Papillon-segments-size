package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	. "github.com/blacked/go-zabbix"
)

type baseSegType struct {
	files   int
	cluster int
	size    int64
}

func writeBase(db *sql.DB, seg *[]SegmentDBH) {
	var metrics []*Metric
	var discovery []*Metric
	var baseseg []string
	baseprocent := make(map[string]int)
	baseData := make(map[string]baseSegType)

	sel, err := db.Prepare("select size,files from segments where base=? and seg=?")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer sel.Close()

	ins, err := db.Prepare("insert into segments (base,seg,type,tsize,status,flags,membox,macro,size,files,date,procent,fadd,sadd) " +
		"values (?,?,?,?,?,?,?,?,?,?,?,?,?,?) on duplicate key update " +
		"type=?,tsize=?,status=?,flags=?,membox=?,macro=?,size=?, files=?,date=?,procent=?,fadd=?,sadd=?")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer ins.Close()
	var fadd int = 0
	var sadd int = 0
	//var en SegmentDBH
	for _, en := range *seg {
		err := sel.QueryRow(en.base, en.segment).Scan(&sadd, &fadd)
		if err != nil {
			if err == sql.ErrNoRows {
				fadd = 0
				sadd = 0
			} else {
				log.Fatalln(err.Error())
			}
		}
		procent := 0
		el := en.base + "_" + en.stype + "_" + en.flags
		switch en.status {
		case "e":
			procent = 0
			///?????
			if t, ok := baseData[el]; ok {
				t.files = t.files + en.segFiles
				t.cluster = t.cluster + en.cluster
				t.size = t.size + int64(en.segSize)
			}
		case "o":
			procent = 100
		case "":
			if en.tsize == "m635" {
				if en.segSize > 0 {
					procent = int(float64(en.segSize) / (float64(en.cluster) * 635 * 1024 * 1204) * 100)
				} else {
					procent = 0
				}
			}
			if en.tsize == "n1000" {
				if en.segFiles > 0 {
					procent = int(float64(en.segFiles) / (float64(en.cluster) * 1000) * 100)
				} else {
					procent = 0
				}
			}

		}
		if debug == 1 {
			log.Printf("Segment %s,%s files - %d, size - %d,procent - %d", en.base, en.segment, en.segFiles, en.segSize, procent)
		}
		en.segSize = int(float64(en.segSize) / (1024 * 1024)) //in Mb

		datetime := time.Now()
		date := datetime.Format("2006-01-02 15:04:05")
		fadd := en.segFiles - fadd
		sadd := en.segSize - sadd
		en.procent = procent
		_, err = ins.Exec(en.base, en.segment, en.stype, en.tsize, en.status, en.flags, en.placeall, en.cluster, en.segSize, en.segFiles, date, procent, fadd, sadd,
			en.stype, en.tsize, en.status, en.flags, en.placeall, en.cluster, en.segSize, en.segFiles, date, procent, fadd, sadd)

		if err != nil {
			log.Fatalln(err.Error())

		}
		if en.status != "o" {
			el := en.base + "_" + en.stype
			if t, ok := baseprocent[el]; ok {
				if procent > t {
					baseprocent[el] = procent
				}
			} else {
				baseprocent[el] = procent
			}
		}

	}

	for base, pr := range baseprocent {
		baseseg = append(baseseg, "{\"{#ITEMNAME}\": \""+base+"\"}")
		metrics = append(metrics, NewMetric(zabbixName, "base["+base+"]", strconv.Itoa(pr), time.Now().Unix()))
	}
	s := strings.Join(baseseg, ",")
	s = "{\"data\":[" + s + "] }"

	//Discovery
	// Create instance of Packet class
	// Send packet to zabbix
	discovery = append(discovery, NewMetric(zabbixName, "base.discovery", s, time.Now().Unix()))
	packetDiscovery := NewPacket(discovery)
	zD := NewSender(zabbixHost, zabbixPort)
	zD.Send(packetDiscovery)

	//Data
	// Create instance of Packet class
	// Send packet to zabbix

	packet := NewPacket(metrics)
	z := NewSender(zabbixHost, zabbixPort)
	z.Send(packet)
}
