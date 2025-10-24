package filemanager

type FileListInfo struct {
	Path    string `json:"path"`
	Newname string `json:"newname"`
	Dest    string `json:"dest"`
	Ondup   string `json:"ondup"`
}

type FileManagerArg struct {
	Opera    string         `json:"opera"`
	Async    string         `json:"async"`
	FileList []FileListInfo `json:"filelist"`
}

func NewFileManagerArg(opera string, async string, fileList []FileListInfo) *FileManagerArg {
	s := new(FileManagerArg)
	s.Opera = opera
	s.Async = async
	s.FileList = fileList
	return s
}

type FileManagerReturn struct {
	Errno     int                      `json:"errno"`
	Info      []map[string]interface{} `json:"info"`
	Taskid    int                      `json:"taskid"` // 异步下返回此参数
	RequestId int                      `json:"request_id"`
}
