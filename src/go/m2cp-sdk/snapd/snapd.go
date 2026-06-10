package snapd

import (
	"encoding/json"
	"fmt"
	"m2cp"
	"strings"
	"time"
)

type FindResult struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status-code"`
	Status     string `json:"status"`
	Result     []Snap `json:"result,omitempty"`
}

type FindStatus struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status-code"`
	Status     string `json:"status"`
}

type FindError struct {
	Result map[string]string `json:"result,omitempty"`
}
type Snap struct {
	Apps             []SnapApp           `json:"apps"`
	Base             string              `json:"base"`
	Channel          string              `json:"channel"`
	Confinement      string              `json:"confinement"`
	Contact          string              `json:"contact"`
	Description      string              `json:"description"`
	DevMode          bool                `json:"devmode"`
	Developer        string              `json:"developer"`
	Icon             string              `json:"icon"`
	Id               string              `json:"id"`
	IgnoreValidation bool                `json:"ignore-validation"`
	InstallDate      time.Time           `json:"install-date"`
	JailMode         bool                `json:"jailmode"`
	License          string              `json:"license"`
	Links            map[string][]string `json:"links"`
	MountedFrom      string              `json:"mounted-from"`
	Name             string              `json:"name"`
	Private          bool                `json:"private"`
	Publisher        struct {
		DisplayName string `json:"display-name"`
		Id          string `json:"id"`
		Username    string `json:"username"`
		Validation  string `json:"validation"`
	} `json:"publisher"`
	Revision        string `json:"revision"`
	Status          string `json:"status"`
	StoreUrl        string `json:"store-url"` // This is a M2CP extension
	Summary         string `json:"summary"`
	TrackingChannel string `json:"tracking-channel"`
	Type            string `json:"type"`
	Version         string `json:"version"`
}

type SnapApp struct {
	Snap        string `json:"snap"`
	Name        string `json:"name"`
	Daemon      string `json:"daemon,omitempty"`
	DaemonScope string `json:"daemon-scope,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	Active      bool   `json:"active,omitempty"`
}

func GetSnapsInstalled(ctp m2cp.ContextPlus) ([]Snap, error) {
	c, err := newClient(ctp)
	if err != nil {
		return nil, err
	}
	return c.getSnapsInstalled()

}

func GetModelAssertion(ctp m2cp.ContextPlus) (*ModelAssertion, error) {
	c, err := newClient(ctp)
	if err != nil {
		return nil, err
	}
	return c.getModelAssertion()
}

func StopService(ctp m2cp.ContextPlus, service string) error {
	c, err := newClient(ctp)
	if err != nil {
		return err
	}

	res, err := c.apiPost("/v2/snaps/"+service, map[string]interface{}{"action": "disable"})
	if err != nil {
		return err
	}

	var result resultServiceControl
	err = json.Unmarshal(res, &result)
	if err != nil {
		return err
	}
	if result.StatusCode == 400 && strings.Contains(result.Result.Message, "already disabled") {
		return nil
	}
	if result.StatusCode != 202 {
		return fmt.Errorf("failed to stop service: %s", result.Result.Message)
	}

	return awaitChangeDone(ctp, result.Change)
}

func StartService(ctp m2cp.ContextPlus, service string) error {
	c, err := newClient(ctp)
	if err != nil {
		return err
	}
	res, err := c.apiPost("/v2/snaps/"+service, map[string]interface{}{"action": "enable"})
	if err != nil {
		return err
	}

	var result resultServiceControl
	err = json.Unmarshal(res, &result)
	if err != nil {
		return err
	}
	if result.StatusCode == 400 && strings.Contains(result.Result.Message, "already enabled") {
		return nil
	}
	if result.StatusCode != 202 {
		return fmt.Errorf("failed to stop service: %s", result.Result.Message)
	}

	return awaitChangeDone(ctp, result.Change)
}

func Find(ctp m2cp.ContextPlus, query string) (*FindResult, error) {
	path := fmt.Sprintf("v2/find?q=%s", query)
	c, err := newClient(ctp)
	if err != nil {
		return nil, err
	}
	res, err := c.apiGet(path)
	if err != nil {
		return nil, fmt.Errorf("find %s failed: %s", query, err)
	}

	var status FindStatus
	err = json.Unmarshal(res, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to parse find status: %w", err)
	}

	if status.StatusCode != 200 {
		var statusError FindError
		err = json.Unmarshal(res, &statusError)
		if err != nil {
			return nil, fmt.Errorf("could not decode error response: %s", err)
		}

		return nil, fmt.Errorf("failed calling appstore: %d - %s, %s", status.StatusCode, status.Status, statusError.Result["message"])
	}

	var result FindResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse find result: %w", err)
	}

	return &result, nil
}

func awaitChangeDone(ctp m2cp.ContextPlus, changeId string) error {
	c, err := newClient(ctp)
	if err != nil {
		return err
	}
	for {
		res, err := c.apiGet("/v2/changes/" + changeId)
		if err != nil {
			return err
		}

		var result resultObserveChange
		err = json.Unmarshal(res, &result)
		if err != nil {
			return err
		}
		if result.Result.Ready {
			if result.Result.Status != "Done" {
				return fmt.Errorf("failed to await change done: %s", result.Result.Err)
			}
			return nil
		}
		if result.StatusCode != 200 {
			return fmt.Errorf("failed to await change done: %s", result.Result.Err)
		}
		ctp.Sleep(time.Second)
	}
}

type resultServiceControl struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status-code"`
	Status     string `json:"status"`
	Result     struct {
		Message string `json:"message"`
	} `json:"result,omitempty"`
	Change string `json:"change,omitempty"`
}

type resultObserveChange struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status-code"`
	Status     string `json:"status"`
	Result     struct {
		Id      string `json:"id"`
		Kind    string `json:"kind"`
		Summary string `json:"summary"`
		Status  string `json:"status"`
		Ready   bool   `json:"ready"`
		Err     string `json:"err"`
	}
}
