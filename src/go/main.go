package main

import (
	"context"
	"m2cpcli/cmd"
	_ "m2cpcli/cmd/asset"
	_ "m2cpcli/cmd/asset/model"
	_ "m2cpcli/cmd/asset/system"
	_ "m2cpcli/cmd/device" // see also https://stackoverflow.com/a/73986980
	_ "m2cpcli/cmd/device/coap"
	_ "m2cpcli/cmd/device/dgroup"
	_ "m2cpcli/cmd/device/message"
	_ "m2cpcli/cmd/device/node"
	_ "m2cpcli/cmd/device/snap"
	_ "m2cpcli/cmd/device/ssh"
	_ "m2cpcli/cmd/device/uptime"
	_ "m2cpcli/cmd/device/virtual"
	_ "m2cpcli/cmd/firmware/manifest"
	_ "m2cpcli/cmd/image"
	_ "m2cpcli/cmd/store"
	_ "m2cpcli/cmd/store/app"
	_ "m2cpcli/cmd/store/app/rate"
	_ "m2cpcli/cmd/store/dgroup"
	_ "m2cpcli/cmd/store/dgroup/admin"
	_ "m2cpcli/cmd/store/dgroup/app"
	_ "m2cpcli/cmd/store/model"
	_ "m2cpcli/cmd/store/snap"
	_ "m2cpcli/cmd/store/snap/rate"
	_ "m2cpcli/cmd/store/system"
	_ "m2cpcli/cmd/store/system/model"
	_ "m2cpcli/cmd/store/system/snapdseed"
	_ "m2cpcli/cmd/user"
	_ "m2cpcli/cmd/user/roles"
	_ "m2cpcli/cmd/user/tenant"
)

func main() {
	ctx := context.Background()
	cmd.Execute(ctx)
}
