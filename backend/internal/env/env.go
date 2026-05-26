package env

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/Ethanol2/book-organizer/internal/auth"
	"github.com/joho/godotenv"
)

const (
	// User editable
	GoogleApiKey  = "GOOGLE_BOOKS_API_KEY"
	Port          = "PORT"
	ScanFrequency = "DOWNLOADS_SCAN_FREQUENCY"

	// Hidden
	DataPath      = "DATA_PATH"
	DownloadsPath = "DOWNLOADS_PATH"
	LibraryPath   = "LIBRARY_PATH"
	TokenSecret   = "TOKEN_SECRET"
)

type Env struct {
	mu     *sync.Mutex
	values map[string]string
	path   string
}

func getVarWithFallback(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		fmt.Println(val)
		return val
	}
	return fallback
}

func GetEnv() (Env, error) {

	if _, err := os.Stat(".env"); err == nil {
		// If it exists, load it. This fills up os.Getenv with your master values!
		if err := godotenv.Load(); err != nil {
			return Env{}, err
		}
		fmt.Println("Loaded master .env file successfully")
	}

	envPath := "./data/.env"

	err := os.MkdirAll(path.Dir(envPath), 0755)
	if err != nil {
		return Env{}, err
	}

	if _, err := os.Stat(envPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Initializing env file")
			initialValues := map[string]string{
				Port:          getVarWithFallback(Port, "8080"),
				TokenSecret:   getVarWithFallback(TokenSecret, ""),
				DataPath:      getVarWithFallback(DataPath, "./data"),
				DownloadsPath: getVarWithFallback(DownloadsPath, "./downloads"),
				LibraryPath:   getVarWithFallback(LibraryPath, "./library"),
				GoogleApiKey:  getVarWithFallback(GoogleApiKey, ""),
			}

			if err := godotenv.Write(initialValues, envPath); err != nil {
				return Env{}, err
			}
		} else {
			return Env{}, err
		}
	}

	if err := godotenv.Overload(envPath); err != nil {
		return Env{}, err
	}

	values, err := godotenv.Read(envPath)
	if err != nil {
		return Env{}, err
	}

	needsUpdate := false

	dataPath := ""
	if path, ok := values[DataPath]; !ok || path == "" {
		fmt.Println("No data path defined in the .env. Setting value to \"./data/\"")
		dataPath = "./data/"
		values[DataPath] = dataPath
		needsUpdate = true
	} else {
		dataPath = path
	}

	if dPath, ok := values[DownloadsPath]; !ok || dPath == "" {
		fmt.Println("No downloads path defined in the .env. Setting value to \"./downloads/\"")
		values[DownloadsPath] = "./downloads/"
		needsUpdate = true
	}

	if lPath, ok := values[LibraryPath]; !ok || lPath == "" {
		fmt.Println("No library path defined in the .env. Setting value to \"./library\"")
		values[LibraryPath] = "./library/"
		needsUpdate = true
	}

	if port, ok := values[Port]; !ok || port == "" {
		fmt.Println("No port defined in the .env. Setting value to 8080")
		values[Port] = "8080"
		needsUpdate = true
	}

	if key, ok := values[GoogleApiKey]; !ok || key == "" {
		fmt.Println("No google books api key defined in the .env. Google Books will not work without it. This can also be set via the settings")
	}

	if secret, ok := values[TokenSecret]; !ok || secret == "" {
		fmt.Println("No secret key defined in the .env. Applying a newly generated key.")
		values[TokenSecret] = auth.MakeToken()
		needsUpdate = true
	}

	if fStr, ok := values[ScanFrequency]; !ok || fStr == "" {
		fmt.Println("No scan frequency was defined in the .env. Setting value to 5 seconds")
		values[ScanFrequency] = "5"
		needsUpdate = true
	} else {
		seconds, err := strconv.Atoi(fStr)
		if err != nil || seconds < 1 {
			log.Fatal("scan frequency must be a positive integer greater than 0")
		}
	}

	if needsUpdate {
		err = godotenv.Write(values, envPath)
		if err != nil {
			return Env{}, err
		}
	}

	return Env{
		&sync.Mutex{},
		values,
		envPath,
	}, nil
}

func (env *Env) Get(key string) string {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()
	return env.values[key]
}

func (env *Env) GetInt(key string) int {
	str := env.Get(key)
	if str == "" {
		log.Fatalf("Number \"%s\" requested from the env variables, but nothing exists", key)
	}

	num, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalf("Number \"%s\" requested from the env variables, but it's an invalid value: %v", key, err)
	}

	return num
}

func (env *Env) GetMultiple(keys ...string) map[string]string {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()

	values := map[string]string{}
	for _, key := range keys {
		values[key] = env.values[key]
	}

	return values
}

func (env *Env) Set(key, value string) error {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()

	env.values[key] = value
	return godotenv.Write(env.values, env.path)
}

func (env *Env) SetMultiple(values map[string]string) error {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()

	for key, value := range values {
		env.values[key] = value
	}

	return godotenv.Write(env.values, env.path)
}
