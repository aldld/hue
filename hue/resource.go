package hue

type ResourceType string

const (
	RTypeDevice                     ResourceType = "device"
	RTypeBridgeHome                 ResourceType = "bridge_home"
	RTypeRoom                       ResourceType = "room"
	RTypeZone                       ResourceType = "zone"
	RTypeLight                      ResourceType = "light"
	RTypeButton                     ResourceType = "button"
	RTypeRelativeRotary             ResourceType = "relative_rotary"
	RTypeTemperature                ResourceType = "temperature"
	RTypeLightLevel                 ResourceType = "light_level"
	RTypeMotion                     ResourceType = "motion"
	RTypeEntertainment              ResourceType = "entertainment"
	RTypeGroupedLight               ResourceType = "grouped_light"
	RTypeDevicePower                ResourceType = "device_power"
	RTypeZigbeeBridgeConnectivity   ResourceType = "zigbee_bridge_connectivity"
	RTypeZigbeeConnectivity         ResourceType = "zigbee_connectivity"
	RTypeZgpConnectivity            ResourceType = "zgp_connectivity"
	RTypeBridge                     ResourceType = "bridge"
	RTypeZigbeeDeviceDiscovery      ResourceType = "zigbee_device_discovery"
	RTypeHomekit                    ResourceType = "homekit"
	RTypeMatter                     ResourceType = "matter"
	RTypeMatterFabric               ResourceType = "matter_fabric"
	RTypeScene                      ResourceType = "scene"
	RTypeEntertainmentConfiguration ResourceType = "entertainment_configuration"
	RTypePublicImage                ResourceType = "public_image"
	RTypeAuthV1                     ResourceType = "auth_v1"
	RTypeBehaviorScript             ResourceType = "behavior_script"
	RTypeBehaviorInstance           ResourceType = "behavior_instance"
	RTypeGeofence                   ResourceType = "geofence"
	RTypeGeofenceClient             ResourceType = "geofence_client"
	RTypeGeolocation                ResourceType = "geolocation"
	RTypeSmartScene                 ResourceType = "smart_scene"
)

type Resource interface {
	Type() ResourceType
}

type ResourceRef struct {
	ID   string       `json:"rid"`
	Type ResourceType `json:"rtype"`
}
