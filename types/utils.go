package types

import (
	"encoding/json"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/runtime/protoiface"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var issueRandStr map[string]bool

func init() {
	issueRandStr = make(map[string]bool)
}

func GetRandString(size ...uint) string {
	var str string
	for {
		str = uuid.New().String()
		if len(size) != 0 {
			str = str[:size[0]]
		}
		if ok := issueRandStr[str]; !ok {
			break
		}
	}

	return str
}

func UnmarshalProtoMessage(data string, obj protoiface.MessageV1) {
	if err := jsonpb.UnmarshalString(data, obj); err != nil {
		log.WithFields(log.Fields{"error": err, "data": data}).Fatal("error unmarshal message")
	}
}

func MarshalProtoMessage(obj proto.Message) (string, error) {
	m := jsonpb.Marshaler{}
	d, err := m.MarshalToString(obj)
	if err != nil {
		return "", err
	}

	return d, nil
}

func CopyProtoMessage(src protoiface.MessageV1, dst protoiface.MessageV1) error {
	d, err := MarshalProtoMessage(src)
	if err != nil {
		return err
	}

	UnmarshalProtoMessage(d, dst)
	return nil
}

func FileExists(filePath string) bool {
	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	if stat.IsDir() {
		return false
	}

	return true
}

func MkDir(dir string) error {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.Mkdir(dir, os.ModePerm)
	}
	if stat.IsDir() {
		return nil
	}

	return fmt.Errorf("plugin directory can not be avaiable")
}

func DownloadFile(url, filePath string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	openFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer openFile.Close()

	if _, err := io.Copy(openFile, res.Body); err != nil {
		return err
	}

	return nil
}

func ReadPlugin(filePath string) ([]byte, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if string(fileContent[:len(KplayerPluginSignHeader)]) == KplayerPluginSignHeader {
		encryptData := fileContent[len(KplayerPluginSignHeader):]

		// aes decrypt
		decryptData, err := openssl.AesCBCDecrypt(encryptData, []byte(CipherKey), []byte(CipherIV), openssl.PKCS5_PADDING)
		if err != nil {
			log.Fatal(err)
		}
		return decryptData, nil
	}

	return fileContent, nil
}

func FormatYamlProtoMessage(msg proto.Message) (string, error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func GetClientContextFromCommand(cmd *cobra.Command) *ClientContext {
	var clientCtx *ClientContext
	if ptr, err := GetCommandContext(cmd, ClientContextKey); err != nil {
		log.Fatalf("get client context failed. error: %s", err)
	} else {
		clientCtx = ptr.(*ClientContext)
	}

	return clientCtx
}