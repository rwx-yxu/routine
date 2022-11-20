package routine

import (
	"context"
	"fmt"
	"log"
	"net/http"

	Z "github.com/rwxrob/bonzai/z"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func GetClient(config *oauth2.Config) *http.Client {
	tok, err := TokenFromFile(config)
	if err != nil {
		tok = GetTokenFromWeb(config)
		SaveToken(tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func TokenFromFile(config *oauth2.Config) (*oauth2.Token, error) {
	//Get token from bonzai conf
	c, err := Z.Conf.Data()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = yaml.Unmarshal([]byte(c), tok)
	//Refresh token if expired
	if !tok.Valid() {
		src := config.TokenSource(context.Background(), tok)
		newToken, err := src.Token()
		if err != nil {
			return tok, err
		}
		//Check that the new token is different to expired token before
		//saving to bonzai conf
		if newToken.AccessToken != tok.AccessToken {
			SaveToken(newToken)
			return newToken, nil
		}
	}
	return tok, err
}

// Saves a token to bonzai config
func SaveToken(token *oauth2.Token) {
	Z.Conf.OverWrite(token)
}
