# GraphQL and Command Line Interface Examples

We offer 3 levels of permissions:

• **All**: This level allows users to perform operations across all tenants. For instance, with permissions like SnapDeclaration.Read.All and SnapDeclaration.Create.All, users can read across all tenants and create new ones within the current tenant.

• **Tenant**: Users at this level can only perform operations within their specific tenant. Any data retrieved from the database will be filtered to match the current tenant.

• **Own**: Users with this level can only conduct operations on records they own. This restriction ensures that they can solely interact with records that belong to them.

*NOTE: the tenant "mlpa" has a special role: it's providing the basic operating system snaps. So all tenants must be able to use (=read) those snaps and install them on their devices. Only the "mlpa" tenant himself has write-access, though.

## m2cp user login

NOTE: Authentication for CLI and GUI works differently. This sections describes CLI authentication.
This is about authentication CLI<-->snapstore-API using ssh-keys, 

```graphql
mutation{
    userLogin(userEmail: "testuser@ml-pa.com",
    sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDBmefSwoqMqG+9wLNq+O5nI/MP4Sjq/7ArGRlX/UlDG4kQSbaB9yF9OknPQYGVzDAOfuKfEITiYq2yyrvAvdyc/iAE/YHHb5LzRm1JeRGaE6m+RAha5rb7BCwSjGxg9wA+3VSTuRm7I31gpdgYA/n2G1yl2cew7XpP9LFT8XhJ5KBEaMqFBikjOxBTTrAj7fBG2NGid1oSC1GhpNWbiOvwKKUoB4gLj4vRmFXFeFjl+fWXX3Zqcr8X6ennTjPeaqHpVWJY5i3WE2YaJtfIOEum2gnSlmN4dnnNiu6DAfodV9k1NAC6ZqzZ/19+qBWlRrBjfWOaI5pRzvRfv0sKOhGzSlKZ2WEKg3ZXWbHOHtKyM5pUIzvLZ+A6GHCzBVV8qvKs/RVYE/VQzHFhGrVeKAGgj5vZWuOZ2UetarTnduF1I2Uqb2gjaukWFHeSee9d3DWrFO9fvJmPcNyjbDq6N5QXEeSz03W38JIoWOl40d84cSC0G5kVogKXhKj3eKvK/CM= flo@MLPA-NB105")
    {
        success
        challenge
        errorMessage
        errorCode
    }
}
```
plus second step
```graphql
mutation{
    userLoginChallengeResponse(challenge: "53018564BF1A19A7111CBC958FD6ADD79868363AACF91A3FD0991A0C17DB2720A9A45E383F6066431B6783D0EA3A80B7",
    response: "SIGNED_NONCE_VALUE_HERE"){
        success
        token
        errorMessage
        errorCode
    }
}
```
case does not matter for hex strings.

JWT must be passed in the header of all following queries and mutations!
Like this:

```
HTTP Headers
Name           Value
Authorization  Bearer eyJhbGciOi...
```
That value is `"Bearer" + " " + JWT`.

## m2cp user status

Works in CLI. Needs the backend to map tenantId to tenantAlias.

## m2cp user logout

Works.

## List all existing users

Not planned to be integrated in `m2cp` CLI?
```graphql
query ListAllExistingUsers{
  users{
  items{
    id
    displayName
    email
    tenant {
      id
      tenantName
      alias
    }    
  }
  pageInfo{
    hasNextPage
    hasPreviousPage
  }  
  }  
}
```

## m2cp user tenant list

```graphql
query ListAllExistingTenants{
  tenants(take: 1000, skip: 0){
    items{
      id
      tenantName
      alias          
    }
    pageInfo{ hasNextPage hasPreviousPage }  
    totalCount  
  }
}

query TenantById{
  tenants(where: {id: {eq: "226df810-b146-459e-bc64-defdfc5e6062"}}) {
    items{
      tenantName
      alias          
    }
    pageInfo{
      hasNextPage
      hasPreviousPage
    }  
    totalCount  
  }
}
```

## m2cp user tenant switch

Select a different **tenant** for the currently logged in user

A single snapstore can host several tenants, which are "substore", e.g. our dev-store has "mlpa" and "knorr" as tenants.
A user has a default tenant registered in the DB, so when he logs in (starts a new session), he's in that tenant. The currently
selected tenant must be in the JWT token, not in the DB - because the user might have another client open with another tenant selected.
A user can use this command to switch to another _tenant_.
This mutation has two arguments, "tenantId" and "alias," both of which have null as their default values. To clarify that at least one of these arguments should have a non-null value.

```graphql
mutation{
  userSwitchTenant(alias: "mlpa")
  {
    success
    token
    errorMessage
  }
}
```
OR
```graphql
mutation{
  userSwitchTenant(tenantId: "226df810-b146-459e-bc64-defdfc5e6062")
  {
    success
    token
    errorMessage
  }
}
```

## m2cp user tenant set-default

It's also possible to change the default tenant for new sessions:
The `UserSetting` table is designed to store user-specific settings and configuration data within a database.
The `DefaultTenantId` key is used to define and store the current user's default tenant ID. 

```graphql
mutation($userSettings: [UserSettingUpdateInput!]!){
  updateUserSettings(userSettings: $userSettings){
    key
    value
  }
}
```

with variables:

```json
{
  "userSettings" : [
    {
      "key" :"DefaultTenantId",
      "value": "226df810-b146-459e-bc64-defdfc5e6062"
    }
  ]
}
```

or

```graphql
mutation{
  updateUserSettings(userSettings: [
    {
      key : "DefaultTenantId",
      value :"226df810-b146-459e-bc64-defdfc5e6062"
    }
  ])
  {
    key
    value
  }
}
```

To learn, which tenants exist, you can query:
The query returns authorized tenants for the current user. "alias" is the short name of the tenants.

```graphql
query{
  tenants{
    items{
      id
      tenantName
      alias
    }
    pageInfo{
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```


## devices
return list of devices, filters allowed, include online status (queried from IoT hub)
- querying online status is implemented in DeviceHub

`deviceLastActivity` is the last time the device contacted the store. This is updated every 24h by default, but can be triggered manually by sending a refresh call to the device via rpc.
`isDeviceActivated` true: device is allowed to talk to the store and to the IoT hub. When a new device contacts the store for the first time - and is not known there, yet - it will be added to the DB, but not activated, yet. An administrator will see the new device listed and can activate it.
`deviceSerial` (uuid) is the official name of the device used by snapd. It's unique and can't be changed. The `id` is the DB-internal id of the device and musn't be confused with the serial!

```graphql
query {
  result:edgeDevices {
    items {
      id
      deviceSerial
      deviceArchitecture
      deviceName
      deviceDescription
      deviceLastActivity
      isDeviceActivated
      deviceOnlineStatus
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```

Example: filter by activation status
```graphql
query {
  result:edgeDevices(where: { isDeviceActivated: { eq: true } }) {
    items {
      id
      # ...
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}

```

Example: filter by last activity and sort by last activity
```graphql
query {
  result:edgeDevices(where: { deviceLastActivity: { gt: "2023-01-25" } }, order: { deviceLastActivity: ASC }) {
    items {
      id
      # ...
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```


## device

Devices are known to users by their _name_ and their _serial_. The database though uses an internal db id, which is not 
identical to the serial. To fetch the id of a device from the serial, use this query:

```graphql
query {
  result:edgeDevices(where: { deviceSerial: { eq: "d09b1153-f9ff-4c3d-97d2-46976208dd92" } }) {
    items {
      id
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```

Or if you have the name of the device:
```graphql
query {
  result:edgeDevices(where: { deviceName: { eq: "raspi-5" } }) {
    items {
      id
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```

Once you know the id of a device, you can query for more information:

```graphql
query foo {
  result:edgeDevice(id: "3e823a73-edb5-4334-7fab-08db464751da") {
    deviceSerial
    deviceArchitecture
    deviceName
    deviceDescription
    deviceLastActivity
    isDeviceActivated
  }
}
```

NOTE: the id is *not* the deviceSerial!

```graphql
query{
  result:edgeDevice(id: "3e823a73-edb5-4334-7fab-08db464751da") {
    id deviceSerial
  }
}
```

You can resolve the name of a device and query device details in one step:

```graphql
query {
  result:edgeDevices(where: { deviceSerial: { eq: "d09b1153-f9ff-4c3d-97d2-46976208dd92" } }) {
    items {
      id
      deviceSerial
      deviceArchitecture
      deviceName
      deviceDescription
      deviceLastActivity
      isDeviceActivated
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```


To get more details about a device, we need to make a more complex query:

```graphql
query deviceDetails{
   result:device(id: "3e823a73-edb5-4334-7fab-08db464751da") {
      # get basic device info
      id
      deviceSerial
      deviceName
      deviceDescription
      deviceArchitecture
      deviceLastActivity
      
      tenant{
        id
        tenantName
      }
  
      # get current installation state
      edgeDeviceLastInstallations {
         id
         snapRevision {
            id
            snapDeclaration {
               snapName
               snapDescription
            }
         }
      }
      
      # get planned installation state
      edgeDeviceBridgeFleets {
         fleet {
            id
            fleetName
            fleetDescription
            # find out which version of each snap should be installed
            fleetBridgeSnapRevisions {
               snapRevision {
                  id
                  snapRevisionName
                  snapRevisionVersion
                  snapRevisionRevision
                  snapRevisionArchitecture
                  snapRevisionBase
                  snapDeclaration {
                     id
                     snapName
                  }
               }
            }
         }
      }
   }
}
```

The _deviceLastKnownInstallationState_ is only updated, when a device contacts the store. This happens every 24h
by default, but can be triggered manually by sending a refresh call to the device via rpc:
    
```graphql
query deviceRefresh{
  result:deviceRefresh(deviceId: "3e823a73-edb5-4334-7fab-08db464751da") {
    success
    errorMessage
  }
}
```

After a successful refresh call, you can query the device details again to see the updated installation state.

We can also query some more details from a device, which is gathered by the beckend via an rpc call to the device.
This way, we can ask the device for live information:

```graphql
query deviceDetails{
   result:device(id: "3e823a73-edb5-4334-7fab-08db464751da") {
      # get basic device info
      id
      deviceSerial
       
      # ask for active network interfaces
      deviceNetworkInterfaces {
         deviceNetworkInterface {
            id
            deviceNetworkInterfaceName 
            deviceNetworkInterfaceAddresses {
                deviceNetworkInterfaceProtocol
                deviceNetworkInterfaceAddress
            }
         }
      }
      
      # ask for active users (the device will only tell the user email, the backend will query the db for the users id)
      deviceUsers {
         deviceUser {
            id
            deviceUserName
            deviceUserEmail
         }
      }
   }
}
```

When programming for GUI, it's a good idea to do slow queries like this in a separate thread, so the user does not have to wait for the result.

Example output formatted for CLI:
```
$ m2cp device -all d09b1153-f9ff-4c3d-97d2-46976208dd92
Tenant:              mlpa
Serial:              d09b1153-f9ff-4c3d-97d2-46976208dd92
Name:                phy-mlpa-dev-1
Description:         description-gRenObfj
Architecture:        arm64
Last Store Activity: 2023-08-29T09:09:35.473Z 
Fleet Name:          special-fleet-name
Fleet Description:   this fleet is blah blah blah... 

Snaps:
Tenant   Base      Name                           Current      Planned   
mlpa     core20    core18                         -            1.2.3      # means: install       
mlpa     core20    core20                         20.20.20     20.20.21   # means: update                        
mlpa     core22    core22                         22.20.20     -                         
mlpa     core20    dotnet-hello                   0.3          -                         
mlpa     core20    m2cp-borderrouterconnection    0.0.12       -                         
mlpa     core20    m2cp-coap                      0.1.1        -                         
mlpa     core20    m2cp-logstat                   0.7.2        -
mlpa     core20    m2cp-rpc2emit                  0.13.1       -                         
mlpa     core20    m2cp-simsensor                 0.0.9        -                         
mlpa     core20    m2cp-statemachine              0.0.5        -                         
mlpa     core20    m2cp-subscribe2rpc             0.26.1       -                         
mlpa     core20    m2cp-systest                   0.13.2       -                         
mlpa     core20    m2cpd                          0.3.3        -                         
mlpa     core20    phyboard-pollux-gadget         0.0.7        -                         
mlpa     core20    phyboard-pollux-kernel         5.15.71      -                         
mlpa     core20    rabbitmq-server-snap           3.10.8       -                         
mlpa     core20    snapd                          2.57.2.18    -                         

Users: 
Id  Username     Email
1   ubuntupicore tony.kurz@ml-pa.com
2   janos        janos.brodbeck@ml-pa.com
3   jan          jan.mohr@ml-pa.com
4   iryna        iryna.tyshchenko@ml-pa.com

Network Interfaces:
Name      Protocol  Address
can0      -         None
can1      -         None
eth0      IPv4      192.168.99.97
eth0      IPv6      fda7:a4d4:4cf0:0:522d:f4ff:fe2b:2b75
eth0      IPv6      fdde:2cb4:e4f1:0:522d:f4ff:fe2b:2b75
eth0      IPv6      2001:1438:400c:7771:522d:f4ff:fe2b:2b75
eth0      IPv6      2001:1438:400c:7700:522d:f4ff:fe2b:2b75
eth0      IPv6      fe80::522d:f4ff:fe2b:2b75%2
eth1      -         None
lo        IPv4      127.0.0.1
lo        IPv6      ::1
```




## device-modify
update device details
- change name and description of a device (names must be unique)
- change activation state of device (if true, device can talk to store and IoT hub, if false, not. New devices start with false and must be activated by an admin. devices should only be deactivated if they are lost, stolen, broken or similar)
- change updates activation state of device (if true, device will get updates, if false, device will not get updates


```graphql
mutation{
  updateEdgeDevices(
    edgeDevices: [
    {
      id : "97f5ce24-8020-4479-d5c7-08daf328c601",
      deviceName: "New Device Name"
    }
  ],
  setNull: { deviceDescription : true })
  {
    id
  }
}
```

TODO: we need to give an example query for how to set the fleet of a device


## device-ping
send ping to a m2cp address and return uptime of the reached node
- sends a rpc to device via DeviceHub

```bash
$ m2cp device-ping d09b1153-f9ff-4c3d-97d2-46976208dd92
round trip:	876 ms
uptime:		48 days, 20:30:11
```

Query:

```graphql
query {
  result:devicePing(deviceSerial: "3e823a73-edb5-4334-7fab-08db464751da") {
    success
    errorMessage
    roundTripTime
    uptime
  }
}
```

When backend receives this query, it generates a rpc call to the device. This is done using the m2cp-SDK for dotnet.
The message is delivered via IoT Hub. 

This functionality **is already completely implemented** in the legacy service "DeviceHub" and can be taken from there.

## device-stats

get live system statistics of a device
- sends a rpc call to stat.logstat.<deviceserial>

TODO: this needs further discussion. Different implementations are possible:
1) backend offers a generic endpoint for sending rpc calls to devices. The developer can send any rpc call to any device - e.g. to get live statistics. This is the most flexible solution, but requires clients to know the rpc interface of the device. 
2) backend offers a specific endpoint for getting live statistics. The developer can query the backend for a list of available statistics and then query the backend for a specific statistic. This is less flexible, but easier to use for CLI/GUI developers.

Both solutions are possible, but we should decide for one of them.

NOTE: we want to have a generic endpoint for rpc _anyways_. So we can start with that and add a specific endpoint later.


## device-refresh
trigger refresh: install/remove snaps as defined by snap sets of device
- sends a rpc call to rpc.m2cp-gateway.<deviceserial>

```graphql
query {
  result:deviceRefresh(deviceSerial: "3e823a73-edb5-4334-7fab-08db464751da") {
    success
    errorMessage
  }
}
```

As a refresh call just triggers a device to start the automatic update process, the success of the operation cannot be determined immediately. A second call is needed to get the result of the refresh call:

```graphql
query {
  result:deviceRefreshLogs(deviceSerial: "3e823a73-edb5-4334-7fab-08db464751da") {
    success
      # list of log entries from device about successful and unsuccessful installation/uninstallation attempts  
    errorMessage
  }
}
```


## device-set-add
(deprecated - use device-modify to set fleet for a device)

## device-set-remove
(deprecated - use device-modify to set fleet for a device)

## device-ssh-open
establish reverse ssh tunnel to enable ssh connection to device
1) Backend configures ssh-rendevouz server for new connection
2) Backend sends rpc to device to connect to that rendevouz server
3) device will then update its list of users via snapstore and then connect to the rendevouz server
4) Backend sends connection details back to client in GQL response
5) Client connects to device using ssh via rendevouz server

```graphql
query {
  result:deviceSshOpen(deviceSerial: "3e823a73-edb5-4334-7fab-08db464751da") {
      success
      sshConnectionId
      sshConnectionString
      errorMessage
  }
}
```

The _sshConnectionString_ contains the connection details for the rendevouz server. The client can use this to connect to the device via ssh.




## device-ssh-close     
shut down reverse ssh tunnel for device
closes an existing ssh tunnel to a device (if one exists. also: tunnels time-out automatically)
1) Backend tells ssh-rendevouz server to delete connection
2) Backend sends rpc to device to close connection

```graphql
query {
  result:deviceSshOpen(deviceSerial: "3e823a73-edb5-4334-7fab-08db464751da", sshConnectionId: "1234567890") {
    success
    errorMessage
  }
}
```


## nodes
m2cp messaging nodes running on a device
- implemented in DeviceHub
- sends rpc to a device (to: rpc.m2cp-gateway.<deviceserial>) to ask for active nodes on device

TODO: this needs further discussion. See section _device-stats_ for details.

## node
information about a node in the m2cp messaging network
- implemented in DeviceHub
- sends rpc to a specific address (e.g. "sensorxy.<deviceserial>") and asks for rpc documentation

TODO: this needs further discussion. See section _device-stats_ for details.

## node-rpc
rpc command to node
- implemented in DeviceHub
- developer sends rpc-json to api, api creates rpc call, sends it to device, returns result to developer

TODO: this needs further discussion. See section _device-stats_ for details.


## user-logout

logout current developer
- analog

```
mutation{
   UserLogout{
      Success
      ErrorMessage
   }
}
```


## store-logs
(deprecated - we have a generic _logs_ command for all services and devices now)

## store-push
This is the most complex of all commands. We wrote a dedicated chapter about it in the wiki (TODO: add link here).

## m2cp store fleet list

List all fleets of currently selected tenant

```graphql
query ListAllExistingFleets{
  fleets{
    items{id,fleetName,description},
    totalCount,pageInfo{hasNextPage,hasPreviousPage}
  }
}

query ListAllFleetsForTenant($tenantId: UUID){
  fleets(where: {tenantId: {eq: $tenantId}}){
    items{id,fleetName,description},
    totalCount,
    pageInfo{hasNextPage,hasPreviousPage}
  }
}
```

## m2cp store fleet create

create a fleet, name must be unique for tenant (for tenant! not globally...)

```graphql
mutation ($fleets: [FleetCreateType!]!) {
  createFleets(fleets: $fleets){id}}
}

mutation CreateFleet($fleets: [FleetCreateType!]!){
  createFleets(fleets: $fleets)
  {
    id
    architecture
    fleetName
    fleetDescription
  }
}
```
with variables:
```json
{
  "fleets": [
    {
      "fleetName": "fleet1",
      "architecture": "amd64",
      "fleetDescription": "fleet1 description"
    }
  ]
}
```

## store fleet delete

remove a fleet - only possible, if fleet is not set in any device. 

```graphql
mutation ($fleetIds: [UUID!]!) {
  deleteFleets(ids: $fleetIds) {
    success
    errorMessage
  }
}
```
with variables:
```json
{
  "fleetIds": [
    "3e823a73-edb5-4334-7fab-08db464751da"
  ],
}
```


## store fleet info

get information on a fleet in your store
- list details, contents, linked devices, metadata of snap-set

```graphql
query {
  result:fleet(id: "ad742906-c626-4ca8-d5b3-08daf328c601") {
    id
    fleetName
    description
    edgeDevices{
      id
      deviceSerial
    }
    fleetBridgeSnapRevisions{
      snapRevision{
        id
        snapVersion
        revision
      }
    }
  }
}
```



## store fleet snap-add

Add a snap-revision to a fleet. By default, revisions will be selected
- add a snap of a specific revision to a fleet.
- constraints: fleet can't contain two revisions of same snap
- fleet _can_ contain a mix of amd64 and arm64 apps, they are filtered when installed on a device
- you can only add snaps to your fleet, when you own those apps _or_ if those apps come from mlpa (mlpa apps = operating system and essential tools)
- typical store has two tenants e.g. "knorr" and "mlpa"

```graphql 
mutation ($fleetId: ID!, $snapRevision: SnapRevisionInput!) {
  addSnapToFleet(fleetId: $fleetId, snapRevision: $snapRevision) {
    success
    errorMessage
  }
}
```
with variables:
```json
{
  "fleetId": "some-fleet-id",
  "snapRevision": {
    "id": "some-snap-revision-id"
  }
}
```

2023-09-20: We actually have:
```graphql 
mutation AddSnapRevisionToFleet($a: UUID!, $b: UUID!, $c: UUID!){
  createFleetBridgeSnapRevisions(fleetBridgeSnapRevisions: [
    {
      fleetBridgeSnapRevisionId: $a  # not nessessary, Alireza will fix that soon 2023-09-20
      fleetId: $b
      snapRevisionId: $c
    }
    ]
  ) {
    id
  }
}
```


## store fleet snap-modify

modify the revision of a snap in a fleet
the backend must make sure, that:both revisions belong to the same snap
- both revisions exist
- both revisions belong to the same snap
- both revisions have the same architecture

```graphql 
mutation ($fleetId: ID!, $oldSnapRevisionId: ID!, $newSnapRevision: SnapRevisionInput!) {
  modifySnapInFleet(fleetId: $fleetId, oldSnapRevisionId: $oldSnapRevisionId, newSnapRevisionId: $newSnapRevisionId) {
    success
    errorMessage
  }
}
```

with variables:
```json
{
  "fleetId": "some-fleet-id",
  "oldSnapRevisionId": "some-snap-revision-id",
  "newSnapRevisionId": "some-snap-revision-id"
}
```

2023-09-20: We actually have:
```graphql 
mutation ModifySnapRevisionOfFleet($a: UUID!, $b: UUID!, $c: UUID!){
  updateFleetBridgeSnapRevisions(fleetBridgeSnapRevisions: [
    {
      id: $a,
      fleetId: $b,
      snapRevisionId: $c
    }
  ]) {
    id
  }
}
```

## store fleet snap-remove

remove a snap-revision from a fleet
- remove a snap of a specific revision from a fleet by passing FleetBridgeSnapRevisionIds
    
```graphql
mutation{
  deleteFleetBridgeSnapRevisions(ids: ["421c5666-77fd-44f7-cd25-08daf3bc7a9c"]){
    id
  }
}
```

2023-09-20: We actually have:
```graphql 
mutation RemoveSnapRevisionFromFleet($a: UUID!){
  deleteFleetBridgeSnapRevisions(ids:
    [ $a ]
  ) {
    id
  }
}
```

## store snap list

- list all snaps (snap declarations = apps) available to you (that is all of your tenant plus the snaps of mlpa))

```graphql
query{
  result:snapDeclarations{
    items{
      id
      snapName
      snapDescription
    }
    pageInfo{
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```

## store snap info

get information on a snap in your store
- query db for details on that snap (list of revisions, etc)

```graphql
query{
  result:snapDeclaration(id: "3083c00d-fd56-4b1e-31ee-08dad3a6cabf")
  {
    id
    snapName
    snapDescription
    snapDeviceArchitecture
    snapRevisions{
      id
      revision
      createdAt
    }
  }
}
```

## store version

Return information on the backend:
* **Environment**: Value of `ASPNETCORE_ENVIRONMENT` environment variable.
* **Version**: Dotnet assembly version (This value will be provided by the customer's pipeline later).
* **CoreVersion**: Snapstore core assembly version (This value will be provider while `MLPA.Core.Snapstore` NuGet package deployment)
* **DatabaseVersion**: The ID of the last applied migration
* **CompatibleDatabaseVersion**: The ID of the last migration that is compatible with the currently running code
* **friendlyVersion**: $"Snapstore 3.0 - {version}"

If (DatabaseVersion <> CompatibleDatabaseVersion), it indicates that we have manually deployed code that could potentially be broken.

```graphql
query StoreVersion {
  snapStoreVersion {
    environment
    version
    coreVersion
    databaseVersion
    compatibleDatabaseVersion
    friendlyVersion
  }
}
```
