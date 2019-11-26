package gui

type Config struct {
	ConfigFile string
	Log        LogConfig     `yaml:"log"`
	Preview    PreviewConfig `yaml:"preview"`
	IgnoreCase bool          `yaml:"ignore_case"`
}

type LogConfig struct {
	Enable bool   `yaml:"enable"`
	File   string `yaml:"file"`
}

type PreviewConfig struct {
	Enable      bool   `yaml:"enable"`
	Colorscheme string `yaml:"colorscheme"`
}

func DefaultConfig() Config {
	return Config{
		Log: LogConfig{
			Enable: false,
		},
		Preview: PreviewConfig{
			Enable:      false,
			Colorscheme: "monokai",
		},
		IgnoreCase: false,
	}
}
