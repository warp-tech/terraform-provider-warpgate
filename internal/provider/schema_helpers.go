package provider

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const rateLimitBytesPerSecondKey = "rate_limit_bytes_per_second"

func optionalIntPointer(d *schema.ResourceData, key string) *int {
	if !configuredValueExists(d, key) {
		return nil
	}

	intValue := d.Get(key).(int)
	return &intValue
}

func optionalInt64Pointer(d *schema.ResourceData, key string) *int64 {
	if !configuredValueExists(d, key) {
		return nil
	}

	intValue := int64(d.Get(key).(int))
	return &intValue
}

func optionalBoolPointer(d *schema.ResourceData, key string) *bool {
	if !configuredValueExists(d, key) {
		return nil
	}

	boolValue := d.Get(key).(bool)
	return &boolValue
}

func optionalStringPointer(d *schema.ResourceData, key string) *string {
	if !configuredValueExists(d, key) {
		return nil
	}

	stringValue := d.Get(key).(string)
	return &stringValue
}

func optionalFloat64Pointer(d *schema.ResourceData, key string) *float64 {
	if !configuredValueExists(d, key) {
		return nil
	}

	floatValue := d.Get(key).(float64)
	return &floatValue
}

func configuredValueExists(d *schema.ResourceData, key string) bool {
	value, diags := d.GetRawConfigAt(cty.GetAttrPath(key))
	if diags.HasError() {
		return false
	}

	return value.IsKnown() && !value.IsNull()
}

func setOptionalInt(d *schema.ResourceData, key string, value *int) error {
	return d.Set(key, value)
}

func setOptionalInt64(d *schema.ResourceData, key string, value *int64) error {
	return d.Set(key, value)
}

func setOptionalBool(d *schema.ResourceData, key string, value *bool) error {
	return d.Set(key, value)
}
