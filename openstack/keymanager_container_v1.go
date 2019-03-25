package openstack

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/keymanager/v1/containers"
	"github.com/hashicorp/terraform/helper/resource"
	"strings"
)

func keyManagerContainerV1ContainerType(v string) containers.ContainerType {
	var containertype containers.ContainerType
	switch v {
	case "rsa":
		containertype = containers.RSAContainer
	case "generic":
		containertype = containers.GenericContainer
	case "certificate":
		containertype = containers.CertificateContainer
	}

	return containertype
}

func keyManagerContainerV1SecretRefs(v []interface{}) []containers.SecretRef {
	secretRefs := make([]containers.SecretRef, len(v))
	for i, item := range v {
		var secretRef containers.SecretRef
		secretRef.Name = (item.(map[string]interface{}))["name"].(string)
		secretRef.SecretRef = (item.(map[string]interface{}))["secret_ref"].(string)
		secretRefs[i] = secretRef
	}
	return secretRefs
}

func keyManagerContainerV1GetUUIDfromContainerRef(ref string) string {
	// container ref has form https://{barbican_host}/v1/containers/{container_uuid}
	// so we are only interested in the last part
	ref_split := strings.Split(ref, "/")
	uuid := ref_split[len(ref_split)-1]
	return uuid
}

func keyManagerContainerV1WaitForContainerCreation(kmClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		container, err := containers.Get(kmClient, id).Extract()
		if err != nil {
			return "", "NOT_CREATED", nil
		}
		return container, "ACTIVE", nil
	}
}

func keyManagerContainerV1WaitForContainerDeletion(kmClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := containers.Delete(kmClient, id).Err
		if err == nil {
			return "", "DELETED", nil
		}

		if _, ok := err.(gophercloud.ErrDefault404); ok {
			return "", "DELETED", nil
		}

		return nil, "ACTIVE", err
	}
}

func flattenKeyManagerContainerV1ConsumerCreateOpts(v []interface{}) []containers.CreateConsumerOpts {
	consumers := make([]containers.CreateConsumerOpts, len(v))
	for i, item := range v {
		var consumer containers.CreateConsumerOpts
		consumer.Name = (item.(map[string]interface{}))["name"].(string)
		consumer.URL = (item.(map[string]interface{}))["url"].(string)
		consumers[i] = consumer
	}
	return consumers
}
