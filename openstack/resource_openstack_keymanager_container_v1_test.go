package openstack

import (
	"fmt"
	"os"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/keymanager/v1/containers"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccContainerV1_basic(t *testing.T) {
	var container containers.Container
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckKeyManager(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerV1_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerV1Exists(
						"openstack_keymanager_container_v1.container_1", &container),
					resource.TestCheckResourceAttrPtr("openstack_keymanager_container_v1.container_1", "name", &container.Name),
					resource.TestCheckResourceAttrPtr("openstack_keymanager_container_v1.container_1", "type", &container.Type),
				),
			},
		},
	})
}

func TestAccContainerV1_withConsumers(t *testing.T) {
	os.Setenv("OS_PROJECT_NAME", "sreinkemeier")
	os.Setenv("OS_DEBUG", "1")
	os.Setenv("TF_LOG", "DEBUG")
	var container containers.Container
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckKeyManager(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerV1_withConsumers,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerV1Exists(
						"openstack_keymanager_container_v1.container_1", &container),
					resource.TestCheckResourceAttrPtr("openstack_keymanager_container_v1.container_1", "name", &container.Name),
					resource.TestCheckResourceAttrPtr("openstack_keymanager_container_v1.container_1", "type", &container.Type),
				),
			},
		},
	})
}

func testAccCheckContainerV1Destroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	kmClient, err := config.keyManagerV1Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack Keymanager client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_keymanager_container" {
			continue
		}
		_, err = containers.Get(kmClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Container (%s) still exists.", rs.Primary.ID)
		}
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return err
		}
	}
	return nil
}

func testAccCheckContainerV1Exists(n string, container *containers.Container) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		kmClient, err := config.keyManagerV1Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack keymanager client: %s", err)
		}

		var found *containers.Container

		found, err = containers.Get(kmClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}
		*container = *found

		return nil
	}
}

const testAccContainerV1_basic = `
resource "openstack_keymanager_container_v1" "container_1" {
  type = "generic"
  name = "Test Container"
  secret_refs = [
    {
      "secret_ref" =  "${openstack_keymanager_secret_v1.secret_1.secret_ref}",
      "name" = "a secret"
    }
  ]
}

resource "openstack_keymanager_secret_v1" "secret_1" {
  algorithm = "aes"
  bit_length = 256
  mode = "cbc"
  name = "mysecret"
  secret_type = "passphrase"
  payload = ""
}`

const testAccContainerV1_withConsumers = `
resource "openstack_keymanager_container_v1" "container_1" {
  type = "generic"
  name = "Test Container"
  consumers = [
    {
      name = "ConsumerName"
      url = "ConsumerURL"
    }
  ]
  secret_refs = [
    {
      "secret_ref" =  "${openstack_keymanager_secret_v1.secret_1.secret_ref}",
      "name" = "a secret"
    }
  ]
}

resource "openstack_keymanager_secret_v1" "secret_1" {
  algorithm = "aes"
  bit_length = 256
  mode = "cbc"
  name = "mysecret"
  secret_type = "passphrase"
  payload = ""
}`
