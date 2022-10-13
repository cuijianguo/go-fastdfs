package server

import (
	"fmt"
	ssh "github.com/sjqzhang/gssh"
	log "github.com/sjqzhang/seelog"
	"io/ioutil"
	"net/http"
)

func (c *Server) SSHUpload(w http.ResponseWriter, r *http.Request) {
	var (
		address  string
		port     string
		account  string
		pwd      string
		md5sum   string
		remote   string
		result   JsonResult
		fileInfo *FileInfo
		client   *ssh.Client
		err      error
		data     []byte
		//ok      bool
	)
	result.Status = "fail"

	// redirect to upload
	if r.RequestURI == "/" || r.RequestURI == "" ||
		r.RequestURI == "/"+Config().Group ||
		r.RequestURI == "/"+Config().Group+"/" {
		c.Index(w, r)
		return
	}

	if Config().EnableCrossOrigin {
		c.CrossOrigin(w, r)
	}

	body, _ := ioutil.ReadAll(r.Body)

	var bodyStruct = struct {
		Address string `json:"address"`
		Port    string `json:"port"`
		Account string `json:"account"`
		Pwd     string `json:"pwd"`
		Md5     string `json:"md5"`
		Remote  string `json:"remote"`
	}{}
	_ = json.Unmarshal(body, &bodyStruct)

	address = bodyStruct.Address
	port = bodyStruct.Port
	account = bodyStruct.Account
	pwd = bodyStruct.Pwd
	md5sum = bodyStruct.Md5
	remote = bodyStruct.Remote

	client, err = ssh.NewClient(address, port, account, pwd)
	if err != nil {
		result.Message = err.Error()
		log.Error(err.Error())
		w.Write([]byte(c.util.JsonEncodePretty(result)))
		return
	}
	defer client.Close()

	fileInfo, err = c.GetFileInfoFromLevelDB(md5sum)
	if err != nil {
		log.Error(err)
		result.Message = err.Error()
		w.Write([]byte(c.util.JsonEncodePretty(result)))
		return
	}
	if fileInfo != nil {
		if fileInfo.OffSet != -1 {
			if data, err = json.Marshal(fileInfo); err != nil {
				log.Error(err)
			}
			w.Write(data)
			return
		}
		fpath := DOCKER_DIR + fileInfo.Path + "/" + fileInfo.Name
		if fileInfo.ReName != "" {
			fpath = DOCKER_DIR + fileInfo.Path + "/" + fileInfo.ReName
		}
		if !c.util.IsExist(fpath) {
			result.Message = "文件未找到"
			w.Write([]byte(c.util.JsonEncodePretty(result)))
			return
		}

		err = client.Upload(fpath, remote)
		if err != nil {
			result.Message = err.Error()
			w.Write([]byte(c.util.JsonEncodePretty(result)))
			return
		}

		result.Status = "ok"
		result.Message = "ssh上传成功"
		w.Write([]byte(c.util.JsonEncodePretty(result)))
		fmt.Println(11123, fileInfo, c.util.IsExist(fpath), fpath)
		return
	}

	result.Message = "文件未找到"
	w.Write([]byte(c.util.JsonEncodePretty(result)))
	return
}
