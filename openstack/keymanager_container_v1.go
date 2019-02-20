package openstack

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/keymanager/v1/containers"
	"github.com/hashicorp/terraform/helper/resource"
	"strings"
)

func keymanagerContainerV1ContainerType(v string) containers.ContainerType {
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

func keymanagerContainerV1SecretRefs(v []interface{}) []containers.SecretRef {
	secretRefs := make([]containers.SecretRef, len(v))
	for i, item := range v {
		var secretRef containers.SecretRef
		secretRef.Name = (item.(map[string]interface{}))["name"].(string)
		secretRef.SecretRef = (item.(map[string]interface{}))["secret_ref"].(string)
		secretRefs[i] = secretRef
	}
	return secretRefs
}

func keymanagerContainerV1GetUUIDfromContainerRef(ref string) string {
	// container ref has form https://{barbican_host}/v1/containers/{container_uuid}
	// so we are only interested in the last part
	ref_split := strings.Split(ref, "/")
	uuid := ref_split[len(ref_split)-1]
	return uuid
}

func keymanagerContainerV1WaitForContainerCreation(kmClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fmt.Println("[DEBUG] Waiting for openstack_keymanager_container_v1 with ID %v to be created", id)
		container, err := containers.Get(kmClient, id).Extract()
		if err != nil {
			return "", "NOT_CREATED", nil
		}
		return container, "ACTIVE", nil
	}
}

func keymanagerContainerV1WaitForContainerDeletion(kmClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
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
