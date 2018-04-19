package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmMySqlServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmMySqlServerCreate,
		Read:   resourceArmMySqlServerRead,
		Update: resourceArmMySqlServerUpdate,
		Delete: resourceArmMySqlServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"sku": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"B_Gen4_1",
								"B_Gen4_2",
								"B_Gen5_1",
								"B_Gen5_2",
								"GP_Gen4_2",
								"GP_Gen4_4",
								"GP_Gen4_8",
								"GP_Gen4_16",
								"GP_Gen4_32",
								"GP_Gen5_2",
								"GP_Gen5_4",
								"GP_Gen5_8",
								"GP_Gen5_16",
								"GP_Gen5_32",
								"MO_Gen5_2",
								"MO_Gen5_4",
								"MO_Gen5_8",
								"MO_Gen5_16",
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},

						"capacity": {
							Type:     schema.TypeInt,
							Required: true,
							ValidateFunc: validateIntInSlice([]int{
								2,
								4,
								8,
								16,
								32,
							}),
						},

						"tier": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(mysql.Basic),
								string(mysql.GeneralPurpose),
								string(mysql.MemoryOptimized),
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},

						"family": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Gen4",
								"Gen5",
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
					},
				},
			},

			"administrator_login": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"administrator_login_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"version": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(mysql.FiveFullStopSix),
					string(mysql.FiveFullStopSeven),
				}, true),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
				ForceNew:         true,
			},

			"storage_profile": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_mb": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
							ValidateFunc: validateIntInSlice([]int{
								5120,
								179200,
								307200,
								435200,
								563200,
								691200,
								819200,
								947200,
								128000,
								256000,
								384000,
								512000,
								640000,
								768000,
								896000,
								1048576,
							}),
						},

						"backupRetentionDays": {
							Type:     schema.TypeInt,
							Required: false,
							ValidateFunc: validateIntInSlice([]int{
								7,
								8,
								9,
								10,
								11,
								12,
								13,
								14,
								15,
								16,
								17,
								18,
								19,
								20,
								21,
								22,
								23,
								24,
								25,
								26,
								27,
								28,
								29,
								30,
								31,
								32,
								33,
								34,
								35,
							}),
						},

						"georedundantbackup": {
							Type:     schema.TypeString,
							Required: false,
							ValidateFunc: validation.StringInSlice([]string{
								"Enabled",
								"Disabled",
							}, true),
							DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
						},
					},
				},
			},

			"ssl_enforcement": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(mysql.SslEnforcementEnumDisabled),
					string(mysql.SslEnforcementEnumEnabled),
				}, true),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"createmode": {
				Type:     schema.TypeString,
				Required: false,
				ValidateFunc: validation.StringInSlice([]string{
					"Default",
					"PointInTimeRestore",
				}, true),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmMySqlServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlServersClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM MySQL Server creation.")

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	adminLogin := d.Get("administrator_login").(string)
	adminLoginPassword := d.Get("administrator_login_password").(string)
	sslEnforcement := d.Get("ssl_enforcement").(string)
	version := d.Get("version").(string)
	createMode := d.Get("createmode").(string)
	tags := d.Get("tags").(map[string]interface{})

	sku := expandMySQLServerSku(d)
	storageprofile := expandMySQLStorageProfile(d)

	properties := mysql.ServerForCreate{
		Location: &location,
		Properties: &mysql.ServerPropertiesForDefaultCreate{
			AdministratorLogin:         utils.String(adminLogin),
			AdministratorLoginPassword: utils.String(adminLoginPassword),
			Version:                    mysql.ServerVersion(version),
			SslEnforcement:             mysql.SslEnforcementEnum(sslEnforcement),
			StorageProfile:             storageprofile,
			CreateMode:                 mysql.CreateMode(createMode),
		},
		Sku:  sku,
		Tags: expandTags(tags),
	}

	future, err := client.Create(ctx, resourceGroup, name, properties)
	if err != nil {
		return err
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read MySQL Server %q (resource group %q) ID", name, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmMySqlServerRead(d, meta)
}

func resourceArmMySqlServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlServersClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM MySQL Server update.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	adminLoginPassword := d.Get("administrator_login_password").(string)
	sslEnforcement := d.Get("ssl_enforcement").(string)
	version := d.Get("version").(string)
	sku := expandMySQLServerSku(d)
	storageprofile := expandMySQLStorageProfile(d)
	tags := d.Get("tags").(map[string]interface{})

	properties := mysql.ServerUpdateParameters{
		ServerUpdateParametersProperties: &mysql.ServerUpdateParametersProperties{
			StorageProfile:             storageprofile,
			AdministratorLoginPassword: utils.String(adminLoginPassword),
			Version:                    mysql.ServerVersion(version),
			SslEnforcement:             mysql.SslEnforcementEnum(sslEnforcement),
		},
		Sku:  sku,
		Tags: expandTags(tags),
	}

	future, err := client.Update(ctx, resourceGroup, name, properties)
	if err != nil {
		return err
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read MySQL Server %q (resource group %q) ID", name, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmMySqlServerRead(d, meta)
}

func resourceArmMySqlServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlServersClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["servers"]

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure MySQL Server %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}

	d.Set("administrator_login", resp.AdministratorLogin)
	d.Set("version", string(resp.Version))
	d.Set("ssl_enforcement", string(resp.SslEnforcement))

	if err := d.Set("sku", flattenMySQLServerSku(d, resp.Sku)); err != nil {
		return err
	}

	if err := d.Set("server_profile", flattenMySQLStorageProfile(d, resp.StorageProfile)); err != nil {
		return err
	}

	flattenAndSetTags(d, resp.Tags)

	// Computed
	d.Set("fqdn", resp.FullyQualifiedDomainName)

	return nil
}

func resourceArmMySqlServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlServersClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["servers"]

	future, err := client.Delete(ctx, resourceGroup, name)
	if err != nil {
		return err
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return err
	}

	return nil
}

func expandMySQLServerSku(d *schema.ResourceData) *mysql.Sku {
	skus := d.Get("sku").(*schema.Set).List()
	sku := skus[0].(map[string]interface{})

	name := sku["name"].(string)
	capacity := sku["capacity"].(int)
	tier := sku["tier"].(string)
	family := sku["family"].(string)

	return &mysql.Sku{
		Name:     utils.String(name),
		Tier:     mysql.SkuTier(tier),
		Capacity: utils.Int32(int32(capacity)),
		Family:   utils.String(family),
	}
}

func expandMySQLStorageProfile(d *schema.ResourceData) *mysql.StorageProfile {
	storageprofiles := d.Get("storageprofile").(*schema.Set).List()
	storageprofile := storageprofiles[0].(map[string]interface{})

	backupRetentionDays := storageprofile["backupretentiondays"].(int)
	geoRedundantBackup := storageprofile["geoRedundantBackup"].(string)
	storageMB := storageprofile["storageMB"].(int)

	return &mysql.StorageProfile{
		BackupRetentionDays: utils.Int32(int32(backupRetentionDays)),
		StorageMB:           utils.Int32(int32(storageMB)),
		GeoRedundantBackup:  mysql.GeoRedundantBackup(geoRedundantBackup),
	}
}

func flattenMySQLServerSku(d *schema.ResourceData, resp *mysql.Sku) []interface{} {
	values := map[string]interface{}{}

	values["name"] = *resp.Name
	values["capacity"] = int(*resp.Capacity)
	values["tier"] = string(resp.Tier)
	values["family"] = string(*resp.Family)

	sku := []interface{}{values}
	return sku
}

func flattenMySQLStorageProfile(d *schema.ResourceData, resp *mysql.StorageProfile) []interface{} {
	values := map[string]interface{}{}

	values["storageMB"] = int(*resp.StorageMB)
	values["backupRetentionDays"] = int(*resp.BackupRetentionDays)
	values["geoRedundantBackup"] = mysql.GeoRedundantBackup(resp.GeoRedundantBackup)

	storageprofile := []interface{}{values}
	return storageprofile
}
