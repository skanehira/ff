package gui

type LogConfig struct {
	Enable bool   `yaml:"enable"`
	File   string `yaml:"file"`
}

type PreviewConfig struct {
	Enable      bool   `yaml:"enable"`
	Colorscheme string `yaml:"colorscheme"`
}

type BookmarkConfig struct {
	Enable bool   `yaml:"enable"`
	File   string `yaml:"file"`
	Log    bool   `yaml:"log"`
}

type Config struct {
	ConfigDir  string
	ConfigFile string
	Log        LogConfig      `yaml:"log"`
	Preview    PreviewConfig  `yaml:"preview"`
	Bookmark   BookmarkConfig `yaml:"bookmark"`
	IgnoreCase bool           `yaml:"ignore_case"`
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
		Bookmark: BookmarkConfig{
			Enable: false,
			Log:    false,
		},
		IgnoreCase: false,
	}
}
