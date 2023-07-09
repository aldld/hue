package hue

import "golang.org/x/exp/slog"

type Scene struct {
	ID       string        `json:"id"`
	IDv1     string        `json:"id_v1"`
	Actions  []SceneAction `json:"actions"`
	Metadata SceneMetadata `json:"metadata"`
	Group    ResourceRef   `json:"group"`
	Status   *SceneStatus  `json:"status,omitempty"`
}

func (_ Scene) Type() ResourceType { return RTypeScene }

type SceneStatus struct {
	Active string `json:"active"`
}

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

func (c *Client) GetScene(id string) (Scene, error) {
	var emptyScene Scene
	var res GetScenesResponse
	if err := c.get("/scene/"+id, &res); err != nil {
		return emptyScene, err
	}
	if len(res.Errors) != 0 {
		return emptyScene, joinHueErrors(res.Errors)
	}
	if len(res.Data) == 0 {
		return emptyScene, nil
	}
	if len(res.Data) > 1 {
		c.log.Warn("got more than one scene", slog.String("id", id))
	}

	return res.Data[0], nil
}

type SceneUpdate struct {
	Actions *[]SceneAction `json:"actions,omitempty"`
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
