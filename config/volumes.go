package config

type VolumeClassConfigs map[string]VolumeClassConfig

type VolumeClassConfig struct {
	Retain              int
	DeviceName          string
	Propagate           bool
	VolumeSize          int
	DeleteOnTermination bool
	MountPoint          string
}

func DefaultVolumeClasses() VolumeClassConfigs {
	defaultVolumes := make(VolumeClassConfigs)

	defaultVolumes["git"] = VolumeClassConfig{
		Retain:              5,
		DeviceName:          "/dev/xvdf",
		Propagate:           true,
		VolumeSize:          30,
		DeleteOnTermination: true,
		MountPoint:          "/mnt/git",
	}

	return defaultVolumes
}
