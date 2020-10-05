package src

type Command interface {
	Help() string
	Run(config *Config, args []string) error
}
