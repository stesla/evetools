package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stesla/evetools/config"
	"github.com/stesla/evetools/esi"
	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
)

var cfgFile string

var rootCmd = &cobra.Command{Use: "evetools-cli"}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./evetools.yaml", "config file")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")
	viper.BindPFlag("cli.quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.SetDefault("cli.token", "./evetools.token")

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

	// market
	marketCmd := &cobra.Command{Use: "market"}
	rootCmd.AddCommand(marketCmd)

	cmd = &cobra.Command{
		Use: "fetchPrices",
		Run: fetchPricesCommand,
	}
	marketCmd.AddCommand(cmd)
}

func initConfig() {
	if err := config.Initialize(cfgFile); err != nil {
		log.Fatalln("error initializing config:", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}

func getTokenCmd(cmd *cobra.Command, args []string) {
	fmt.Printf("Go to this URL in your browser:\n\n%s\n", config.OAuthForCLI.AuthCodeURL("evetools"))
}

func refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	oldTok := oauth2.Token{RefreshToken: refreshToken}
	tokSrc := config.OAuthForCLI.TokenSource(ctx, &oldTok)
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
		log.Fatalln("error refreshing token:", err)
	}
	if err = saveToken(token.RefreshToken); err != nil {
		log.Fatalln("error saving token:", err)
	}
}

func refreshTokenCmd(cmd *cobra.Command, args []string) {
	token, err := getToken()
	if err != nil {
		log.Fatalln("error loading token:", err)
	}
	json.NewEncoder(os.Stdout).Encode(&token)
}

func listStructuresCmd(cmd *cobra.Command, args []string) {
	token, err := getToken()
	if err != nil {
		log.Fatalln("error loading token:", err)
	}

	var client http.Client
	eclient := esi.NewClient(&client)

	ids, err := eclient.GetStructures()
	if err != nil {
		log.Fatalln("error fetching structures:", err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, esi.AccessTokenKey, token.AccessToken)
	encoder := json.NewEncoder(os.Stdout)
	for _, id := range ids {
		station, err := eclient.GetStructure(ctx, id)
		if err != nil {
			log.Fatalln("error fetching structure:", err)
		}
		station.ID = id
		system := sde.SolarSystems[station.SystemID]
		station.RegionID = system.RegionID
		encoder.Encode(&station)
	}
}

func fetchPricesCommand(cmd *cobra.Command, args []string) {
	db, err := model.Initialize(config.DatabaseFile())
	if err != nil {
		log.Fatalln("error initializing model:", err)
	}

	stations, err := db.AllUserStations()
	if err != nil {
		log.Fatalln("error fetching stations:", err)
	}

	var client http.Client
	eclient := esi.NewClient(&client)

	for _, s := range stations {
		if !viper.GetBool("cli.quiet") {
			log.Println("Fetching", s.Name)
		}
		prices, err := eclient.GetMarketPrices(s.ID, s.RegionID)
		if err != nil {
			log.Fatalln("error fetching prices:", err)
		}

		for id, price := range prices {
			if err := db.SavePrice(s.ID, id, price); err != nil {
				log.Fatalln(err)
			}
		}
	}
}
