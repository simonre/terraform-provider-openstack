package openstack

import (
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud/openstack/keymanager/v1/containers"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceKeyManagerContainerV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeyManagerContainerV1Create,
		Read:   resourceKeyManagerContainerV1Read,
		Update: resourceKeyManagerContainerV1Update,
		Delete: resourceKeyManagerContainerV1Delete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"container_ref": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"creator_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"rsa", "generic", "certificate",
				}, true),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consumers": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"secret_refs": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret_ref": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceKeyManagerContainerV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keyManagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack keymanager client: %s", err)
	}

	var createOpts containers.CreateOptsBuilder

	containertype := keyManagerContainerV1ContainerType(d.Get("type").(string))
	secretRefs := keyManagerContainerV1SecretRefs(d.Get("secret_refs").([]interface{}))

	createOpts = &containers.CreateOpts{
		Name:       d.Get("name").(string),
		Type:       containertype,
		SecretRefs: secretRefs,
	}

	log.Printf("[DEBUG] Create Options for resource_keymanager_container_v1: %#v", createOpts)

	var container *containers.Container
	container, err = containers.Create(kmClient, createOpts).Extract()

	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican container: %s", err)
	}

	uuid := keyManagerContainerV1GetUUIDfromContainerRef(container.ContainerRef)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"NOT_CREATED"},
		Target:     []string{"ACTIVE"},
		Refresh:    keyManagerContainerV1WaitForContainerCreation(kmClient, uuid),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return CheckDeleted(d, err, "Error creating openstack_keymanager_container_v1")
	}

	d.SetId(uuid)

	return resourceKeyManagerContainerV1Read(d, meta)
}

func resourceKeyManagerContainerV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keyManagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican client: %s", err)
	}

	container, err := containers.Get(kmClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "container")
	}

	log.Printf("[DEBUG] Retrieved openstack_keymanager_container_v1 with id %s: %+v", d.Id(), container)

	d.Set("name", container.Name)

	d.Set("creator_id", container.CreatorID)
	d.Set("type", container.Type)
	d.Set("consumers", container.Consumers)
	d.Set("created", container.Created.Format(time.RFC3339))
	d.Set("updated", container.Updated.Format(time.RFC3339))
	d.Set("status", container.Status)
	d.Set("container_ref", container.ContainerRef)
	d.Set("secret_refs", container.SecretRefs)

	// Set the region
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceKeyManagerContainerV1Update(d *schema.ResourceData, meta interface{}) error {
	// Cannot be updated
	return resourceKeyManagerContainerV1Read(d, meta)
}

func resourceKeyManagerContainerV1Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keyManagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    keyManagerContainerV1WaitForContainerDeletion(kmClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return err
	}

	return nil
}
