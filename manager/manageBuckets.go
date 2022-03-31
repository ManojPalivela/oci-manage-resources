package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	identity "github.com/ManojPalivela/oci-manage-resources/identity"
	utls "github.com/ManojPalivela/oci-manage-resources/utils"
	"github.com/oracle/oci-go-sdk/v61/common"
)

var logFile string
var session = utls.GetTimeStamp()

func main() {

	_, exists := os.LookupEnv("WORKING_DIR")
	if !exists {
		fmt.Println("WORKING_DIR environment variable not set")
		os.Exit(-1)
	}
	ociprofile, exists := os.LookupEnv("OCI_PROFILE")

	if !exists {
		fmt.Println("OCI_PROFILE environment variable not set")
		os.Exit(-1)
	}

	os.MkdirAll(filepath.Join(os.Getenv("WORKING_DIR"), "logs", session), 0750)
	os.MkdirAll(filepath.Join(os.Getenv("WORKING_DIR"), "results", session), 0750)
	logFile = filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, "manageBuckets.log")

	utls.WriteLog(logFile, "Setting global retry stratergy")
	identity.SetGlobalRetryStratergy()
	//create file  WORKING_DIR/tenancyAdmins with OCI user data
	/*
		[<tenancy name>]
		user=<user ocid>
		key_file=<absolute path to private key file ( API signing key)>
		tenancyName=<tenancy name>
		tenancy=<tenancy ocid>
		region=<base region>
		fingerprint=<API signing key fingerprint>
	*/
	config, err := identity.GetConfigProvider(filepath.Join(os.Getenv("WORKING_DIR"), "tenancyAdmins"), ociprofile, "")
	if err != nil {
		fmt.Println("Error occured while try to get config provider")
		utls.WriteLog(logFile, "Error occured while try to get config provider")
		utls.WriteLog(logFile, err.Error())
	}
	// get all subscribed regions
	// get all compartments which are active

	subscribedregions := identity.GetSubscribedRegions(config)

	for _, v := range subscribedregions {
		fmt.Println(v)
	}
	activeCompartments := identity.GetActiveCompartments(config)
	fmt.Println("number of compartments : " + strconv.Itoa(len(activeCompartments.Compartments)))
	var compartmentOCIDS []string
	for _, ac := range activeCompartments.Compartments {
		fmt.Println(ac.CompartmentName)
		compartmentOCIDS = append(compartmentOCIDS, ac.CompartmentOCID)
	}

	//create tenancy dir
	os.MkdirAll(filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, ociprofile), 0750)
	// process each region
	excludedRegions := getRegionExclusionList(ociprofile)
	wg := new(sync.WaitGroup)
	for _, v := range subscribedregions {
		fmt.Println(v)
		regionToBeExcluded := false
		for _, v1 := range excludedRegions {
			if v == v1 {
				regionToBeExcluded = true
			}
		}
		if !regionToBeExcluded {
			wg.Add(1)
			utls.WriteLog(logFile, "Processing region "+v+" in "+ociprofile)
			GetBucketsInRegion(wg, config, ociprofile, v, compartmentOCIDS)
		}
	}
	utls.WriteLog(logFile, "Waiting for region processing")
	wg.Wait()
	utls.WriteLog(logFile, "completed region processing")
}

func GetBucketsInRegion(wg *sync.WaitGroup, config common.ConfigurationProvider, tenancy string, region string, compartmentOCIDS []string) {
	defer wg.Done()
	logFile := filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, tenancy, region+".log")
	utls.WriteLog(logFile, "Starting to process region "+region+" in "+tenancy)
	//read compartment exclusion list
	compartmentExclusionList := getCompartmentExclusionList(tenancy, region)
	wg1 := new(sync.WaitGroup)
	for _, v := range compartmentOCIDS {
		excludeCompartment := false
		for _, v1 := range compartmentExclusionList {
			if v == v1 {
				excludeCompartment = true
				break
			}
		}
		if !excludeCompartment {
			wg1.Add(1)
			utls.WriteLog(logFile, "Processing compartment "+v+" in "+region+" in "+tenancy)
			GetBucketsInCompartmentInRegion(wg1, config, tenancy, region, v)
		}
	}
	utls.WriteLog(logFile, "waiting for compartment procesing : "+region+" in "+tenancy)
	wg1.Wait()
	utls.WriteLog(logFile, "completed compartment procesing : "+region+" in "+tenancy)
}

func GetBucketsInCompartmentInRegion(wg *sync.WaitGroup, config common.ConfigurationProvider, tenancy string, region string, compartment string) {
	defer wg.Done()
	logFile := filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, tenancy, region+"_"+compartment+".log")
	utls.WriteLog(logFile, "Starting to process compartment "+compartment+" in "+region+" in "+tenancy)
	time.Sleep(5 * time.Second)
	utls.WriteLog(logFile, "completed processing compartment "+compartment+" in "+region+" in "+tenancy)
}

func getRegionExclusionList(tenancy string) []string {
	var regionexclusonlist map[string][]string
	var regionlist []string

	workingDir, _ := os.LookupEnv("WORKING_DIR")

	jsonFile, err := os.Open(filepath.Join(workingDir, "config", "regionExclusionList.json"))

	if err != nil {
		fmt.Println("not able to open file")
		return regionlist
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &regionexclusonlist)
	fmt.Printf("regionexclusonlist: %v\n", regionexclusonlist)
	for k, _ := range regionexclusonlist {
		if k == tenancy {
			for _, v := range regionexclusonlist[k] {
				regionlist = append(regionlist, v)
			}
		}
	}
	return regionlist
}

func getCompartmentExclusionList(tenancy string, region string) []string {
	var compartmentList []string
	var compartmentExclusionList map[string]interface{}
	workingDir, _ := os.LookupEnv("WORKING_DIR")

	jsonFile, err := os.Open(filepath.Join(workingDir, "config", "compartmentExclusionList.json"))

	if err != nil {
		fmt.Println("not able to open file")
		return compartmentList
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &compartmentExclusionList)

	for k, _ := range compartmentExclusionList {
		if k == tenancy {
			innermaps, _ := compartmentExclusionList[k].(map[string][]string)

			for k1, v1 := range innermaps {
				if strings.ToLower(k1) == "global" || k1 == region {
					for _, v2 := range v1 {
						compartmentList = append(compartmentList, v2)
					}
				}
			}
		}
	}

	return compartmentList
}
