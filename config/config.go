package config

import (
	"fmt"

	"golang.org/x/oauth2"

	"github.com/spf13/viper"
)

func Initialize(cfgFile string) error {
	viper.SetDefault("esi.basePath", "https://esi.evetech.net")
	viper.SetDefault("http.cache.dir", "./cache")
	viper.SetDefault("model.database", "./evetools.sqlite3")
	viper.SetDefault("oauth.basePath", "https://login.eveonline.com")
	viper.SetDefault("sde.dir", "./data")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".evetools")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME")
	}
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error loading config file: %s", err)
	}

	if err := InitializeOAuth(); err != nil {
		return fmt.Errorf("error initializing oauth: %s", err)
	}

	return nil
}

var OAuthForHTTP = oauth2.Config{
	Scopes: []string{
		"esi-markets.read_character_orders.v1",
		"esi-ui.open_window.v1",
		"esi-characters.read_standings.v1",
		"esi-skills.read_skills.v1",
		"esi-wallet.read_character_wallet.v1",
		"publicData",
	},
}

var OAuthForCLI = oauth2.Config{
	Scopes: []string{
		"esi-search.search_structures.v1",
		"esi-universe.read_structures.v1",
		"esi-markets.structure_markets.v1",
	},
}

func InitializeOAuth() error {
	OAuthForHTTP.ClientID = viper.GetString("oauth.clientID")
	if OAuthForHTTP.ClientID == "" {
		return fmt.Errorf("must provide oauth.clientID")
	}
	OAuthForHTTP.ClientSecret = viper.GetString("oauth.clientSecret")
	if OAuthForHTTP.ClientSecret == "" {
		return fmt.Errorf("must provide oauth.clientSecret")
	}
	OAuthForHTTP.RedirectURL = viper.GetString("oauth.redirectURL")
	if OAuthForHTTP.RedirectURL == "" {
		return fmt.Errorf("must provide oauth.redirectURL")
	}
	OAuthForCLI.ClientID = viper.GetString("cli.oauth.clientID")
	if OAuthForCLI.ClientID == "" {
		return fmt.Errorf("must provide cli.oauth.clientID")
	}
	OAuthForCLI.ClientSecret = viper.GetString("cli.oauth.clientSecret")
	if OAuthForCLI.ClientSecret == "" {
		return fmt.Errorf("must provide cli.oauth.clientSecret")
	}
	OAuthForCLI.RedirectURL = viper.GetString("cli.oauth.redirectURL")
	if OAuthForCLI.RedirectURL == "" {
		return fmt.Errorf("must provide cli.oauth.redirectURL")
	}
	basePath := viper.GetString("oauth.basePath")
	endpoint := oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/v2/oauth/authorize", basePath),
		TokenURL: fmt.Sprintf("%s/v2/oauth/token", basePath),
	}
	OAuthForHTTP.Endpoint = endpoint
	OAuthForCLI.Endpoint = endpoint
	return nil
}

func CacheDir() string {
	return viper.GetString("http.cache.dir")
}

func DatabaseFile() string {
	return viper.GetString("model.database")
}

func EsiBasePath() string {
	return viper.GetString("esi.basePath")
}
