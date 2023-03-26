package cmd

type Config struct {
	SshAddress         string `mapstructure:"address"`
	SshUsername        string `mapstructure:"username"`
	SshPassword        string `mapstructure:"password"`
	SshRootPath        string `mapstructure:"root"`
	PrivateKeyFilePath string `mapstructure:"private-key"`
	LocalServePort     string `mapstructure:"serve-port"`
	MountOptions       string `mapstructure:"mount-options"`
	MountPath          string `mapstructure:"mount-path"`
}
