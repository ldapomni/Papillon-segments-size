package main

import (
	"bufio"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const (
	versionNumber = "1.0"
	versionInfo   = "LDA Papillon DBH status size"
	osWindows     = 1
	osLinux       = 2
	//zabbixHost    = "localhost"
	//zabbixPort    = 10051
)

type scanStore struct {
	sync.Mutex
	scanResult []string
}

func (ds *scanStore) save(str string) {
	ds.Lock()
	ds.scanResult = append(ds.scanResult, str)
	ds.Unlock()
}

type SegmentDBH struct {
	base     string
	segment  string
	path1    string
	path2    string
	path3    string
	stype    string
	tsize    string
	status   string
	placeall string
	place    string
	flags    string
	cluster  int
	segFiles int
	segSize  int
	elapsed  float64
	path     string
	procent  int
}

var wgMB sync.WaitGroup //dir scan gourotine
var pathChannel chan int
var cmdChannel chan string
var curSegment SegmentDBH
var segments []SegmentDBH
var procMax int = 4 // runtime.NumCPU()
var debug int = 0
var zabbixHost string
var zabbixName string
var zabbixPort int

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}
func main() {
	//! ssh mb01  "du -sb /papillon1.db/04f80001.ss" в байтах

	var papillonDBHFile string
	var segmentsTotal int = 0

	var startTime time.Time

	path, err := os.Executable()
	if err != nil {
		log.Println(err)
	}
	papillonDBHFile = filepath.Dir(path) + string(os.PathSeparator) + "papillon.dbh"
	papillonDBHFile = "/papillon1/conf/papillon.dbh"

	dsnenv, exists := os.LookupEnv("PAPILLON_DSN")
	if !exists {
		log.Println("No DSN in .env")
	}

	flag.StringVar(&zabbixName, "zabbix_name", "Papillon1.DB", "Zabbix monitoring hostname")
	flag.StringVar(&zabbixHost, "zabbix_host", "zabbix.mvd.udm.ru", "Zabbix server")
	flag.IntVar(&zabbixPort, "zabbix_port", 10051, "Zabbix port")
	dsn := flag.String("dsn", dsnenv, "MySQL DSN String - user:pass@tcp(ip:3306)/base")
	//	flag.StringVar(&dsn,"dsn", "root@tcp(db.mvd.udm.ru:3306)/lscan", "MySQL DSN String - user:pass@tcp(ip:3306)/base")
	flag.StringVar(&papillonDBHFile, "p", papillonDBHFile, "Locate Papillon dbh file, default /papillon1/conf/papillon.dbh")
	flag.IntVar(&debug, "d", 0, "Debug level: 0-off,1-on")
	flag.Parse()

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		log.Fatalln(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	defer db.Close()

	dbh, err := os.Open(papillonDBHFile)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(dbh)

	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		s = strings.Join(strings.Fields(s), " ")
		arr := strings.Split(s, "#")
		s = arr[0]
		arr = strings.Split(s, " ")
		if len(arr) < 7 {
			continue
		}
		if !strings.HasPrefix(s, "#") {
			//	fmt.Println(s)
			curSegment.cluster = 1
			curSegment.flags = ""
			curSegment.elapsed = 0
			curSegment.segFiles = 0
			curSegment.segSize = 0
			curSegment.status = ""

			curSegment.base = arr[0]
			curSegment.segment = arr[1]
			curSegment.path1 = arr[2]

			curSegment.path2 = arr[3]
			curSegment.path3 = arr[4]
			curSegment.stype = arr[5]
			curSegment.tsize = arr[6]
			for i := 7; i < len(arr); i++ {
				t := strings.Split(arr[i], ":")
				switch t[0] {
				case "e", "o":
					curSegment.status = t[0]
				case "d", "x", "g", "j", "i", "r", "u", "k", "s", "y":
					curSegment.flags = curSegment.flags + t[0]
				case "c":
					if len(t) == 1 {
						curSegment.cluster = 1
					} else {
						curSegment.cluster, err = strconv.Atoi(t[1])
						if err != nil {
							log.Fatalln(err)
						}
					}
				case "b":
					curSegment.placeall = t[1]
					l := strings.Split(t[1], ",")
					curSegment.place = l[0]
				}
			}
			curSegment.path = curSegment.path1 + "/" + curSegment.base + curSegment.segment + "."
			if curSegment.cluster == 1 {
				curSegment.path = curSegment.path + "s"
			} else {
				curSegment.path = curSegment.path + "ss"
			}

			segments = append(segments, curSegment)
			segmentsTotal++
		}

	}
	//fmt.Println(segments)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	startTime = time.Now()

	pathChannel = make(chan int)
	cmdChannel = make(chan string)
	registerShutdown(cmdChannel, &wgMB)

	if debug == 1 {
		log.Println("Start goroutins")
	}
	for i := 0; i < procMax; i++ {
		wgMB.Add(1)
		go scanMB(i, &segments, pathChannel, cmdChannel, &wgMB)
	}
	if debug == 1 {
		log.Println("Start checks")
	}
	for idx, _ := range segments {
		//	fmt.Println(idx)
		pathChannel <- idx
	}

	shutdownProcesses("End work", cmdChannel, &wgMB)
	elapsed := int(time.Since(startTime) / time.Second)
	if debug == 1 {
		log.Println(elapsed)
		log.Println(procMax)
	}
	writeBase(db, &segments)

}

func registerShutdown(cmdChannel chan string, wg *sync.WaitGroup) {

	sys := make(chan os.Signal, 1)
	signal.Notify(sys, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		for s := range sys {

			if debug == 1 {
				log.Printf("got signal: %v\n", s)
				log.Println("Shutdown workers")
			}
			shutdownProcesses("interrupt", cmdChannel, wg)
			if debug == 1 {
				log.Println("Exit interrupt")
			}
			os.Exit(0)

		}
	}()
}
func shutdownProcesses(cmd string, cmdChannel chan string, wg *sync.WaitGroup) {
	for i := 0; i < procMax; i++ {
		cmdChannel <- "shutdown"
	}
	wgMB.Wait()
}
