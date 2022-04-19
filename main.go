package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
)

const (
	instanceIDURL = `http://169.254.169.254/opc/v2/instance/id`
)

func getInstanceId() (instanceID string, err error) {
	// get instance ID from local metadata

	client := &http.Client{}
	req, err := http.NewRequest("GET", instanceIDURL, nil)
	helpers.FatalIfError(err)

	req.Header.Set("Authorization", "Bearer Oracle")
	resp, err := client.Do(req)
	helpers.FatalIfError(err)

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)
	return bodyString, nil

}

func reboot(client core.ComputeClient, instanceId string) {
	// make API call to reboot instance

	reboot_request := core.InstanceActionRequest{InstanceId: &instanceId, Action: core.InstanceActionActionSoftreset}
	r, err := client.InstanceAction(context.Background(), reboot_request)
	client.InstanceAction(context.Background(), reboot_request)
	helpers.FatalIfError(err)

	// initiate syslog and log reboot message
	syslog, err := syslog.New(syslog.LOG_INFO, "oci-reboot")
	helpers.FatalIfError(err)
	log.SetOutput(os.Stdout)
	log.Printf("Reboot in progress ( API %s ) | Instance: %s | User: %d", r.RawResponse.Status, instanceId, os.Getuid())
	log.SetOutput(syslog)
	log.Printf("Reboot in progress ( API %s ) | Instance: %s | User: %d", r.RawResponse.Status, instanceId, os.Getuid())
	os.Exit(0)
}

func main() {
	// parse flags
	var force bool
	flag.BoolVar(&force, "f", false, "force reboot")
	flag.Parse()

	provider, err := auth.InstancePrincipalConfigurationProvider()
	helpers.FatalIfError(err)

	client, err := core.NewComputeClientWithConfigurationProvider(provider)
	helpers.FatalIfError(err)

	reboot_due_date := ""

	instanceId, err := getInstanceId()
	helpers.FatalIfError(err)

	instance_request := core.GetInstanceRequest{InstanceId: &instanceId}
	instance_metadata, err := client.GetInstance(context.Background(), instance_request)

	if err != nil {
		// if failed to retrieve the metadata (this might be the case if we ONLY have permissions to reboot)
		fmt.Println(err)
		reboot_due_date = time.Now().String()
	} else {
		if instance_metadata.TimeMaintenanceRebootDue != nil {
			// convert to string if found
			reboot_due_date = instance_metadata.TimeMaintenanceRebootDue.String()
		}
	}

	// if reboot date is not set do nothing unless force is set
	if (reboot_due_date != "") || (force) {
		reboot(client, instanceId)
	} else {
		fmt.Println("Not maintanance date set. Skipping reboot.")
		os.Exit(2)
	}

}
