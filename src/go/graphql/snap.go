package graphql

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
)

// TODO: move this function upwards and re-use
func RenderQueryOrMutation(queryOrMutationTemplate string, params interface{}) (string, error) {
	tmpl, err := template.New("arbitrary").Parse(queryOrMutationTemplate) // TODO: Why does the template have a name?
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, params)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func buildSnapDatabaseIdBySnapIdQueryString(snapId string) (string, error) {
	queryTemplate := `query SnapDatabaseIdBySnapId($take: Int, $skip: Int){
result: snapDeclarations(
take: $take
skip: $skip
order: {id:ASC}
where: {
	{{- with .SnapId -}}
		snapId: {eq: {{- printf "%q" . -}} }
	{{- end }}
}
)
{
    items{
        id
    }
    pageInfo{ hasNextPage hasPreviousPage }  
    totalCount  
  }
}`

	type Parameters struct {
		SnapId string
	}
	params := Parameters{snapId}
	return RenderQueryOrMutation(queryTemplate, params)
}

// SnapDatabaseIdBySnapId returns the `id` from the database, which is different from the `snapId`.
func SnapDatabaseIdBySnapId(ctx context.Context, snapId string) (UUID, error) {
	queryString, err := buildSnapDatabaseIdBySnapIdQueryString(snapId)
	if err != nil {
		return "", err
	}

	client, request := PrepareClientAndRequest(ctx, queryString)

	snapDeclarations, err := RunPaginated[SnapDeclaration](ctx, client, request)
	if err != nil {
		return "", err
	}

	if len(snapDeclarations) > 1 {
		return "", fmt.Errorf("snapId \"%s\" is not unique", snapId)
		//TODO: be nice and print a table of the matches?
	}
	if len(snapDeclarations) == 0 {
		return "", fmt.Errorf("snapId \"%s\" not found", snapId)
	}
	return snapDeclarations[0].Id, nil
}

func buildSnapDatabaseIdBySnapNameAndArchitectureQueryString(snapName, snapDeviceArchitecture string) (string, error) {
	queryTemplate := `query SnapDatabaseIdBySnapNameAndArch($take: Int, $skip: Int){
result: snapDeclarations(
take: $take
skip: $skip
order: {id:ASC}
where: {
	snapName: {eq: {{- printf "%q" .SnapName -}} }
	snapDeviceArchitecture: {eq: {{- printf "%q" .SnapDeviceArchitecture -}} }	
}
)
{
    items{
        id
    }
    pageInfo{ hasNextPage hasPreviousPage }  
    totalCount  
  }
}`

	type Parameters struct {
		SnapName               string
		SnapDeviceArchitecture string
	}
	params := Parameters{snapName, snapDeviceArchitecture}
	return RenderQueryOrMutation(queryTemplate, params)
}

// SnapDatabaseIdByNameAndArchitecture returns the `id` from the database. Attention! Not the `snapId`!
func SnapDatabaseIdByNameAndArchitecture(ctx context.Context, snapName, snapDeviceArchitecture string) (UUID, error) {
	queryString, err := buildSnapDatabaseIdBySnapNameAndArchitectureQueryString(snapName, snapDeviceArchitecture)
	if err != nil {
		return "", err
	}

	client, request := PrepareClientAndRequest(ctx, queryString)
	accumulatedSnapDeclarations, err := RunPaginated[SnapDeclaration](ctx, client, request)
	if err != nil {
		return "", err
	}

	if len(accumulatedSnapDeclarations) > 1 {
		return "", fmt.Errorf("snapName \"%s\" and snapDeviceArchitecture \"%s\" is not unique",
			snapName, snapDeviceArchitecture)
		//TODO: be nice and print a table of the matches?
	}
	if len(accumulatedSnapDeclarations) == 0 {
		return "", fmt.Errorf("snapName \"%s\" and snapDeviceArchitecture \"%s\" not found",
			snapName, snapDeviceArchitecture)
	}
	return accumulatedSnapDeclarations[0].Id, nil
}

func SnapDownloadUrlBySnapRevisionId(ctx context.Context, snapRevisionId UUID) (string, error) {
	queryString := `
	query SnapDownloadUrlBySnapRevisionId($snapRevisionId: UUID!){
		downloadSnap(input: {
    		snapRevisionId: $snapRevisionId
  		}) 
		{
			downloadUrl
		}
	}`
	client, request := PrepareClientAndRequest(ctx, queryString)
	request.Var("snapRevisionId", snapRevisionId)

	var result struct {
		SnapDownloadOutput SnapDownloadOutput `json:"downloadSnap"`
	}

	err := client.Run(ctx, request, &result)
	if err != nil {
		return "", err
	}
	return result.SnapDownloadOutput.DownloadUrl, nil
}

func buildSnapRevisionIdBySnapDeclarationIdAndRevisionQueryString(snapDeclarationId UUID, revision int) (string, error) {
	queryTemplate := `query{
result: snapRevisions(
   take: 1,
   skip: 0,
   where: {
    snapDeclaration: {
      id: {eq: {{- printf "%q" .SnapDeclarationId -}} }
    }
	{{if .Revision}}
	revision: {eq: {{- printf "%d" .Revision -}} }   
	{{end}}
   },
   order: {revision:DESC}){
    items{
      id
      revision
      snapVersion
      snapDeclaration {
        id
        snapName
        snapDeviceArchitecture
      }
    }
    totalCount  
  }
}`

	type Parameters struct {
		SnapDeclarationId UUID
		Revision          int
	}
	params := Parameters{snapDeclarationId, revision}
	return RenderQueryOrMutation(queryTemplate, params)
}

func SnapRevisionIdBySnapDeclarationIdAndRevision(ctx context.Context, snapDeclarationId UUID, revision int) (UUID, error) {
	queryString, err := buildSnapRevisionIdBySnapDeclarationIdAndRevisionQueryString(snapDeclarationId, revision)
	if err != nil {
		return "", err
	}

	client, request := PrepareClientAndRequest(ctx, queryString)
	snapRevisions, err := RunPaginated[SnapRevision](ctx, client, request)
	if err != nil {
		return "", err
	}

	if len(snapRevisions) == 0 {
		return "", fmt.Errorf("snapDeclarationId \"%s\" and revision \"%d\" not found",
			snapDeclarationId, revision)
	}
	return snapRevisions[0].Id, nil
}

func buildSnapDeclarationByDatabaseIdQueryString(withDeclaration bool, withRevisions bool) (string, error) {
	queryTemplate := `query SnapInfoById($snapDatabaseId: UUID!){
result: snapDeclaration(id: $snapDatabaseId){
	id
{{ with .Declaration }}
    snapName
    snapDeviceArchitecture
    tenantId
    tenant {
		tenantId
		alias
    }
    snapSummary
    snapDescription
    snapBase
    snapId
    createdAt
	assertionId
{{ end }}
{{ with .Revisions }}
	snapRevisions {
      id
      uploadMessage
      revision
      snapDownloadSize
      createdAt
      snapVersion
      snapSha3
	  snapStatus {
		name
	 	description
		isUnrated
		isUsable
	  }
      description
      fleetBridgeSnapRevisions {
        fleet{
			id
          	fleetName
        }
      }
    }
{{ end }}
  }
}`

	type Parameters struct {
		Declaration bool
		Revisions   bool
	}
	params := Parameters{
		Declaration: withDeclaration,
		Revisions:   withRevisions,
	}

	return RenderQueryOrMutation(queryTemplate, params)
}

func SnapDeclarationByDatabaseId(ctx context.Context, snapDatabaseId UUID, withDeclaration bool, withRevisions bool) (*SnapDeclaration, error) {
	queryString, err := buildSnapDeclarationByDatabaseIdQueryString(withDeclaration, withRevisions)
	if err != nil {
		return nil, err
	}

	client, request := PrepareClientAndRequest(ctx, queryString)
	request.Var("snapDatabaseId", snapDatabaseId)

	var result struct {
		Info SnapDeclaration `json:"result"`
		// TODO: The revisions table is not sorted by revision!
	}

	err = client.Run(ctx, request, &result)
	if err != nil {
		return nil, err
	}

	declaration := result.Info

	// reformat associated fleets
	if declaration.SnapRevisions != nil {
		for i := range declaration.SnapRevisions {
			if declaration.SnapRevisions[i].FleetBridgeSnapRevision != nil {
				for _, fleetBridgeSnapRevision := range declaration.SnapRevisions[i].FleetBridgeSnapRevision {
					declaration.SnapRevisions[i].Fleets = append(declaration.SnapRevisions[i].Fleets, fleetBridgeSnapRevision.Fleet)
				}
			}
			declaration.SnapRevisions[i].FleetBridgeSnapRevision = nil
		}
	}

	return &declaration, nil
}

func SnapRevisionDatabaseIdByNameArchAndRevision(ctx context.Context, snapName string, arch string, rev int) (*UUID, error) {
	queryString := `
query RevisionsByNameArchRevision($snapName: String!, $snapArch: String!, $revision: Int!){
  result: snapRevisions(take: 1, skip: 0,
   where: {
    snapDeclaration: {
      snapName: {eq: $snapName}
      snapDeviceArchitecture: {eq: $snapArch}
    }    
    revision: {eq: $revision}
   },
   order: {revision:DESC}){
    items{
      id
      revision
      snapVersion
      snapDeclaration {
        snapName
        snapDeviceArchitecture
      }
    }
    totalCount  
  }
}`
	client, request := PrepareClientAndRequest(ctx, queryString)
	request.Var("snapName", snapName)
	request.Var("snapArch", arch)
	request.Var("revision", rev)

	var result struct {
		Output SnapRevisionsOutput `json:"result"`
	}

	err := client.Run(ctx, request, &result)
	if err != nil {
		return nil, err
	}
	if len(result.Output.SnapRevisions) < 1 {
		return nil, fmt.Errorf("no snap found with name %s, architecture %s, and revision %d found", snapName, arch, rev)
	}
	snapDatabaseId := result.Output.SnapRevisions[0].Id

	return &snapDatabaseId, nil
}

func GenerateSnapUploadUrl(ctx context.Context, fileSize int) (*SnapUploadUrl, error) {
	mutationString := `
mutation GenerateSnapUploadUrl($fileSize: Int!) {
  generateSnapUploadUrl(input: {
    fileSize: $fileSize 
  }) {
    uploadUrl
    continuesToken
  }
}`
	client, request := PrepareClientAndRequest(ctx, mutationString)
	request.Var("fileSize", fileSize)

	var result struct {
		Output SnapUploadUrl `json:"generateSnapUploadUrl"`
	}

	err := client.Run(ctx, request, &result)
	if err != nil {
		return nil, err
	}

	return &result.Output, nil
}
