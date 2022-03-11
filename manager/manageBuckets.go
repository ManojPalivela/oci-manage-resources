package main

import (
	"fmt"
	"os"
	"path/filepath"

	identity "github.com/ManojPalivela/oci-manage-resources/identity"
	utls "github.com/ManojPalivela/oci-manage-resources/utils"
)

func main() {
	var logFile string

	_, exists := os.LookupEnv("WORKING_DIR")
	if !exists {
		fmt.Println("WORKING_DIR environment variable not set")
		os.Exit(-1)
	}
	session := utls.GetTimeStamp()

	os.MkdirAll(filepath.Join(os.Getenv("WORKING_DIR"), "logs", session), 0750)
	os.MkdirAll(filepath.Join(os.Getenv("WORKING_DIR"), "results", session), 0750)
	logFile = filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, "manageBuckets.log")

	utls.WriteLog(logFile, "Setting global retry stratergy")
	identity.SetGlobalRetryStratergy()

	config, err := identity.GetConfigProvider("/Users/mpalivel/Documents/tenancyAdmins", "DBAASPORT", "")
	if err != nil {
		fmt.Println("Error occured while try to get config provider")
		utls.WriteLog(logFile, "Error occured while try to get config provider")
		utls.WriteLog(logFile, err.Error())
	}
	// get all subscribed regions
	// get all compartments which are active
	// exclude any user specified regions ( per tenancy config )
	// exclude any user specified compartments ( can be specified at tenancy or region level )

	subscribedregions := identity.GetSubscribedRegions(config)

	for _, v := range subscribedregions {
		fmt.Println(v)
	}

}
