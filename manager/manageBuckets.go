package main

import (
	"bufio"
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
	objectStorage "github.com/ManojPalivela/oci-manage-resources/storage"
	utls "github.com/ManojPalivela/oci-manage-resources/utils"
	"github.com/oracle/oci-go-sdk/v61/common"
)

var logFile string
var session = utls.GetTimeStamp()
var mutex sync.Mutex
var mutext1 sync.Mutex

func main() {
	var mu sync.Mutex
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

	WriteLog(logFile, &mutex, "Setting global retry stratergy")
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
		WriteLog(logFile, &mutex, "Error occured while try to get config provider")
		WriteLog(logFile, &mutex, err.Error())
	}

	bucketList := objectStorage.GetBucketsIncompartment(config, "")

	buckets := bucketList.Buckets

	for i := 0; i < len(buckets); i++ {
		fmt.Println("================")
		fmt.Println("bucket name : " + buckets[i].BucketName)
		fmt.Printf("bucket details : %v\n", buckets[i])
		fmt.Println("================")
	}

	os.Exit(0)
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
		fmt.Println(ac.CompartmentOCID)
		compartmentOCIDS = append(compartmentOCIDS, ac.CompartmentOCID)
	}
	tencyocid, _ := config.TenancyOCID()
	compartmentOCIDS = append(compartmentOCIDS, tencyocid)
	fmt.Printf("compartmentOCIDS: %v\n", compartmentOCIDS)
	//create tenancy dir
	os.MkdirAll(filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, ociprofile), 0750)
	// process each region
	excludedRegions := getRegionExclusionList(ociprofile)
	wg := sync.WaitGroup{}
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
			WriteLog(logFile, &mutex, "Processing region "+v+" in "+ociprofile)
			go GetBucketsInRegion(&wg, &mu, config, ociprofile, v, compartmentOCIDS)
		}
	}
	WriteLog(logFile, &mutex, "Waiting for region processing")
	wg.Wait()
	WriteLog(logFile, &mutex, "completed region processing")
}

func GetBucketsInRegion(wg *sync.WaitGroup, mu *sync.Mutex, config common.ConfigurationProvider, tenancy string, region string, compartmentOCIDS []string) {
	mu.Lock()
	var mu1 sync.Mutex

	logFile1 := filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, tenancy, region+".log")

	WriteLog(logFile1, &mu1, "Starting to process region "+region+" in "+tenancy)

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
			WriteLog(logFile1, &mu1, "Processing compartment "+v+" in "+region+" in "+tenancy)
			go GetBucketsInCompartmentInRegion(wg1, &mu1, config, tenancy, region, v)
		} else {
			WriteLog(logFile1, &mu1, "Skipping processing of compartment "+v+" in "+region+" in "+tenancy)
		}
	}
	mu.Unlock()
	WriteLog(logFile1, &mu1, "waiting for compartment procesing : "+region+" in "+tenancy)
	wg1.Wait()
	WriteLog(logFile1, &mu1, "completed compartment procesing : "+region+" in "+tenancy)
	WriteLog(logFile1, &mu1, "completed processing region "+region+" in "+tenancy)
	wg.Done()
}

func GetBucketsInCompartmentInRegion(wg *sync.WaitGroup, mu1 *sync.Mutex, config common.ConfigurationProvider, tenancy string, region string, compartment string) {
	mu1.Lock()
	logFile1 := filepath.Join(os.Getenv("WORKING_DIR"), "logs", session, tenancy, region+".log")
	mu1.Unlock()

	WriteLog(logFile1, mu1, "Starting to process compartment "+compartment+" in "+region+" in "+tenancy)

	//business logic
	time.Sleep(5 * time.Second)

	WriteLog(logFile1, mu1, "completed processing compartment "+compartment+" in "+region+" in "+tenancy)

	wg.Done()

}

func WriteLog(logFile1 string, mu2 *sync.Mutex, message string) {
	mu2.Lock()

	file, err := os.OpenFile(logFile1, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(file)
	ts := utls.GetTimeStamp()

	_, err = fmt.Fprintf(w, "%s : %s\n", ts, message)
	if err != nil {
		panic(err)
	}
	w.Flush()
	file.Close()

	mu2.Unlock()
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
			innermaps, _ := compartmentExclusionList[k].(map[string]interface{})
			fmt.Printf("innermaps: %v\n", innermaps)
			for k1, v1 := range innermaps {

				if strings.ToLower(k1) == "global" || k1 == region {
					for _, v2 := range v1.([]interface{}) {
						compartmentList = append(compartmentList, v2.(string))
					}
				}
			}
		}
	}
	fmt.Printf("compartmentList: %v : %s\n", compartmentList, region)
	return compartmentList
}
