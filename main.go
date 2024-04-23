package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

var BasePath = "data"

func hello(w http.ResponseWriter, req *http.Request) {
	_, err := fmt.Fprintf(w, "hi!")
	if err != nil {
		return
	}
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
	props, err := ReadProperties(BasePath + "/server.properties")
	if err != nil {
		fmt.Println("Error reading properties file:", err)
		return
	}
	allProps := make(map[string]string)
	for key, value := range props {
		allProps[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(allProps)
	if err != nil {
		fmt.Println("Error encoding properties to JSON:", err)
		http.Error(w, "Error encoding properties to JSON", http.StatusInternalServerError)
		return
	}

}

// TODO: Return modified properties as JSON
func modifyProps(w http.ResponseWriter, req *http.Request) {
	type prop struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	var p prop
	err := json.NewDecoder(req.Body).Decode(&p)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	props, err := ReadProperties(BasePath + "/server.properties")
	if err != nil {
		fmt.Println("Error reading properties file:", err)
		http.Error(w, "Error reading properties file", http.StatusInternalServerError)
		return
	}

	props[p.Key] = p.Value
	err = OverwriteProperties(BasePath+"/server.properties", props)
	if err != nil {
		fmt.Println("Error writing properties file:", err)
		http.Error(w, "Error writing properties file", http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprint(w, "success")
	if err != nil {
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
		fmt.Println("Error getting list of mods:", err)
		http.Error(w, "Error getting list of mods", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(mods)
	if err != nil {
		fmt.Println("Error encoding list of mods to JSON:", err)
		http.Error(w, "Error encoding list of mods to JSON", http.StatusInternalServerError)
		return
	}
}

func addMod(w http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(100 << 20)
	if err != nil {
		fmt.Println("Error parsing file:", err)

		http.Error(w, "Error parsing file", http.StatusBadRequest)
		return
	}

	file, handler, err := req.FormFile("file")
	if err != nil {
		fmt.Println("Error parsing file:", err)
		http.Error(w, "Error parsing file", http.StatusBadRequest)
		return

	}
	defer file.Close()

	dst, err := os.Create(BasePath + "/mods/" + handler.Filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		fmt.Println("Error saving file:", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	modlist, err := getModFilenames()
	if err != nil {
		fmt.Println("Error getting list of mods:", err)
		http.Error(w, "Error getting list of mods", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(modlist)
	if err != nil {
		fmt.Println("Error encoding list of mods to JSON:", err)
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
		fmt.Println("Error decoding JSON:", err)
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	err = os.Remove(BasePath + "/mods/" + m.Name)
	if err != nil {
		fmt.Println("Error deleting mod:", err)
		http.Error(w, "Error deleting mod", http.StatusInternalServerError)
		return

	}
	modlist, err := getModFilenames()
	if err != nil {
		fmt.Println("Error getting list of mods:", err)
		http.Error(w, "Error getting list of mods", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(modlist)
	if err != nil {
		fmt.Println("Error encoding list of mods to JSON:", err)
		http.Error(w, "Error encoding list of mods to JSON", http.StatusInternalServerError)
		return
	}
}
func main() {
	fmt.Println("test")

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/properties", handleProperties)
	http.HandleFunc("/mods", handleMods)

	err := http.ListenAndServe(":8090", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func errorChecker(e error) {
	if e != nil {
		fmt.Println("An error occured: " + e.Error())
		return
	}
}
