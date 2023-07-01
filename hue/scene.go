package hue

type Scene struct {
	ID       string        `json:"id"`
	IDv1     string        `json:"id_v1"`
	Actions  []SceneAction `json:"actions"`
	Metadata SceneMetadata `json:"metadata"`
	Group    ResourceRef   `json:"group"`
}

func (_ Scene) Type() ResourceType { return RTypeScene }

type SceneAction struct {
	Target ResourceRef `json:"target"`
	Action Action      `json:"action"`
}

type Action struct {
	On               *LightOn                `json:"on,omitempty"`
	Dimming          *DimmingAction          `json:"dimming,omitempty"`
	Color            *ColorAction            `json:"color,omitempty"`
	ColorTemperature *ColorTemperatureAction `json:"color_temperature,omitempty"`

	// TODO: This is not complete.
}

type DimmingAction struct {
	Brightness float64 `json:"brightness"`
}

type ColorAction struct {
	XY XY `json:"xy"`
}

type ColorTemperatureAction struct {
	Mirek int `json:"mirek"`
}

type SceneMetadata struct {
	Name string `json:"name"`
}

type GetScenesResponse struct {
	Errors []HueError `json:"errors"`
	Data   []Scene    `json:"data"`
}

func (c *Client) GetScenes() ([]Scene, error) {
	var res GetScenesResponse
	if err := c.get("/scene", &res); err != nil {
		return nil, err
	}
	if len(res.Errors) != 0 {
		return nil, joinHueErrors(res.Errors)
	}

	return res.Data, nil
}

type SceneUpdate struct {
	Actions *[]SceneAction `json:"action,omitempty"`
}

type UpdateSceneResponse struct {
	Errors []HueError    `json:"errors"`
	Data   []ResourceRef `json:"data"`
}

func (c *Client) UpdateScene(ID string, update SceneUpdate) error {
	var res UpdateSceneResponse
	if err := c.put("/scene/"+ID, update, &res); err != nil {
		return err
	}
	if len(res.Errors) != 0 {
		return joinHueErrors(res.Errors)
	}

	return nil
}
