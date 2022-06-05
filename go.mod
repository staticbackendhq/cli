module github.com/staticbackendhq/cli

go 1.13

require (
	github.com/google/uuid v1.3.0
	github.com/gookit/color v1.2.2
	github.com/gorilla/websocket v1.4.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0
	github.com/staticbackendhq/backend-go v0.0.0-20201215215817-6e321a842def
	github.com/staticbackendhq/core v1.2.1
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa
)

replace github.com/staticbackendhq/backend-go => ../backend-go

replace github.com/staticbackendhq/core => ../core
