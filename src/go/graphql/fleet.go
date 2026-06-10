package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"m2cpcli/tools"
)

type FleetInfo int

const (
	_ FleetInfo = iota
	FleetAdmin
	FleetCoAdmins
	FleetApps
	FleetDevices
	FleetModel
)

func FleetIdByNameOrId(ctx context.Context, fleetNameOrId string) (UUID, error) {
	var err error
	var fleetId UUID
	if !tools.IsValidUuid(fleetNameOrId) {
		currentName := fleetNameOrId
		fleetId, err = FleetIdByName(ctx, currentName)
		if err != nil {
			return fleetId, err
		}
	} else {
		fleetId = UUID(fleetNameOrId)
	}
	return fleetId, nil
}

func FleetIdByName(ctx context.Context, fleetName string) (UUID, error) {

	queryString := `query($fleetName: String){
fleets(where: {fleetName: {eq: $fleetName}}, order: {id:ASC})
  {
	items{
		id
    } 
    totalCount
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("fleetName", fleetName)

	var result struct {
		FleetByName FleetCollectionSegment `json:"fleets"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	if result.FleetByName.TotalCount > 1 {
		return "", fmt.Errorf("name \"%s\" is not unique", fleetName)
	}
	if result.FleetByName.TotalCount == 0 {
		return "", fmt.Errorf("name \"%s\" not found", fleetName)
	}
	return result.FleetByName.Items[0].Id, nil
}

func FleetById(ctx context.Context, fleetId UUID, fleetInfos []FleetInfo) (*Fleet, error) {
	queryExtension := ""
	for _, info := range fleetInfos {
		switch info {
		case FleetAdmin:
			queryExtension += `
			ownerUser {
				id: userId
				email
				displayName
			}`
		case FleetCoAdmins:
			queryExtension += `
			fleetAdministrators {
				user {
					id: userId
					email
					displayName
				}
			}`
		case FleetModel:
			queryExtension += `
			edgeDeviceModel {      
			  id
			  modelType
			  modelName
			  modelRevision
			  modelArchitecture
			  edgeDeviceModelBridgeSnapDeclarations {
				snapDeclaration{
				  id
				  snapId
				  snapName
				  snapDescription
				}
			  }
			}`
		case FleetDevices:
			queryExtension += `
			edgeDevices {
  			  id
			  deviceSerial
			  deviceLastActivity
			  isDeviceActivated
			  deviceArchitecture
			  fleetId
			  deviceName
			  deviceDescription
			}`
		case FleetApps:
			queryExtension += `
			fleetBridgeSnapRevisions {
			  snapRevision {
				id
				revision
				snapVersion
				uploadMessage
				snapDownloadSize
				createdAt
				snapDeclarationId
				snapStatusId
				snapDeclaration {
				  id
				  snapId
				  snapName         
				  snapDescription
				  snapBase
				  snapDeviceArchitecture
				  snapSummary
				  assertionId
				  tenantId
				  createdAt
				}
				snapStatus{
				  id
				  name
				  description          
				}
				revision
			  }
			}`
		default:
			continue
		}

	}

	queryString := `query($fleetId: UUID!){
	fleet(id: $fleetId) {
		id
		architecture
		fleetName
		description
		` + queryExtension + `
	  }
	}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("fleetId", fleetId)

	var result struct {
		FleetById interface{} `json:"fleet"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	// we convert back to json, because we need to parse multiple times to different target structs
	jsonRaw, _ := json.Marshal(result.FleetById)

	// first we parse the fleet itself
	var fleet Fleet
	if err = json.Unmarshal(jsonRaw, &fleet); err != nil {
		return nil, err
	}

	// then we parse the branched app data, simplify it and attach it to the fleet
	var appsRaw struct {
		EdgeDeviceModel struct {
			EdgeDeviceModelBridgeSnapDeclarations []struct {
				SnapDeclaration SnapDeclaration `json:"snapDeclaration"`
			} `json:"edgeDeviceModelBridgeSnapDeclarations"`
		} `json:"edgeDeviceModel"`
		FleetBridgeSnapRevisions []struct {
			SnapRevision SnapRevision `json:"snapRevision"`
		} `json:"fleetBridgeSnapRevisions"`
	}
	if err = json.Unmarshal(jsonRaw, &appsRaw); err != nil {
		return nil, err
	}

	// identify core apps
	for _, app := range appsRaw.EdgeDeviceModel.EdgeDeviceModelBridgeSnapDeclarations {

		fleet.CoreSnaps = append(fleet.CoreSnaps, app.SnapDeclaration)
	}

	// merge apps into fleet
	fleet.SnapRevisions = []SnapRevision{}
	for _, app := range appsRaw.FleetBridgeSnapRevisions {

		fleet.SnapRevisions = append(fleet.SnapRevisions, app.SnapRevision)
	}

	// parse wrapped co-administrators
	var coAdminsRaw struct {
		FleetAdministrators []struct {
			User User `json:"user"`
		} `json:"fleetAdministrators"`
	}
	if err = json.Unmarshal(jsonRaw, &coAdminsRaw); err != nil {
		return nil, err
	}
	// merge co-admins into fleet
	fleet.CoAdmins = []User{}
	for _, coAdmin := range coAdminsRaw.FleetAdministrators {
		fleet.CoAdmins = append(fleet.CoAdmins, coAdmin.User)
	}

	//
	return &fleet, nil
}

func FleetBridgeSnapRevisionId(ctx context.Context, fleetId UUID, snapDeclarationId UUID) (UUID, error) {
	queryString := `query($fleetId: UUID! $snapDeclarationId: UUID!){
  fleetBridgeSnapRevisions(where: {and:
   [{fleetId: { eq: $fleetId}}, {snapRevision: {snapDeclaration: {id: {eq: $snapDeclarationId}}}}]
   }){
    items{
      id
	  fleetId
      snapRevisionId
    }
    totalCount  
  }
}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("fleetId", fleetId)
	req.Var("snapDeclarationId", snapDeclarationId)

	var result struct {
		Bridge FleetBridgeSnapRevisionCollectionSegment `json:"fleetBridgeSnapRevisions"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	if result.Bridge.TotalCount > 1 {
		return "", fmt.Errorf("the fleetBridgeSnapRevisionId for the tuple \"%s %s\" is not unique",
			fleetId, snapDeclarationId)
	}
	if result.Bridge.TotalCount == 0 {
		return "", fmt.Errorf("fleetBridgeSnapRevisionId not found")
	}
	return result.Bridge.Items[0].Id, nil
}

func FleetBridgeSnapRevisions(ctx context.Context, fleetId UUID) ([]FleetBridgeSnapRevision, error) {
	// TODO: we fetch much more data than we need.
	queryString := `query($fleetId: UUID) { 
  result: fleetBridgeSnapRevisions(where: {fleetId: {eq: $fleetId}}) {
    items {
      id
	  fleetId
      snapRevisionId
      snapRevision {
        uploadMessage
        revision
		snapDownloadSize
        createdAt
		snapVersion
        snapDeclarationId
        snapDeclaration {
          snapId
          snapDeviceArchitecture
          snapName
          snapDescription
        }
      }
    }
    totalCount
  }
}`
	// TODO: use proper pagination
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("fleetId", fleetId)

	var result struct {
		Bridge FleetBridgeSnapRevisionCollectionSegment `json:"result"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return result.Bridge.Items, nil
}
