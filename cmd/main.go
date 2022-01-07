package main

import bilicoin "github.com/r3inbowari/zupdate"

func main() {
	up := bilicoin.InitUpdater(bilicoin.UpdateOptions{
		EntryName:   "hola",
		Mode:        bilicoin.DEV,
		CheckSource: "http://120.77.33.188:3000/bilicoin/bin/test.json",
		Callback:    nil,
		EntryArgs:   []string{"-a"},
	})

	up.IncludeFile("5c74bf9c1face2dcb47bae100f2c664cdbd43407", bilicoin.File{
		Name:  "hola",
		Major: 1,
		Minor: 0,
		Patch: 1,
	})

	up.IncludeFile("5c74bf9c1face2dcb47bae100f2c664cdbd43400", bilicoin.File{
		Name:  "abs.dll",
		Major: 1,
		Minor: 0,
		Patch: 1,
	})

	// 1.
	up.CheckAndUpdateWithGap()

	// 2.
	//up.Update(&bilicoin.File{
	//	Major:          1,
	//	Minor:          0,
	//	Patch:          1,
	//	Digest:         "80f784885493ddb5eda3435cd20ad488",
	//	DownloadSource: "http://120.77.33.188:3000/bilicoin/bin/bilicoin_windows_amd64_v1.0.11.exe",
	//	Reload:         true,
	//	EntryName:      "hola.exe",
	//})
	//up.Reload()
}
