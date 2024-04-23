package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var BasePath = "/srv/data"

func getPropertiesMap() (map[string]string, error) {
	props, err := ReadProperties(BasePath + "/server.properties")
	if err != nil {
		log.Print("Error reading properties file:", err)
		return make(map[string]string), err
	}
	allProps := make(map[string]string)
	for key, value := range props {
		allProps[key] = value
	}
	return allProps, nil
}

func handleProperties(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		getProps(w, req)
	case "POST":
		modifyProps(w, req)
	}
}

func getProps(w http.ResponseWriter, req *http.Request) {
	props, err := getPropertiesMap()
	if err != nil {
		log.Print("Error reading properties file:", err)
		http.Error(w, "Error reading properties file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(props)
	if err != nil {
		log.Print("Error encoding properties to JSON:", err)
		http.Error(w, "Error encoding properties to JSON", http.StatusInternalServerError)
		return
	}

}

func modifyProps(w http.ResponseWriter, req *http.Request) {
	type prop struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	var p prop
	err := json.NewDecoder(req.Body).Decode(&p)
	if err != nil {
		log.Print("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	props, err := ReadProperties(BasePath + "/server.properties")
	if err != nil {
		log.Print("Error reading properties file:", err)
		http.Error(w, "Error reading properties file", http.StatusInternalServerError)
		return
	}

	props[p.Key] = p.Value
	err = OverwriteProperties(BasePath+"/server.properties", props)
	if err != nil {
		log.Print("Error writing properties file:", err)
		http.Error(w, "Error writing properties file", http.StatusInternalServerError)
		return
	}

	newProps, err := getPropertiesMap()
	if err != nil {
		log.Print("Error reading properties file:", err)
		http.Error(w, "Error reading properties file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(newProps)
	if err != nil {
		log.Print("Error encoding properties to JSON:", err)
		http.Error(w, "Error encoding properties to JSON", http.StatusInternalServerError)
		return
	}
}

func handleMods(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		getMods(w, req)
	case "POST":
		addMod(w, req)
	case "DELETE":
		deleteMod(w, req)
	}
}

func getModFilenames() ([]string, error) {
	files, err := os.ReadDir(BasePath + "/mods/")
	if err != nil {
		return nil, err
	}

	var mods []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		mods = append(mods, file.Name())
	}
	return mods, nil
}

func getMods(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mods, err := getModFilenames()
	if err != nil {
		log.Print("Error getting list of mods:", err)
		http.Error(w, "Error getting list of mods", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(mods)
	if err != nil {
		log.Print("Error encoding list of mods to JSON:", err)
		http.Error(w, "Error encoding list of mods to JSON", http.StatusInternalServerError)
		return
	}
}

func addMod(w http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(100 << 20)
	if err != nil {
		log.Print("Error parsing file:", err)

		http.Error(w, "Error parsing file", http.StatusBadRequest)
		return
	}

	file, handler, err := req.FormFile("file")
	if err != nil {
		log.Print("Error parsing file:", err)
		http.Error(w, "Error parsing file", http.StatusBadRequest)
		return

	}
	defer file.Close()

	dst, err := os.Create(BasePath + "/mods/" + handler.Filename)
	if err != nil {
		log.Print("Error creating file:", err)
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		log.Print("Error saving file:", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	modlist, err := getModFilenames()
	if err != nil {
		log.Print("Error getting list of mods:", err)
		http.Error(w, "Error getting list of mods", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(modlist)
	if err != nil {
		log.Print("Error encoding list of mods to JSON:", err)
		http.Error(w, "Error encoding list of mods to JSON", http.StatusInternalServerError)
		return
	}

}

func deleteMod(w http.ResponseWriter, req *http.Request) {
	type mod struct {
		Name string `json:"name"`
	}

	var m mod
	err := json.NewDecoder(req.Body).Decode(&m)
	if err != nil {
		log.Print("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	err = os.Remove(BasePath + "/mods/" + m.Name)
	if err != nil {
		log.Print("Error deleting mod:", err)
		http.Error(w, "Error deleting mod", http.StatusInternalServerError)
		return

	}
	modlist, err := getModFilenames()
	if err != nil {
		log.Print("Error getting list of mods:", err)
		http.Error(w, "Error getting list of mods", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(modlist)
	if err != nil {
		log.Print("Error encoding list of mods to JSON:", err)
		http.Error(w, "Error encoding list of mods to JSON", http.StatusInternalServerError)
		return
	}
}

func wrapRconResponse(a string) map[string]string {
	var wa = make(map[string]string)
	wa["answer"] = a
	return wa
}

func getDifficulty(w http.ResponseWriter, r *http.Request) {
	response, err := getDifficultyRCON()
	if err != nil {
		log.Print("Error getting difficulty:", err)
		http.Error(w, "Error getting difficulty", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(wrapRconResponse(response))
	if err != nil {
		log.Print("Error encoding to JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

func changeDifficulty(w http.ResponseWriter, r *http.Request) {
	type cmd struct {
		Text string `json:"text"`
	}
	var c cmd
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		log.Print("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	response, err := changeDifficultyRCON(c.Text)
	if err != nil {
		log.Print("Error changing difficulty:", err)
		http.Error(w, "Error changing difficulty", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(wrapRconResponse(response))
	if err != nil {
		log.Print("Error encoding to JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

func changeGamerule(w http.ResponseWriter, r *http.Request) {
	type gamerule struct {
		Rule  string
		Value string
	}
	var gr gamerule
	err := json.NewDecoder(r.Body).Decode(&gr)
	if err != nil {
		log.Print("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	response, err := changeGameruleRCON(gr.Rule, gr.Value)
	if err != nil {
		log.Print("Error changing gamerule:", err)
		http.Error(w, "Error changing gamerule", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(wrapRconResponse(response))
	if err != nil {
		log.Print("Error encoding to JSON:", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("Auth middleware logic")
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	fmt.Println("Running server with env:")

	fmt.Printf("RCON_HOST: %s\nRCON_PORT: %s\nRCON_PASS: %s\n", RCON_HOST, RCON_PORT, RCON_PASS)

	mux.Handle("GET /rcon/difficulty", authMiddleware(http.HandlerFunc(getDifficulty)))
	http.HandleFunc("POST /rcon/difficulty", changeDifficulty)
	http.HandleFunc("POST /rcon/gamerule", changeGamerule)
	http.HandleFunc("/properties", handleProperties)
	http.HandleFunc("/mods", handleMods)

	err := http.ListenAndServe(":8090", mux)
	if errors.Is(err, http.ErrServerClosed) {
		log.Print("server closed\n")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func errorChecker(e error) {
	if e != nil {
		fmt.Println("An error occured: " + e.Error())
		return
	}
}
