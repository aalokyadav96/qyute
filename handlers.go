package main

import (
    "fmt"
    "net/http"
    "log"
	"io"
	"os"
	"time"
	"os/exec"
	"crypto/rand"
	rndm "math/rand"
	"crypto/md5"
	"path/filepath"

    "github.com/julienschmidt/httprouter"
)

// Gif - We will be using this Gif type to perform crud operations
type GIF struct {
	Title  string
	Author string
	Tags   []string
	Date   string
	URL    string
	Views  int
	Likes  int
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}


func UploadVideoFileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	enableCors(&w)
	if r.Method == "POST" {           
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			renderError(w, "CANT_PARSE_FORM", http.StatusInternalServerError)
		}
		fmt.Println(r.FormValue("csrftoken"))
		var count int
		var fileEndings string
		var folderpath string
		var fileName string
		var postType string
		var postName string = GenerateName(12)

		files := r.MultipartForm.File["file"]
		for _, fileHeader := range files {
			log.Println("hao")
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer file.Close()
			log.Println("file OK")
			count++
			fileSize := fileHeader.Size
			fmt.Printf("File size (bytes): %v\n", fileSize)
			// validate file size
			if fileSize > maxUploadSize {
				renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			}
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				renderError(w, "INVALID_FILE"+http.DetectContentType(fileBytes), http.StatusBadRequest)
			}
			//~ // check file type, detectcontenttype only needs the first 512 bytes
			detectedFileType := http.DetectContentType(fileBytes)
			switch detectedFileType {
			case "video/mp4":
				fileEndings = ".mp4"
				folderpath = "./videos"
				postType = "video"
				break
			case "video/webm":
				fileEndings = ".webm"
				folderpath = "./videos"
				postType = "video"
				break
			case "image/jpg", "image/jpeg", "image/png", "image/webp":
				fileEndings = ".png"
				folderpath = "./images"
				postType = "image"
				break
			default:
				renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			}
			// if fileName exists in Redis, again GenerateName(rndmToken(12))
			//		fileEndings, err := mime.ExtensionsByType(detectedFileType)
			//~ fileName = GenerateName(16)
			fileName = fmt.Sprintf("%v_%d",postName,count)
			if err != nil {
				renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			}
			newFileName := fileName + fileEndings

			newPath := filepath.Join(folderpath, newFileName)
			fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

			// write file
			newFile, err := os.Create(newPath)
			if err != nil {
				renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			}
			defer newFile.Close() // idempotent, okay to call twice
			if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
				renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			}
		}
			FFConvert(fileName , fileEndings )
			
			//~ fmt.Fprintf(w, "{\"postname\": \""+postName+"\", \"postcount\": "+fmt.Sprintf("%d",count)+"}")
			fmt.Fprintf(w, "{\"postname\":\""+postName+"\", \"postcount\":"+fmt.Sprintf("%d",count)+", \"posttype\":\""+ fmt.Sprintf("%s",postType) +"\"}")
	}
}

func FFConvert(fileName string, fileEndings string) {
	getFrom := "./videos" + "/" + fileName + fileEndings
	saveAs := "./streams" + "/" + fileName + ".mp4"
	cmd := exec.Command("ffmpeg", "-i", getFrom, "-filter:v", "scale=-2:640:flags=lanczos", "-c:a", "copy", "-pix_fmt", "yuv420p", saveAs)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
}

func sendImageAsBytes(w http.ResponseWriter, r *http.Request, a httprouter.Params) {
	buf, err := os.ReadFile("./images/"+a.ByName("imageName"))
	if err != nil {
		log.Print(err)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
}


func CSRF(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	enableCors(&w)
	fmt.Fprintf(w,GenerateName(8))
}


func Res(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	enableCors(&w)
	var translated = "trns"
	var lang = "EN"
	fmt.Fprintf(w,"{\"translated\":\""+translated+"\", \"lang\":\""+lang+"\"}")
}

func Translate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	enableCors(&w)
	var translated = r.FormValue("trns")
	var lang = "EN"
	fmt.Fprintf(w, "{\"translated\":\""+translated+"\", \"lang\":\""+lang+"\"}")
}

/*
func UploadVideoFileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	enableCors(&w)
	if r.Method == "POST" {           
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			renderError(w, "CANT_PARSE_FORM", http.StatusInternalServerError)
		}
		var postname string = GenerateName(12)
		var fileEndings string
		var fileName string
		var folderpath string

		files := r.MultipartForm.File["file"]
		for filecount, fileHeader := range files {
			fmt.Println(filecount)
			log.Println("hao")
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer file.Close()
			log.Println("file OK")
			title := r.FormValue("title")
			tags := strings.ToLower(r.FormValue("tags"))
			fmt.Println(tags)
			// Get and print outfile size
			
			f := func(c rune) bool {
			return !unicode.IsLetter(c) && !unicode.IsNumber(c)
			}
			titleArr := strings.FieldsFunc(strings.ToLower(title), f)
			fmt.Printf("Fields are: %q", titleArr)
			fileSize := fileHeader.Size
			fmt.Printf("File size (bytes): %v\n", fileSize)
			// validate file size
			if fileSize > maxUploadSize {
				renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			}
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				renderError(w, "INVALID_FILE"+http.DetectContentType(fileBytes), http.StatusBadRequest)
			}
			
			//~ // check file type, detectcontenttype only needs the first 512 bytes
			detectedFileType := http.DetectContentType(fileBytes)
			switch detectedFileType {
			case "video/mp4":
				fileEndings = ".mp4"
				folderpath = "./videos"
				break
			case "video/webm":
				fileEndings = ".webm"
				folderpath = "./videos"
				break
			case "image/gif":
				fileEndings = ".gif"
				folderpath = "./images"
				break
			case "image/png":
				fileEndings = ".png"
				folderpath = "./images"
				break
			case "image/webp":
				fileEndings = ".webp"
				folderpath = "./images"
				break
			case "image/jpg":
				fileEndings = ".jpg"
				folderpath = "./images"
				break
			case "image/jpeg":
				fileEndings = ".jpeg"
				folderpath = "./images"
				break
			default:
				renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			}
			fileName = r.FormValue("key")
			fmt.Println("fileName : ", fileName)
			if err != nil {
				renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			}
			//~ newFileName := fileName + fileEndings
			//~ newPath := filepath.Join(uploadPath, newFileName)
			//~ newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileEndings)
			//~ newFileName := GenerateName(16)+fmt.Sprintf("%d",filecount)+fileEndings
			newFileName := postname+fmt.Sprintf("%d",filecount)+fileEndings
			fmt.Println(newFileName)
			newPath := filepath.Join(folderpath, newFileName)
			//~ newPath := fmt.Sprintf("./images/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
			//~ fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

			fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
			fmt.Printf("File Size: %+v\n", fileHeader.Size)
			fmt.Printf("MIME Header: %+v\n", fileHeader.Header)
			// write file
			newFile, err := os.Create(newPath)
			if err != nil {
				renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			}
			defer newFile.Close() // idempotent, okay to call twice
			if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
				renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			}
		}
			fmt.Fprintf(w, postname)
//				tmpl.ExecuteTemplate(w, "show.html", fileName)
	}
}*/

func GenerateName(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rndm.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	rndm.Seed(time.Now().UnixNano())
}


func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func rndmToken(len int) int64 {
	b := make([]byte, len)
	n, _ := rand.Read(b)
	return int64(n)
}

func EncrypIt(strToHash string) string {
	data := []byte(strToHash)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func SessionVerify(sessionKey string) string {
	return fmt.Sprintf(sessionKey)
}

func Ignore(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "favicon.png")
}

