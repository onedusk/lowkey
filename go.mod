module lowkey

go 1.22

require (
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.17.0
)

replace github.com/spf13/cobra => ./vendor/github.com/spf13/cobra

replace github.com/spf13/viper => ./vendor/github.com/spf13/viper
