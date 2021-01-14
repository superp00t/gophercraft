package ngdp

type BuildConfig struct {
	Root                    Hash     `ccfg:"root"`
	Install                 []Hash   `ccfg:"install"`
	InstallSize             []uint64 `ccfg:"install-size"`
	Download                []Hash   `ccfg:"download"`
	DownloadSize            []uint64 `ccfg:"download-size"`
	Size                    []Hash   `ccfg:"size"`
	SizeSize                []uint64 `ccfg:"size-size"`
	Encoding                []Hash   `ccfg:"encoding"`
	EncodingSize            []uint64 `ccfg:"encoding-size"`
	Patch                   Hash     `ccfg:"patch"`
	PatchSize               uint64   `ccfg:"patch-size"`
	PatchConfig             Hash     `ccfg:"patch-config"`
	BuildName               string   `ccfg:"build-name"`
	BuildUID                string   `ccfg:"build-uid"`
	BuildProduct            string   `ccfg:"build-product"`
	BuildPlayBuildInstaller string   `ccfg:"build-playbuild-installer"`
	BuildPartialPriority    []string `ccfg:"build-partial-priority"`
}
