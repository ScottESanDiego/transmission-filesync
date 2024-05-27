package main

import (
	"os"
	"fmt"
	"flag"
	"context"
	"sort"
	"slices"
	"net/url"
	"github.com/hekmon/transmissionrpc/v3"
)

func main() {
	// Variables that'll be command line arguments
	var hostname string	// IP or FQDN of the Transmission server
	var username string	// Username for Transmission-RPC
	var password string	// Password for Transmission-RPC
	var filedir string	// Directory in the filesystem where the torrents live

	flag.StringVar(&hostname, "hostname", "127.0.0.1", "IP or FQDN of the Transmission server")
	flag.StringVar(&username, "username", "", "Username for Transmission-RPC")
	flag.StringVar(&password, "password", "", "Password for Transmission-RPC")
	flag.StringVar(&filedir, "directory", "/var/torrents/", "Directory in the filesystem where the torrents live with trailing slash")
	dryrun := flag.Bool("dryrun", false, "Don't actually delete files, just print what would happen")
	flag.Parse()

	// Create a new Transmission RPC client
	endpoint, err := url.Parse("http://"+username+":"+password+"@"+hostname+":9091/transmission/rpc")
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := transmissionrpc.New(endpoint,nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the list of torrents
	torrents, err := client.TorrentGetAll(context.TODO())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Build a slice of torrent names from the list of torrents
	var torrentnames  []string
	for _, torrent := range torrents {
		torrentnames = append(torrentnames, *torrent.Name)
	}

	// Sort slice so that we can search it later
	sort.Strings(torrentnames)

	// Iterate over all files in a directory
	files, err := os.ReadDir(filedir)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		_, found := slices.BinarySearch(torrentnames, file.Name())
		if !found {
			// Super-duper extra check that file.Name() isn't null since os.RemoveAll will remove filedir otherwise!
			if len(file.Name()) > 0 {
				if *dryrun == true {
					fmt.Printf("Dry-run, not deleting: ")
				} else {
					fmt.Printf("Deleting unowned file: ")
					err := os.RemoveAll(filedir+file.Name())
					if err != nil {
						fmt.Println(err)
						return
					}
			}
			} else {
				fmt.Println("ERROR: Filename was empty!")
			}

			fmt.Printf("%s\n", filedir+file.Name())
		}
	}
}

