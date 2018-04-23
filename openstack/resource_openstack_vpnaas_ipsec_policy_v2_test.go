package openstack

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/vpnaas/ipsecpolicies"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccIPSecPolicyV2_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckVPN(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIPSecPolicyV2Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccIPSecPolicyV2_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSecPolicyV2Exists(
						"openstack_vpnaas_ipsec_policy_v2.policy_1", "", ""),
				),
			},
		},
	})
}

func TestAccIPSecPolicyV2_withLifetime(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckVPN(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIPSecPolicyV2Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccIPSecPolicyV2_withLifetime,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSecPolicyV2Exists(
						"openstack_vpnaas_ipsec_policy_v2.policy_1", "", ""),
				),
			},
		},
	})
}

func TestAccIPSecPolicyV2_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckVPN(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIPSecPolicyV2Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccIPSecPolicyV2_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSecPolicyV2Exists(
						"openstack_vpnaas_ipsec_policy_v2.policy_1", "", ""),
				),
			},
			resource.TestStep{
				Config: testAccIPSecPolicyV2_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSecPolicyV2Exists(
						"openstack_vpnaas_ipsec_policy_v2.policy_1", "updatedname", ""),
				),
			},
		},
	})
}

func TestAccIPSecPolicyV2_withLifetimeUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckVPN(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIPSecPolicyV2Destroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccIPSecPolicyV2_withLifetime,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSecPolicyV2Exists(
						"openstack_vpnaas_ipsec_policy_v2.policy_1", "", ""),
				),
			},
			resource.TestStep{
				Config: testAccIPSecPolicyV2_withLifetimeUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSecPolicyV2Exists(
						"openstack_vpnaas_ipsec_policy_v2.policy_1", "", ""),
				),
			},
		},
	})
}

func testAccCheckIPSecPolicyV2Destroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	networkingClient, err := config.networkingV2Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_vpnaas_ipsec_policy_v2" {
			continue
		}
		_, err = ipsecpolicies.Get(networkingClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("IPSec policy (%s) still exists.", rs.Primary.ID)
		}
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return err
		}
	}
	return nil
}

func testAccCheckIPSecPolicyV2Exists(n, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		networkingClient, err := config.networkingV2Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack networking client: %s", err)
		}

		found, err := ipsecpolicies.Get(networkingClient, rs.Primary.ID).Extract()
		switch {
		case name != found.Name:
			err = fmt.Errorf("Expected name <%s>, but found <%s>", name, found.Name)
		case description != found.Description:
			err = fmt.Errorf("Expected description <%s>, but found <%s>", description, found.Description)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

const testAccIPSecPolicyV2_basic = `
resource "openstack_vpnaas_ipsec_policy_v2" "policy_1" {
}
`

const testAccIPSecPolicyV2_Update = `
resource "openstack_vpnaas_ipsec_policy_v2" "policy_1" {
	name = "updatedname"
}
`

const testAccIPSecPolicyV2_withLifetime = `
resource "openstack_vpnaas_ipsec_policy_v2" "policy_1" {
	auth_algorithm = "sha256"
	pfs = "group14"
	lifetime {
		units = "seconds"
		value = 1200
	}
}
`

const testAccIPSecPolicyV2_withLifetimeUpdate = `
resource "openstack_vpnaas_ipsec_policy_v2" "policy_1" {
	auth_algorithm = "sha256"
	pfs = "group14"
	lifetime {
		units = "seconds"
		value = 1400
	}
}
`
