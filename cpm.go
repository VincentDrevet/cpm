package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/shirou/gopsutil/host"
	"gopkg.in/ini.v1"
)

const inipath = "cpm.conf"

// Configuration correspondant au information chargé depuis le fichier de configuration
type Configuration struct {
	baseURL    string
	version    string
	collection []string
	localCache string
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
	loadConfiguration.collection = parseCollections

	loadConfiguration.localCache = config_file.Section("Main").Key("local_cache").String()
	return loadConfiguration

}

func DownloadFile(url string, destdir string) error {

	var parseURL []string = strings.Split(url, "/")

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erreur lors du téléchargement du fichier : %v", err)
	}
	defer resp.Body.Close()

	file, errorfile := os.Create(destdir + parseURL[len(parseURL)-1])
	if errorfile != nil {
		fmt.Printf("Erreur lors de la création du fichier : %v", errorfile)
	}
	defer file.Close()

	_, errorwrite := io.Copy(file, resp.Body)

	return errorwrite
}

func GetArchitecture() string {
	system, err := host.Info()
	if err != nil {
		fmt.Println("Erreur lors de la récupération des informations système : %v", err)
	}
	return system.KernelArch
}

func main() {

	// Chargement de la configuration
	var conf Configuration = LoadSettings(inipath)

	fmt.Println(conf)

	// Récupération des arguments
	args := os.Args

	if args[1] == "update" {
		fmt.Println(GetArchitecture())
	}

	//DownloadFile("http://ftp.debian.org/debian/dists/buster/main/binary-amd64/Packages.gz", "")

}
