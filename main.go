package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gowebapp/controllers"
	"gowebapp/models"
	"log"
	"net/http"
	"text/template"
)

var tpl *template.Template
var Db *sql.DB

func init() {
	var err error
	if Db, err = sql.Open("postgres", "user=devtest dbname=gowebapp password=password sslmode=disable"); err != nil {
		panic(err)
	}

	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	// Controllers
	uc := controllers.NewUserController(Db)
	fc := controllers.NewFollowController(Db)
	tc := controllers.NewTweetController(Db)

	gmux := mux.NewRouter()

	// Registration Routes
	gmux.HandleFunc("/register", uc.New).Methods("GET")
	gmux.HandleFunc("/register", uc.Create).Methods("POST")
	gmux.HandleFunc("/delete-user", uc.Delete).Methods("POST")

	// Session Routes
	gmux.HandleFunc("/login", user_login).Methods("GET")
	gmux.HandleFunc("/login", login).Methods("POST")
	gmux.HandleFunc("/logout", logout).Methods("GET")

	// User Account Info Routes
	gmux.HandleFunc("/settings", uc.Edit).Methods("GET")
	gmux.HandleFunc("/update-user-info", uc.UpdateInfo).Methods("POST")
	gmux.HandleFunc("/update-user-meta", uc.UpdateMeta).Methods("POST")
	gmux.HandleFunc("/update-user-password", uc.UpdatePassword).Methods("POST")

	gmux.HandleFunc("/profile/{user_id}", uc.Show).Methods("GET")
	gmux.HandleFunc("/follow-user/{user_id}", fc.Create).Methods("GET")
	gmux.HandleFunc("/unfollow-user/{user_id}", fc.Delete).Methods("GET")

	/* Tweet Routes */
	gmux.HandleFunc("/create-tweet", tc.Create).Methods("POST")
	gmux.HandleFunc("/delete-tweet/{tweet_id}", tc.Delete).Methods("GET")
	gmux.HandleFunc("/create-retweet/{tweet_id}", tc.Retweet).Methods("GET")

	/* Tweet Favorite Routes */
	gmux.HandleFunc("/favorite/{tweet_id}", tc.Favorite).Methods("GET")
	gmux.HandleFunc("/unfavorite/{tweet_id}", tc.Unfavorite).Methods("GET")

	/* Base Routes */
	gmux.HandleFunc("/favicon.ico", handlerIcon).Methods("GET")
	gmux.HandleFunc("/", home).Methods("GET")

	fmt.Printf("About to listen on port :3000. Go to https://127.0.0.1:3000/ (localhost)\n")
	log.Fatal(http.ListenAndServe(":3000", gmux))
}

func handlerIcon(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("\nUser accessed the '%s' url path.\n", req.URL.Path)
}

func home(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("\nUser accessed the '%s' url path.\n", req.URL.Path)

	// Create map to pass data to template
	pageData := map[string]interface{}{
		"Title":      "Home",
		"BodyHeader": "Welcome to the Starting Block",
		"Paragraph":  "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor.",
	}

	user_id, err := req.Cookie("session_uid")

	if err == nil {
		pageData["isUserLoggedIn"] = true

		if foundTweets, userTweets := getTweets(user_id.Value); foundTweets == true {
			pageData["foundTweets"] = true
			pageData["Tweets"] = userTweets
		} else {
			pageData["foundTweets"] = false
		}

	} else {
		pageData["isUserLoggedIn"] = false
	}

	fmt.Printf("\n")
	fmt.Println(pageData)
	fmt.Printf("\n")

	tpl.ExecuteTemplate(res, "index.html", pageData)
}

func getTweets(user_id string) (bool, []Models.Tweet) {
	fmt.Printf("\nGetting Tweets for user $s :o\n", user_id)

	var foundTweets = true
	var tweets []Models.Tweet

	rows, err := Db.Query("select distinct (t.id), user_id, msg, name, username, is_retweet, origin_tweet_id, origin_user_id, origin_name, origin_username, t.created_at from tweets t inner join user_follows f on t.user_id = f.following_id where f.follower_id = $1 or t.user_id = $1 order by t.created_at desc", user_id)
	// old sql statement: select t.id, t.user_id, msg, t.created_at, count(*) from user_follows f left join tweets t on f.follower_id = $1 and f.following_id = t.user_id or t.user_id = $1 group by t.id order by t.created_at desc
	// select distinct (t.id), t.user_id, t.msg, u.name, u.username, t.is_retweet, t.origin_tweet_id, t.origin_user_id, t.origin_name, t.origin_username, t.created_at from tweets t inner join users u on t.user_id = u.id inner join user_follows f on t.user_id = f.following_id where f.follower_id = $1 or t.user_id = $1 order by t.created_at desc

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		tweet := Models.Tweet{}

		err := rows.Scan(&tweet.Id, &tweet.User_id, &tweet.Message, &tweet.Name, &tweet.Username, &tweet.Is_retweet, &tweet.Otweet_id, &tweet.Ouser_id, &tweet.Oname, &tweet.Ousername, &tweet.Created_at)

		switch {
		case err == sql.ErrNoRows:
			foundTweets = false
			log.Printf("\nNo tweets found.\n")

			return foundTweets, tweets
		case err != nil:
			log.Fatal(err)
		default:
			fmt.Printf("\nAdded a tweet.")
			tweets = append(tweets, tweet)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return foundTweets, tweets
}
