package config

type Config struct {
	Server     Server      `mapstructure:"server"`
	Log        Log         `mapstructure:"logging"`
	Database   Database    `mapstructure:"database"`
	PostgreSQL *PostgreSQL `mapstructure:"postgres"`
	Postgre    *PostgreSQL `mapstructure:"postgre"`
	MySQL      *MySQL      `mapstructure:"mysql"`
	SQLite3    *SQLite3    `mapstructure:"sqlite3"`
	Sonic      Sonic       `mapstructure:"sonic"`
	AI         AI          `mapstructure:"ai"`
}

type AI struct {
	// APIKey is the Anthropic API key. Required to enable AI features.
	APIKey string `mapstructure:"api_key"`
	// Model is the Claude model ID. Defaults to claude-haiku-4-5 if empty.
	Model string `mapstructure:"model"`
}

type Database struct {
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	ConnMaxIdleHour int `mapstructure:"conn_max_idle_hour"`
}

type PostgreSQL struct {
	Dsn      string `mapstructure:"dsn"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DB       string `mapstructure:"db"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`
}

type MySQL struct {
	Dsn string `mapstructure:"dsn"`
}

type SQLite3 struct {
	Enable bool `mapstructure:"enable"`
	File   string
}
type Server struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}
type Log struct {
	FileName string `mapstructure:"filename"`
	Levels   Levels `mapstructure:"level"`
	MaxSize  int    `mapstructure:"maxsize"`
	MaxAge   int    `mapstructure:"maxage"`
	Compress bool   `mapstructure:"compress"`
}
type Levels struct {
	App  string `mapstructure:"app"`
	Gorm string `mapstructure:"gorm"`
}

type LogMode string

const (
	Console LogMode = "console"
	File    LogMode = "file"
)

type Sonic struct {
	Mode              string  `mapstructure:"mode"`
	LogMode           LogMode `mapstructure:"log_mode"`
	WorkDir           string  `mapstructure:"work_dir"`
	UploadDir         string
	LogDir            string `mapstructure:"log_dir"`
	TemplateDir       string `mapstructure:"template_dir"`
	ThemeDir          string
	AdminResourcesDir string
	AdminURLPath      string `mapstructure:"admin_url_path"`
}

func (c *Config) GetPostgreSQL() *PostgreSQL {
	if c.PostgreSQL != nil {
		return c.PostgreSQL
	}
	return c.Postgre
}
