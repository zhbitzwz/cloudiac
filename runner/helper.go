package runner

import (
	"bytes"
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type IaCTemplate struct {
	TemplateUUID string
	TaskId       string
}

type StateStore struct {
	SaveState           bool   `json:"save_state"`
	Backend             string `json:"backend" default:"consul"`
	Scheme              string `json:"scheme" default:"http"`
	StateKey            string `json:"state_key"`
	StateBackendAddress string `json:"state_backend_address"`
	Lock                bool   `json:"lock" defalt:"true"`
}

// ReqBody from reqeust
type ReqBody struct {
	Repo         string     `json:"repo"`
	RepoCommit   string     `json:"repo_commit"`
	RepoBranch   string     `json:"repo_branch"`
	TemplateUUID string     `json:"template_uuid"`
	TaskID       string     `json:"task_id"`
	DockerImage  string     `json:"docker_image" defalut:"mt5225/tf-ansible:v0.0.1"`
	StateStore   StateStore `json:"state_store"`
	Env          map[string]string
	Timeout      int    `json:"timeout" default:"600"`
	Mode         string `json:"mode" default:"plan"`
	Varfile      string `json:"varfile"`
	Extra        string `json:"extra"`
	Playbook     string `json:"playbook" form:"playbook" `
}

type CommitedTask struct {
	TemplateId       string `json:"templateId"`
	TaskId           string `json:"taskId"`
	ContainerId      string `json:"containerId"`
	LogContentOffset int    `json:"offset"`
}

// 判断目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 创建目录
func PathCreate(path string) error {
	pathExists, err := PathExists(path)
	if err != nil {
		return err
	}
	if pathExists == true {
		return nil
	} else {
		err := os.MkdirAll(path, os.ModePerm)
		return err
	}
}

// 从指定位置读取日志文件
func ReadLogFile(filepath string, offset int, maxLines int) ([]string, error) {
	var lines []string
	// TODO(ZhengYue): 优化文件读取，考虑使用seek跳过偏移行数
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return lines, err
	}
	buf := bytes.NewBuffer(file)
	lineCount := 0
	for {
		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err != nil {
				if err == io.EOF {
					break
				}
				return lines, err
			}
		}
		lineCount += 1
		if lineCount > offset {
			// 未达到偏移位置，继续读取
			lines = append(lines, line)
		}
		if len(lines) == maxLines {
			// 达到最大行数，立即返回
			return lines, err
		}
		if err != nil && err != io.EOF {
			return lines, err
		}
	}
	return lines, nil
}

func GetTaskWorkDir(templateUUID string, taskId string) string {
	conf := configs.Get()
	return filepath.Join(conf.Runner.StoragePath, templateUUID, taskId)
}

func FetchTaskLog(templateUUID string, taskId string) ([]byte, error) {
	logFile := filepath.Join(GetTaskWorkDir(templateUUID, taskId), TaskLogName)
	return ioutil.ReadFile(logFile)
}

func MakeTaskWorkDir(tplId string, taskId string) (string, error) {
	workDir := GetTaskWorkDir(tplId, taskId)
	err := PathCreate(workDir)
	return workDir, err
}

func ReqToTask(req *http.Request) (*CommitedTask, error) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	logger := logs.Get()
	logger.Debugf("task status request: %s", bodyData)

	var d CommitedTask
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ReqToCommand create command structure to run container
// from POST request
func ReqToCommand(req *http.Request) (*Command, *StateStore, error) {
	bodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}

	logger := logs.Get()
	logger.Tracef("new task request: %s", bodyData)
	var d ReqBody
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, nil, err
	}

	c := new(Command)

	state := d.StateStore

	if d.DockerImage == "" {
		conf := configs.Get()
		c.Image = conf.Runner.DefaultImage
	} else {
		c.Image = d.DockerImage
	}

	for k, v := range d.Env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range AnsibleEnv {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// TODO(ZhengYue): 优化命令组装方式
//	var cmdList []string
//	logCmd := fmt.Sprintf(">> %s%s 2>&1 ", ContainerLogFilePath, ContainerLogFileName)
//	//ansibleCmd := fmt.Sprint(" if [ -e run.sh ];then chmod +x run.sh && ./run.sh;fi")
//	//ansibleCmd := fmt.Sprint("cd ansible")
//	//ansiblePlaybook:=fmt.Sprintf("ansible-playbook -i ./terraform.py playbook.yml")
//	ansiblePlaybook := fmt.Sprintf("ansible-playbook -i %s  %s", "", d.Playbook)
//	cmdList = append(cmdList, fmt.Sprintf("git clone %s %s &&", d.Repo, logCmd))
//	// get folder name
//	s := strings.Split(d.Repo, "/")
//	f := s[len(s)-1]
//	f = f[:len(f)-4]
//
//	cmdList = append(cmdList, fmt.Sprintf("cd %s %s &&", f, logCmd))
//	cmdList = append(cmdList, fmt.Sprintf("git checkout  %s %s &&", d.RepoBranch, logCmd))
//	cmdList = append(cmdList, fmt.Sprintf("git checkout -b run_branch %s %s &&", d.RepoCommit, logCmd))
//	cmdList = append(cmdList, fmt.Sprintf("cp %sstate.tf . &&", ContainerLogFilePath))
//	cmdList = append(cmdList, fmt.Sprintf("terraform init  -plugin-dir %s %s &&", ContainerProviderPath, logCmd))
//	if d.Mode == "apply" {
//		log.Println("entering apply mode ...")
//		if d.Varfile != "" {
//			cmdList = append(cmdList, fmt.Sprintf("%s %s %s && %s %s &&%s %s",
//				"terraform apply -auto-approve -var-file ", d.Varfile, logCmd, ansiblePlaybook ,logCmd, d.Extra, logCmd))
//		} else {
//			cmdList = append(cmdList, fmt.Sprintf("%s %s &&%s %s &&%s %s", "terraform apply -auto-approve ", logCmd, ansiblePlaybook, logCmd, d.Extra, logCmd))
//		}
//
//	} else if d.Mode == "destroy" {
//		log.Println("entering destroy mode ...")
//		cmdList = append(cmdList, fmt.Sprintf("%s %s&&%s", "terraform destroy -auto-approve -var-file", d.Varfile, d.Extra))
//	} else if d.Mode == "pull" {
//		log.Println("show state info ...")
//		cmdList = append(cmdList, fmt.Sprintf("%s&&%s", "terraform state pull", d.Extra))
//	} else {
//		if d.Varfile != "" {
//			cmdList = append(cmdList, fmt.Sprintf("%s %s %s", "terraform plan -var-file ", d.Varfile, logCmd))
//		} else {
//			cmdList = append(cmdList, fmt.Sprintf("%s %s", "terraform plan  ", logCmd))
//		}
//
//		//cmdList = append(cmdList, fmt.Sprintf("%s %s&&%s", "terraform plan -var-file", d.Varfile, d.Extra))
//=======
	workingDir, err := MakeTaskWorkDir(d.TemplateUUID, d.TaskID)
	if err != nil {
		return nil, nil, err
	}

	c.TaskWorkdir = workingDir
	scriptPath := filepath.Join(c.TaskWorkdir, TaskScriptName)
	if err := GenScriptContent(&d, scriptPath); err != nil {
		return nil, nil, err
	}

	containerScriptPath := filepath.Join(ContainerIaCDir, TaskScriptName)
	containerLogPath := filepath.Join(ContainerIaCDir, TaskLogName)
	c.Commands = []string{"sh", "-c", fmt.Sprintf("%s >>%s 2>&1", containerScriptPath, containerLogPath)}

	// set timeout
	c.Timeout = d.Timeout
	c.ContainerInstance = new(Container)
	c.ContainerInstance.Context = context.Background()
	log.Printf("new task command: %#v", c)
	return c, &state, nil
}

func LineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
