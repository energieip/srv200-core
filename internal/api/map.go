package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/srv200-coreservice-go/internal/core"
	"github.com/romana/rlog"
)

func (api *API) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if api.hasAccessMode(w, r, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	var p *multipart.Part
	var err error

	mr, err := r.MultipartReader()
	if err != nil {
		rlog.Error("Hit error while opening multipart reader: ", err.Error())
		api.sendError(w, APIErrorBodyParsing, "Error while fetching file", http.StatusInternalServerError)
		return
	}

	chunk := make([]byte, 4096) // 4k size byte slice
	tempFile, err := ioutil.TempFile(api.dataPath, "temp-file")
	if err != nil {
		rlog.Error("Hit error while creating temp file: ", err.Error())
		api.sendError(w, APIErrorBodyParsing, "Error while fetching file", http.StatusInternalServerError)
		return
	}

	err = os.Chmod(tempFile.Name(), 0644)
	if err != nil {
		rlog.Error("Hit error while creating temp file: ", err.Error())
		api.sendError(w, APIErrorBodyParsing, "Error while fetching file", http.StatusInternalServerError)
		return
	}
	newFilename := ""

	// continue looping through all parts, *multipart.Reader.NextPart() will
	// return an End of File when all parts have been read.
	for {
		p, err = mr.NextPart()
		if err == io.EOF {
			// err is io.EOF, files upload completes.
			tempFile.Close()
			rlog.Info("Hit last part of multipart upload / do post treatment")
			go func(filename string) {
				cmd := exec.Command("ifcparser.py", "-i", filename)
				out, err := cmd.CombinedOutput()
				if err != nil {
					rlog.Error("cmd.Run() failed with status " + err.Error() + " : " + string(out))
					os.Remove(tempFile.Name())
					return
				}
				mapInfo := core.MapInfo{}
				err = json.Unmarshal(out, &mapInfo)
				if err != nil {
					rlog.Error("Cannot parse command ", err.Error())
					os.Remove(tempFile.Name())
					return
				}
				rlog.Info("Rename " + tempFile.Name() + " into " + newFilename)
				err = os.Rename(tempFile.Name(), newFilename)
				if err != nil {
					rlog.Error("Cannot parse command ", err.Error())
					os.Remove(tempFile.Name())
					return
				}
				cmd = exec.Command("ifc2gltf.py", "-i", newFilename)
				out, err = cmd.CombinedOutput()
				if err != nil {
					rlog.Error("ifc2gltf.py failed with status " + err.Error() + " : " + string(out))
					return
				}

				event := make(map[string]interface{})
				event["map"] = mapInfo
				api.EventsToBackend <- event
			}(tempFile.Name())
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			break
		}
		if err != nil {
			// A normal error occurred
			tempFile.Close()
			os.Remove(tempFile.Name())
			rlog.Error("Hit error while fetching next part: ", err.Error())
			api.sendError(w, APIErrorBodyParsing, "Error while fetching file", http.StatusInternalServerError)
			return
		}

		newFilename = api.dataPath + "/" + p.FileName()
		rlog.Info("Uploaded filename: " + newFilename)
		uploaded := false

		// continue reading the part stream of this loop until either done or err.
		for !uploaded {
			n, err := p.Read(chunk)
			if err != nil {
				if err != io.EOF {
					rlog.Error("Hit error while writing chunk: ", err.Error())
					api.sendError(w, APIErrorBodyParsing, "Error while fetching file", http.StatusInternalServerError)
					return
				}
				uploaded = true
			}
			if _, err = tempFile.Write(chunk[:n]); err != nil {
				rlog.Error("Hit error while writing chunk: ", err.Error())
				api.sendError(w, APIErrorBodyParsing, "Error while fetching file", http.StatusInternalServerError)
				return
			}
		}
	}
}
