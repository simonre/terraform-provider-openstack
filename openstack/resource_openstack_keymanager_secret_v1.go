package openstack

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/keymanager/v1/secrets"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceKeymanagerSecretV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeymanagerSecretV1Create,
		Read:   resourceKeymanagerSecretV1Read,
		Update: resourceKeymanagerSecretV1Update,
		Delete: resourceKeymanagerSecretV1Delete,

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
			"bit_length": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"algorithm": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"creator_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
			"secret_ref": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"payload": {
				Type:     schema.TypeString,
				Required: true,
			},
			"payload_content_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"payload_content_encoding": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceKeymanagerSecretV1Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keymanagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack keymanager client: %s", err)
	}

	var createOpts secrets.CreateOptsBuilder

	secretType := resourceKeymanagerSecretV1SecretType(d.Get("secret_type").(string))

	createOpts = &secrets.CreateOpts{
		Name:                   d.Get("name").(string),
		Algorithm:              d.Get("algorithm").(string),
		BitLength:              d.Get("bit_length").(int),
		Mode:                   d.Get("mode").(string),
		Payload:                d.Get("payload").(string),
		PayloadContentType:     d.Get("payload_content_type").(string),
		PayloadContentEncoding: d.Get("payload_content_encoding").(string),
		SecretType:             secretType,
	}

	log.Printf("[DEBUG] Create Options for resource_keymanager_secret_v1: %#v", createOpts)

	var secret *secrets.Secret
	secret, err = secrets.Create(kmClient, createOpts).Extract()

	uuid := getUUIDfromSecretRef(secret.SecretRef)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"NOT_CREATED"},
		Target:     []string{"ACTIVE"},
		Refresh:    waitForSecretCreation(kmClient, uuid),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican secret: %s", err)
	}

	d.SetId(uuid)

	return resourceKeymanagerSecretV1Read(d, meta)
}

func resourceKeymanagerSecretV1Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keymanagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican client: %s", err)
	}

	d.Id()
	secret, err := secrets.Get(kmClient, d.Id()).Extract()
	if err != nil {
		return CheckDeleted(d, err, "secret")
	}

	log.Printf("[DEBUG] Retrieved secret %s: %+v", d.Id(), secret)

	d.Set("name", secret.Name)

	d.Set("bit_length", secret.BitLength)
	d.Set("algorithm", secret.Algorithm)
	d.Set("creator_id", secret.CreatorID)
	d.Set("mode", secret.Mode)
	d.Set("secret_ref", secret.SecretRef)
	d.Set("secret_type", secret.SecretType)
	d.Set("status", secret.Status)
	d.Set("created", secret.Created)
	d.Set("updated", secret.Updated)
	d.Set("expiration", secret.Expiration)
	d.Set("content_types", secret.ContentTypes)
	// Set the region
	d.Set("region", GetRegion(d, config))

	return nil
}

func resourceKeymanagerSecretV1Update(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keymanagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican client: %s", err)
	}

	var hasChange = false
	var updateOpts secrets.UpdateOpts
	if d.HasChange("payload_content_type") {
		hasChange = true
	}
	// This is not optional so we have to set it regardless
	updateOpts.ContentType = d.Get("payload_content_type").(string)

	if d.HasChange("payload_content_encoding") {
		hasChange = true
		updateOpts.ContentEncoding = d.Get("content_encoding").(string)
	}
	if d.HasChange("payload") {
		hasChange = true
		updateOpts.Payload = d.Get("payload").(string)
	}

	if hasChange {
		err := secrets.Update(kmClient, d.Id(), updateOpts).Err
		if err != nil {
			return err
		}
	}

	return resourceKeymanagerSecretV1Read(d, meta)
}

func resourceKeymanagerSecretV1Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	kmClient, err := config.keymanagerV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("Error creating OpenStack barbican client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    waitForSecretDeletion(kmClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      0,
		MinTimeout: 2 * time.Second,
	}

	if _, err = stateConf.WaitForState(); err != nil {
		return err
	}

	return nil
}

func waitForSecretDeletion(kmClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := secrets.Delete(kmClient, id).Err
		if err == nil {
			return "", "DELETED", nil
		}

		return nil, "ACTIVE", err
	}
}

func resourceKeymanagerSecretV1SecretType(v string) secrets.SecretType {
	var stype secrets.SecretType
	switch v {
	case "symmetric":
		stype = secrets.SymmetricSecret
	case "public":
		stype = secrets.PublicSecret
	case "private":
		stype = secrets.PrivateSecret
	case "passphrase":
		stype = secrets.PassphraseSecret
	case "certificate":
		stype = secrets.CertificateSecret
	case "opaque":
		stype = secrets.OpaqueSecret
	}

	return stype
}

func waitForSecretCreation(kmClient *gophercloud.ServiceClient, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fmt.Println("ID is %v", id)
		secret, err := secrets.Get(kmClient, id).Extract()
		if err != nil {
			return "", "NOT_CREATED", nil
		}
		return secret, "ACTIVE", nil
	}
}

func getUUIDfromSecretRef(ref string) string {
	// secret ref has form https://{barbican_host}/v1/secrets/{secret_uuid}
	// so we are only interested in the last part
	ref_split := strings.Split(ref, "/")
	uuid := ref_split[len(ref_split)-1]
	return uuid
}
