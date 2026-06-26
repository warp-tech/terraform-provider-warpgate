package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const rateLimitBytesPerSecondKey = "rate_limit_bytes_per_second"

func optionalIntPointer(d *schema.ResourceData, key string) *int {
	if value, ok := d.GetOkExists(key); ok {
		intValue := value.(int)
		return &intValue
	}

	return nil
}

func setOptionalInt(d *schema.ResourceData, key string, value *int) error {
	return d.Set(key, value)
}
