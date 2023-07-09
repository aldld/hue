package hue

type LightOn struct {
	On bool `json:"on"`
}

type Dimming struct {
	Brightness  float64 `json:"brightness"`
	MinDimLevel float64 `json:"min_dim_level"`
}

type XY struct {
	x float64
	y float64
}

type Gamut struct {
	Red   XY `json:"red"`
	Green XY `json:"green"`
	Blue  XY `json:"blue"`
}

type Color struct {
	XY        XY     `json:"xy"`
	Gamut     Gamut  `json:"gamut"`
	GamutType string `json:"gamutType"`
}

type ColorTemperature struct {
	Mirek       int         `json:"mirek"`
	MirekValid  bool        `json:"mirek_valid"`
	MirekSchema MirekSchema `json:"mirek_schema"`
}

type MirekSchema struct {
	Min int `json:"mirek_minimum"`
	Max int `json:"mirek_maximum"`
}

type LightMetadata struct {
	Name string `json:"name"`
}

type Light struct {
	ID               string            `json:"id"`
	IDv1             string            `json:"id_v1"`
	Metadata         *LightMetadata    `json:"metadata,omitempty"`
	On               *LightOn          `json:"on,omitempty"`
	Dimming          *Dimming          `json:"dimming,omitempty"`
	Color            *Color            `json:"color,omitempty"`
	ColorTemperature *ColorTemperature `json:"color_temperature,omitempty"`
	Owner            *ResourceRef      `json:"owner,omitempty"`
}

func (_ Light) Type() ResourceType { return RTypeLight }

type GetLightsResponse struct {
	Errors []HueError `json:"errors"`
	Data   []Light    `json:"data"`
}

func (c *Client) GetLights() ([]Light, error) {
	var res GetLightsResponse
	if err := c.get("/light", &res); err != nil {
		return nil, err
	}
	if len(res.Errors) != 0 {
		return nil, joinHueErrors(res.Errors)
	}

	return res.Data, nil
}

type LightUpdate struct {
	On               *LightOn                `json:"on,omitempty"`
	ColorTemperature *ColorTemperatureUpdate `json:"color_temperature,omitempty"`
	Dimming          *DimmingUpdate          `json:"dimming,omitempty"`
	Dynamics         *Dynamics               `json:"dynamics,omitempty"`
}

type ColorTemperatureUpdate struct {
	Mirek int `json:"mirek"`
}

type DimmingUpdate struct {
	Brightness float64 `json:"brightness"`
}

type Dynamics struct {
	DurationMs int `json:"duration"`
}

type UpdateLightResponse struct {
	Errors []HueError    `json:"errors"`
	Data   []ResourceRef `json:"data"`
}

func (c *Client) UpdateLight(ID string, update LightUpdate) error {
	var res UpdateLightResponse
	if err := c.put("/light/"+ID, update, &res); err != nil {
		return err
	}
	if len(res.Errors) != 0 {
		return joinHueErrors(res.Errors)
	}

	return nil
}
