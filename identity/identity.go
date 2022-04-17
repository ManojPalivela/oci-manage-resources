package identity

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v61/common"
	"github.com/oracle/oci-go-sdk/v61/identity"
)

// retry stratergy
// wait for max 5 miniutes
// retries: 15s, 30s, 60s, 120s, 240s
func SetGlobalRetryStratergy() {

	maximumCumulativeBackoff := time.Duration(5) * time.Minute

	customRetryPolicy := common.NewRetryPolicyWithOptions(
		common.WithUnlimitedAttempts(maximumCumulativeBackoff),
		common.WithShouldRetryOperation(func(r common.OCIOperationResponse) bool {
			durationSinceInitialAttempt := time.Now().Sub(r.InitialAttemptTime)
			tooLong := durationSinceInitialAttempt > maximumCumulativeBackoff
			return common.DefaultShouldRetryOperation(r) && !tooLong
		}),
		common.WithNextDuration(func(r common.OCIOperationResponse) time.Duration {
			return time.Duration(math.Pow(float64(2), float64(r.AttemptNumber-1))) * time.Second * 15
		}),
	)

	common.GlobalRetry = &customRetryPolicy
}

func GetConfigProvider(configFilePath string, profile string, privatekeyPassphrase string) (common.ConfigurationProvider, error) {
	//TODO
	//Check for existance of configFilePath

	config, err := common.ConfigurationProviderFromFileWithProfile(configFilePath, profile, privatekeyPassphrase)
	return config, err
}

func GetClient(config common.ConfigurationProvider, region string) (identity.IdentityClient, error) {

	client, err := identity.NewIdentityClientWithConfigurationProvider(config)

	if strings.TrimSpace(region) != "" {
		client.SetRegion(region)
	}
	return client, err

}

func GetSubscribedRegions(config common.ConfigurationProvider) []string {
	var subscribedRegions []string
	client, err := identity.NewIdentityClientWithConfigurationProvider(config)

	if err != nil {
		return subscribedRegions
	}
	tenancyocid, _ := config.TenancyOCID()
	req := identity.ListRegionSubscriptionsRequest{TenancyId: &tenancyocid}

	resp, err := client.ListRegionSubscriptions(context.Background(), req)
	if err != nil {
		return subscribedRegions
	}

	for _, rs := range resp.Items {
		subscribedRegions = append(subscribedRegions, *rs.RegionName)
	}
	return subscribedRegions
}

// array of compartment objects
// compartment attributes: compartment name, compartment ocid

type ActiveCompartments struct {
	Compartments []ActiveCompartment
}
type ActiveCompartment struct {
	CompartmentName string
	CompartmentOCID string
}

func GetActiveCompartments(config common.ConfigurationProvider) ActiveCompartments {

	var compartments ActiveCompartments
	var compartment ActiveCompartment

	client, err := identity.NewIdentityClientWithConfigurationProvider(config)
	if err != nil {
		return compartments
	}
	tenancyocid, _ := config.TenancyOCID()

	req := identity.ListCompartmentsRequest{AccessLevel: identity.ListCompartmentsAccessLevelAny,
		CompartmentIdInSubtree: common.Bool(true),
		LifecycleState:         identity.CompartmentLifecycleStateActive,
		CompartmentId:          &tenancyocid}

	resp, err := client.ListCompartments(context.Background(), req)

	for resp.OpcNextPage != nil {
		req.Page = resp.OpcNextPage
		resp1, _ := client.ListCompartments(context.Background(), req)
		resp.Items = append(resp.Items, resp1.Items...)
		fmt.Println("Processing next page")
	}

	for _, c := range resp.Items {
		compartment.CompartmentOCID = *c.Id
		compartment.CompartmentName = *c.Name
		compartments.Compartments = append(compartments.Compartments, compartment)
	}

	return compartments
}
