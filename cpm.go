package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	"gopkg.in/ini.v1"
)

const inipath = "cpm.conf"

// Configuration correspondant au information chargé depuis le fichier de configuration
type Configuration struct {
	baseURL     string
	version     string
	collections []string
	localCache  string
}

func LoadSettings(conffilepath string) Configuration {
	config_file, err := ini.Load(inipath)
	if err != nil {
		fmt.Printf("Erreur lors du chargement du fichier de configuration : %v", err)
	}
	var loadConfiguration Configuration

	loadConfiguration.baseURL = config_file.Section("Repo").Key("base_url").String()
	loadConfiguration.version = config_file.Section("Repo").Key("version").String()

	var parseCollections []string = strings.Split(config_file.Section("Repo").Key("collections").String(), " ")
	loadConfiguration.collections = parseCollections

	loadConfiguration.localCache = config_file.Section("Main").Key("local_cache").String()
	return loadConfiguration

}

func DownloadFile(url string, destdir string) {

	var parseURL []string = strings.Split(url, "/")

	done := make(chan int64)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erreur lors du téléchargement du fichier : %v", err)
	}
	fmt.Println(resp.Header.Get("Content-Length"))
	defer resp.Body.Close()

	file, errorfile := os.Create(destdir + parseURL[len(parseURL)-1])
	if errorfile != nil {
		fmt.Printf("Erreur lors de la création du fichier : %v", errorfile)
	}
	defer file.Close()

	totalsize, errconvert := strconv.Atoi(resp.Header.Get("Content-Length"))
	if errconvert != nil {
		fmt.Printf("Erreur lors du cast: %v", totalsize)
	}
	go PrintProgress(done, totalsize, destdir+parseURL[len(parseURL)-1])

	bytewrite, errorwrite := io.Copy(file, resp.Body)
	if errorwrite != nil {
		fmt.Printf("Erreur lors de l'écriture du fichier : %v", errorwrite)
	}

	done <- bytewrite

}

func GetArchitecture() string {
	system, err := host.Info()
	if err != nil {
		fmt.Println("Erreur lors de la récupération des informations système : %v", err)
	}
	return system.KernelArch
}

func PrintProgress(channel chan int64, totalsize int, filepath string) {
	var stop bool = false

	for {
		select {
		case <-channel:
			stop = true
		default:
			file, err := os.Open(filepath)
			if err != nil {
				fmt.Printf("Erreur lors de l'ouverture du fichier: %v", err)
			}
			fi, errorstat := file.Stat()
			if errorstat != nil {
				fmt.Printf("Erreur lors de la récupération des statistiques du fichier: %v", err)
			}
			currentsize := fi.Size()

			if currentsize == 0 {
				currentsize = 1
			}

			var percent float64 = float64(currentsize) / float64(totalsize) * 100

			fmt.Printf("%.0f", percent)
			fmt.Println("%")

		}
		if stop {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {

	// Chargement de la configuration
	var conf Configuration = LoadSettings(inipath)

	// Récupération des arguments
	args := os.Args

	if len(args) < 2 {
		fmt.Println("Erreur argument manquant")
		os.Exit(1)
	}
	// Si on met à jour le cache local
	if args[1] == "update" {
		// Si l'arch est en 64bit
		if GetArchitecture() == "x86_64" {
			for _, collection := range conf.collections {
				fmt.Println("Téléchargement du manifeste de la collection " + collection + " :")
				//fmt.Println(conf.baseURL + "dists/" + conf.version + "/" + collection + "/" + "binary-amd64/Packages.gz")
				DownloadFile(conf.baseURL+"dists/"+conf.version+"/"+collection+"/"+"binary-amd64/Packages.gz", "")
			}
		}
	}

	//DownloadFile("http://ftp.debian.org/debian/dists/buster/main/binary-amd64/Packages.gz", "")

}
