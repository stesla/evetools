package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/sde"
)

var (
	cfgFile string
	token   string
)

var rootCmd = &cobra.Command{Use: "evetools-cli"}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./evetools.yaml", "config file")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "refresh token to save")
	viper.SetDefault("cli.token", "./evetools.token")
	viper.SetDefault("esi.basePath", "https://esi.evetech.net")
	viper.SetDefault("oauth.basePath", "https://login.eveonline.com")

	var cmd *cobra.Command

	// token management
	tokenCmd := &cobra.Command{Use: "token"}
	rootCmd.AddCommand(tokenCmd)

	cmd = &cobra.Command{
		Use: "get",
		Run: getTokenCmd,
	}
	tokenCmd.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:  "save TOKEN",
		Args: cobra.ExactArgs(1),
		Run:  saveTokenCmd,
	}
	tokenCmd.AddCommand(cmd)

	cmd = &cobra.Command{
		Use: "refresh",
		Run: refreshTokenCmd,
	}
	tokenCmd.AddCommand(cmd)

	// structures
	structuresCmd := &cobra.Command{Use: "structures"}
	rootCmd.AddCommand(structuresCmd)

	cmd = &cobra.Command{
		Use: "list",
		Run: listStructuresCmd,
	}
	structuresCmd.AddCommand(cmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		die("must provide config file")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		die("error reading config: %v", err)
	}

	if err := initOAuthConfig(); err != nil {
		die(err.Error())
	}
}

func die(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		die(err.Error())
	}
}

var oauthConfig = oauth2.Config{
	Scopes: []string{
		"esi-search.search_structures.v1",
		"esi-universe.read_structures.v1",
		"esi-markets.structure_markets.v1",
	},
}

func initOAuthConfig() error {
	oauthConfig.ClientID = viper.GetString("cli.oauth.clientID")
	if oauthConfig.ClientID == "" {
		return fmt.Errorf("must provide cli.oauth.clientID")
	}
	oauthConfig.ClientSecret = viper.GetString("cli.oauth.clientSecret")
	if oauthConfig.ClientSecret == "" {
		return fmt.Errorf("must provide cli.oauth.clientSecret")
	}
	oauthConfig.RedirectURL = viper.GetString("cli.oauth.redirectURL")
	if oauthConfig.RedirectURL == "" {
		return fmt.Errorf("must provide cli.oauth.redirectURL")
	}
	basePath := viper.GetString("oauth.basePath")
	oauthConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/v2/oauth/authorize", basePath),
		TokenURL: fmt.Sprintf("%s/v2/oauth/token", basePath),
	}
	return nil
}

func getTokenCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("Go to this URL in your browser:\n\n%s\n", oauthConfig.AuthCodeURL("evetools"))
}

func refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	oldTok := oauth2.Token{RefreshToken: refreshToken}
	tokSrc := oauthConfig.TokenSource(ctx, &oldTok)
	token, err := tokSrc.Token()
	if err != nil {
		return nil, err
	}
	return token, nil
}

func getToken() (*oauth2.Token, error) {
	token, err := loadToken()
	if err != nil {
		return nil, err
	}
	err = saveToken(token.RefreshToken)
	return token, err
}

func loadToken() (*oauth2.Token, error) {
	input, err := os.Open(viper.GetString("cli.token"))
	if err != nil {
		return nil, err
	}
	defer input.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(input); err != nil {
		return nil, err
	}
	return refreshToken(context.Background(), buf.String())
}

func saveToken(token string) error {
	output, err := os.Create(viper.GetString("cli.token"))
	if err != nil {
		return err
	}
	defer output.Close()
	fmt.Fprintln(output, token)
	return nil
}

func saveTokenCmd(cmd *cobra.Command, args []string) {
	token, err := refreshToken(context.Background(), args[0])
	if err != nil {
		die("error refreshing token: %v", err)
	}
	if err = saveToken(token.RefreshToken); err != nil {
		die("error saving token: %v", err)
	}
}

func refreshTokenCmd(cmd *cobra.Command, args []string) {
	token, err := getToken()
	if err != nil {
		die("error loading token:", err)
	}
	json.NewEncoder(os.Stdout).Encode(&token)
}

func listStructuresCmd(cmd *cobra.Command, args []string) {
	token, err := getToken()
	if err != nil {
		die("error loading token: %v", err)
	}

	var client http.Client
	eclient := esi.NewClient(&client)

	ids, err := eclient.GetStructures()
	if err != nil {
		die("error fetching structures: %v", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, esi.AccessTokenKey, token.AccessToken)
	encoder := json.NewEncoder(os.Stdout)
	for _, id := range ids {
		station, err := eclient.GetStructure(ctx, id)
		if err != nil {
			die("error fetching structure:", err)
		}
		station.ID = id
		system := sde.SolarSystems[station.SystemID]
		station.RegionID = system.RegionID
		encoder.Encode(&station)
	}
}
