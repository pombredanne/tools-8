package config

// Provides configuration flags etc for ogload

import (
	"os"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func LoadConfig() {
	Ogload.Flags().String("server_root",  "/", "Root webserver directory")
	Ogload.Flags().String("static_files", ".", "Directory to serve static files from")


	Ogload.Flags().StringP("addr", "a", "127.0.0.1", "Address to listen on")
	Ogload.Flags().Int64P( "port", "p",        8080, "Port to listen on")

	Ogload.Flags().String("cert_file", "", "/path/to/tls.cert")
	Ogload.Flags().String("key_file",  "", "/path/to/tls.key")

	viper.BindPFlag("ServerRoot",  Ogload.Flags().Lookup("server_root"))
	viper.BindPFlag("StaticFiles", Ogload.Flags().Lookup("static_files"))

	viper.BindPFlag("ListenPort", Ogload.Flags().Lookup("port"))
	viper.BindPFlag("ListenAddr", Ogload.Flags().Lookup("addr"))

	viper.BindPFlag("CertFile", Ogload.Flags().Lookup("cert_file"))
	viper.BindPFlag("KeyFile", Ogload.Flags().Lookup("key_file"))

	viper.SetEnvPrefix("OGLOAD")
	viper.AutomaticEnv()
}

// Overkill, but whatever. options isn't exported, but it
// lives above the tearline so that it's easy to find 
var options = map[string]interface{}{
	// Basic config
	"ServerRoot":          "/", 
	"StaticFiles":         ".", // We'll watch for any files below here
	"ListenPort":         8080,
	"ListenAddr":  "127.0.0.1",

	// SSL Keys
	"CertFile":    "", // /path/to/ssl.cert
	"KeyFile":     "", // /path/to/ssl.key
}

// --------------------------- Internal --------------------------- //

func init() {
	setDefaults()

	Ogload.AddCommand(version)
	LoadConfig()
}

func setDefaults() {
	for o, v := range options {
		viper.SetDefault(o, v)
	}
}

var Ogload = &cobra.Command {
	Use:   os.Args[0], // Ensure we don't care what we're called
	Short: "Launch " + os.Args[0],
	Long:  "ogload serves and hot-reloads static files in the current working directory",
	Run: nil, // set in main.go for visibility's sake
}

// Version itself is maintained via the makefile on build
var version = &cobra.Command {
	Use:   "version",
	Short: "Print the current version",
	Run:   func(cmd *cobra.Command, args[] string) {
		fmt.Println(Version())
	},
}