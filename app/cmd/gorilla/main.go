package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/aws/aws-sdk-go-v2/config"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gosimple/slug"

	"github.com/rs/cors"
	"technicarium.com/api/app/pkg/recipes"
)

type CognitoClient struct {
	AppClientId string
	*cip.Client
}

const portNumber = ":8080"

type RecipesHandler struct {
	store recipeStore
}

func NewRecipesHandler(s recipeStore) *RecipesHandler {
	return &RecipesHandler{
		store: s,
	}
}

type homeHandler struct{}

func (h *homeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is my home page"))
}

func Init() *CognitoClient {
	// Load the shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	cip.NewFromConfig(cfg)
	return &CognitoClient{
		"",
		cip.NewFromConfig(cfg),
	}
}

func main() {

	arr := []any{"a", "b", "c"}
	test(arr)

	// Create the Store and Recipe Handler
	store := recipes.NewMemStore()
	recipesHandler := NewRecipesHandler(store)
	home := homeHandler{}

	// Create the router
	router := mux.NewRouter()
	//cors optionsGoes Below

	// Register the routes
	// router.Use(accessControlMiddleware)
	router.HandleFunc("/", home.ServeHTTP)
	router.HandleFunc("/recipes", recipesHandler.ListRecipes).Methods("GET")
	router.HandleFunc("/recipes", recipesHandler.CreateRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}", recipesHandler.GetRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}", recipesHandler.UpdateRecipe).Methods("PUT")
	router.HandleFunc("/recipes/{id}", recipesHandler.DeleteRecipe).Methods("DELETE")
	router.HandleFunc("/api/v1/scheduler", recipesHandler.Scheduler).Methods("GET")
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		// AllowCredentials: true,
		AllowedMethods: []string{
			http.MethodGet, //http methods for your app
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},

		AllowedHeaders: []string{
			"*", //or you can your header key values which you are using in your application
		},
	})
	handler := c.Handler(router)

	// Start the server
	http.ListenAndServe(portNumber, handler)
}

type recipeStore interface {
	Add(name string, recipe recipes.Recipe) error
	Get(name string) (recipes.Recipe, error)
	List() (map[string]recipes.Recipe, error)
	Update(name string, recipe recipes.Recipe) error
	Remove(name string) error
}

func (h RecipesHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {

	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe

	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	resourceID := slug.Make(recipe.Name)
	if err := h.store.Add(resourceID, recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h RecipesHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {

	recipesList, err := h.store.List()
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	jsonBytes, err := json.Marshal(recipesList)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h RecipesHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	recipe, err := h.store.Get(id)
	if err != nil {
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}
		InternalServerErrorHandler(w, r)
		return
	}
	jsonBytes, err := json.Marshal(recipe)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h RecipesHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Recipe object that will be populated from JSON payload
	var recipe recipes.Recipe

	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	if err := h.store.Update(id, recipe); err != nil {
		if err == recipes.NotFoundErr {
			NotFoundHandler(w, r)
			return
		}

		InternalServerErrorHandler(w, r)
		return
	}

	jsonBytes, err := json.Marshal(recipe)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h RecipesHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.store.Remove(id); err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h RecipesHandler) Scheduler(w http.ResponseWriter, r *http.Request) {
	// enableCors(&w)
	log.Printf("Scheduler")
	reqToken := r.Header.Get("Authorization")
	// TODO - verify
	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]
	log.Printf("Token: %s", reqToken)
	// Get the JWK Set URL from your AWS region and userPoolId.
	//
	// See the AWS docs here:
	// https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html
	regionID := "eu-central-1"             // TODO Get the region ID for your AWS Cognito instance.
	userPoolID := "eu-central-1_RDs64szop" // TODO Get the user pool ID of your AWS Cognito instance.
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, userPoolID)

	// Create the keyfunc.Keyfunc.
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("Failed to create JWK Set from resource at the given URL.\nError: %s", err)
	}

	// Get a JWT to parse.
	// jwtB64 := // base 64 encoded JWT
	jwtB64 := reqToken

	// Parse the JWT.
	token, err := jwt.Parse(jwtB64, jwks.Keyfunc)
	if err != nil {
		// log.Fatalf("Failed to parse the JWT.\nError: %s", err)
		log.Printf("Failed to parse the JWT.\nError: %s", err)
	}

	// Check if the token is valid.
	if !token.Valid {
		// log.Fatalf("The token is not valid.")
		log.Printf("The token is not valid.")
	} else {
		log.Println("The token is valid.")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("bla bla"))
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 Not Found"))
}

// func enableCors(w *http.ResponseWriter) {
// 	log.Println("Enable cors.")
// 	(*w).Header().Set("Access-Control-Allow-Origin", "*")
// }

// access control and  CORS middleware
func accessControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func test(...any) {
}
