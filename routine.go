package routine

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/rwx-yxu/esc-pos/seq"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/emb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

func Print(e *calendar.Events, name string) {
	fileName := "data.txt"

	f, err := os.Create(fileName)

	if err != nil {

		log.Fatal(err)
	}

	defer f.Close()

	words := []string{}
	words = append(words, seq.Default)
	//Header
	words = append(words, fmt.Sprintf("%v%v%s\n", seq.Center, seq.CharSize(1), "Routine"))
	words = append(words, fmt.Sprintf("%v%15v%v", seq.UL, " ", seq.ULOff))

	for _, i := range e.Items {
		s := i.Start.DateTime
		if s == "" {
			s = i.Start.Date
		}
		sTime, err := time.Parse(time.RFC3339, s)
		if err != nil {
			log.Println(err)
			continue
		}
		e := i.End.DateTime
		eTime, err := time.Parse(time.RFC3339, e)
		words = append(words, fmt.Sprintf("%v%v", seq.Default, seq.GS+"\x21\x01\x03"))
		words = append(words, fmt.Sprintf("\n%v%s %s (%s-%s)\n", seq.Left, i.Summary, sTime, eTime))
		words = append(words, fmt.Sprintf("%s", i.Description))
	}
	words = append(words, fmt.Sprintf("%v%47v%v\n", seq.UL, " ", seq.ULOff))

	words = append(words, seq.Cut)
	for _, word := range words {

		_, err := f.WriteString(word)

		if err != nil {
			log.Fatal(err)
		}
	}

	cmd := exec.Command("lp", "-d", name, fileName)

	err = cmd.Run()
	if err != nil {
		log.Println("cmd err")
		log.Println(err)
	}
}

func GetService() (*calendar.Service, error) {
	ctx := context.Background()

	jsn, _ := emb.Cat("credentials.json")
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(jsn, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := GetClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))

	return srv, err

}

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
