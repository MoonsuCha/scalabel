package main

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"html/template"
	"reflect"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"log"
)

// A collection of Items to be split into Tasks. Represents one unified type
//   of annotation task with a consistent ItemType, LabelType, Category list,
//   and Attribute list. Tasks in this Project are of uniform size.
type Project struct {
	Items    []Item         `json:"items" yaml"items"`
	VendorId int            `json:"vendorId" yaml:"vendorId"`
	Options  ProjectOptions `json:"options" yaml:"options"`
}

func (project *Project) GetPath() string {
	dir := path.Join(
		env.DataDir,
		project.Options.Name,
	)
	os.MkdirAll(dir, 0777)
	return path.Join(dir, "project.json")
}

func (project *Project) Save() {
	path := project.GetPath()
	json, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		Error.Println(err)
	}
	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		Error.Println(err)
	} else {
		Info.Println("Saving project file of", project.Options.Name)
	}
}

// Info about a Project shared by Project and Task.
type ProjectOptions struct {
	Name                 string        `json:"name" yaml:"name"`
	ItemType             string        `json:"itemType" yaml:"itemType"`
	LabelType            string        `json:"labelType" yaml:"labelType"`
	TaskSize             int           `json:"taskSize" yaml:"taskSize"`
	HandlerUrl           string        `json:"handlerUrl" yaml:"handlerUrl"`
	PageTitle            string        `json:"pageTitle" yaml:"pageTitle"`
	Categories           []Category    `json:"categories" yaml:"categories"`
	NumLeafCategories    int           `json:"numLeafCategories" yaml:"numLeafCategories"`
	Attributes           []Attribute   `json:"attributes" yaml:"attributes"`
	VideoMetaData        VideoMetaData `json:"metadata" yaml:"metadata"`
}

// A workably-sized collection of Items belonging to a Project.
type Task struct {
	ProjectOptions ProjectOptions `json:"projectOptions" yaml:"projectOptions"`
	Index          int            `json:"index" yaml:"index"`
	Items          []Item         `json:"items" yaml:"items"`
}

func (task *Task) GetPath() string {
	dir := path.Join(
		env.DataDir,
		task.ProjectOptions.Name,
		"tasks",
	)
	os.MkdirAll(dir, 0777)
	return path.Join(dir, strconv.Itoa(task.Index)+".json")
}

func (task *Task) Save() {
	path := task.GetPath()
	Info.Println(path)
	json, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		Error.Println(err)
	}
	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		Error.Println(err)
	} else {
		Info.Println("Saving task file of", task.ProjectOptions.Name, task.Index)
	}
}

// The actual assignment of a task to a worker. Contains the worker's progress.
type Assignment struct {
	Task            Task                   `json:"task" yaml:"task"`
	WorkerId        string                 `json:"workerId" yaml:"workerId"`
	Labels          []Label                `json:"labels" yaml:"labels"`
	Tracks          []Label                `json:"tracks" yaml:"tracks"`
	Events          []Event                `json:"events" yaml:"events"`
	StartTime       int64                  `json:"startTime" yaml:"startTime"`
	SubmitTime      int64                  `json:"submitTime" yaml:"submitTime"`
	NumLabeledItems int                    `json:"numLabeledItems" yaml:"numLabeledItems"`
	UserAgent       string                 `json:"userAgent" yaml:"userAgent"`
	IpInfo          map[string]interface{} `json:"ipInfo" yaml:"ipInfo"`
}

func (assignment *Assignment) GetAssignmentPath() string {
	dir := path.Join(
		env.DataDir,
		assignment.Task.ProjectOptions.Name,
		"assignments",
		strconv.Itoa(assignment.Task.Index),
	)
	os.MkdirAll(dir, 0777)
	return path.Join(dir, assignment.WorkerId+".json")
}

func (assignment *Assignment) GetSubmissionPath() string {
	dir := path.Join(
		env.DataDir,
		assignment.Task.ProjectOptions.Name,
		"submissions",
		strconv.Itoa(assignment.Task.Index),
		assignment.WorkerId,
	)
	os.MkdirAll(dir, 0777)
	return path.Join(dir, strconv.FormatInt(assignment.SubmitTime, 10)+".json")
}

func (assignment *Assignment) Initialize() {
	path := assignment.GetAssignmentPath()
	assignment.Serialize(path)
}

func (assignment *Assignment) Save() {
	path := assignment.GetSubmissionPath()
	assignment.Serialize(path)
}

func (assignment *Assignment) Serialize(path string) {
	json, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		Error.Println(err)
	}
	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		Error.Println(err)
	} else {
		Info.Println("Saving assignment file of",
			assignment.Task.ProjectOptions.Name, assignment.Task.Index,
			assignment.WorkerId, "to", path)
	}
}

// An item is something to be annotated e.g. Image, PointCloud
type Item struct {
	Url         string  `json:"url" yaml:"url"`
	Index       int     `json:"index" yaml:"index"`
	LabelIds    []int   `json:"labelIds" yaml:"labelIds"`
	GroundTruth []Label `json:"groundTruth" yaml:"groundTruth"`
}

type Vertex struct {
    Id          int                     `json:"id" yaml:"id"`
    X           float32                 `json:"x" yaml:"x"`
    Y           float32                 `json:"y" yaml:"y"`
    Type        string                  `json:"type" yaml:"type"`
}

type Edge struct {
    Id              int                 `json:"id" yaml:"id"`
    Src             int                 `json:"src" yaml:"src"`
    Dest            int                 `json:"dest" yaml:"dest"`
    Type            string              `json:"type" yaml:"type"`
    ControlPoints   []Vertex            `json:"control_points" yaml:"control_points"`
}

type Polyline struct {
    Id              int                 `json:"id" yaml:"id"`
    Vertices        []Vertex            `json:"vertices" yaml:"vertices"`
    Edges           []Edge              `json:"edges" yaml:"edges"`
}

type PolylineDownloadFormat struct {
    Vertices        [][]float32         `json:"vertices" yaml:"vertices"`
    Types           string              `json:"types" yaml:"types"`
}

type Box2d struct {
    X           float32                 `json:"x" yaml:"x"`
    Y           float32                 `json:"y" yaml:"y"`
    W           float32                 `json:"w" yaml:"w"`
    H           float32                 `json:"h" yaml:"h"`
}

type Box2dDownloadFormat struct {
    X1          float32                 `json:"x1" yaml:"x1"`
    X2          float32                 `json:"x2" yaml:"x2"`
    Y1          float32                 `json:"y1" yaml:"y1"`
    Y2          float32                 `json:"y2" yaml:"y2"`
}

type Poly2d struct {
    Closed      bool                     `json:"closed" yaml:"closed"`
    Polys       []Polyline               `json:"polys" yaml:"polys"`
}

type Poly2dDownloadFormat struct {
    Closed      bool                     `json:"closed" yaml:"closed"`
    Polys       []PolylineDownloadFormat `json:"polys" yaml:"polys"`
}

// An annotation for an item, needs to include all possible annotation types
type Label struct {
	Id           int                    `json:"id" yaml:"id"`
	CategoryPath string                 `json:"categoryPath" yaml:"categoryPath"`
	ParentId     int                    `json:"parentId" yaml:"parentId"`
	ChildrenIds  []int                  `json:"childrenIds" yaml:"childrenIds"`
	Attributes   map[string]interface{} `json:"attributes" yaml:"attributes"`
	Box2d        Box2d                  `json:"box2d" yaml:"box2d"`
	Poly2d       Poly2d                 `json:"poly2d" yaml:"poly2d"`
	Keyframe     bool                   `json:"keyframe" yaml:"keyframe"`
}

// A class value for a label.
type Category struct {
	Name          string     `json:"name" yaml:"name"`
	Subcategories []Category `json:"subcategories" yaml:"subcategories"`
}

// A configurable attribute describing a label
type Attribute struct {
	Name         string   `json:"name" yaml:"name"`
	ToolType     string   `json:"toolType" yaml:"toolType"`
	TagText      string   `json:"tagText" yaml:"tagText"`
	TagPrefix    string   `json:"tagPrefix" yaml:"tagPrefix"`
	TagSuffixes  []string `json:"tagSuffixes" yaml:"tagSuffixes"`
	Values       []string `json:"values" yaml:"values"`
	ButtonColors []string `json:"buttonColors" yaml:"buttonColors"`
}

// An event describing an annotator's interaction with the session
type Event struct {
	Timestamp  int64     `json:"timestamp" yaml:"timestamp"`
	Action     string    `json:"action" yaml:"action"`
	ItemIndex  int32     `json:"itemIndex" yaml:"itemIndex"`
	LabelIndex int32     `json:"labelIndex" yaml:"labelIndex"`
	Position   []float32 `json:"position" yaml:"position"`
}

// Contains all the info needed in the dashboards
type DashboardContents struct {
	Project Project `json:"project" yaml:"project"`
	Tasks   []Task  `json:"tasks" yaml:"tasks"`
}

// Download format specifications
type DownloadFormat struct {
	Name            string                     `json:"name" yaml:"name"`
	Attributes      []Attribute                `json:"attributes" yaml:"attributes"`
	Items           []ItemDownloadFormat       `json:"items" yaml:"items"`
}

type ItemDownloadFormat struct {
    Timestamp       int64                       `json:"timestamp" yaml:"timestamp"`
    Index           int                         `json:"index" yaml:"index"`
    Labels          []LabelDownloadFormat       `json:"labels" yaml:"labels"`
}

type LabelDownloadFormat struct {
    Id              int                         `json:"id" yaml:"id"`
    Category        string                      `json:"category" yaml:"category"`
    Attributes      map[string]interface{}      `json:"attributes" yaml:"attributes"`
    Box2d           Box2dDownloadFormat         `json:"box2d" yaml:"box2d"`
    Poly2d          Poly2dDownloadFormat        `json:"poly2d" yaml:"poly2d"`
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
    v := reflect.ValueOf(unk)
    v = reflect.Indirect(v)
    if !v.Type().ConvertibleTo(floatType) {
        return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
    }
    fv := v.Convert(floatType)
    return fv.Float(), nil
}

func parseBox2d(_box2d Box2d) (Box2dDownloadFormat) {
    box2d := Box2dDownloadFormat{}
    box2d.X1 = _box2d.X
    box2d.Y1 = _box2d.Y
    box2d.X2 = _box2d.X + _box2d.W
    box2d.Y2 = _box2d.Y + _box2d.H
    return box2d
}

func pushVertexToDownloadFormat(vertices [][]float32, vertex Vertex) ([][]float32) {
    xy := []float32{vertex.X, vertex.Y}
    vertices = append(vertices, xy)
    return vertices
}

func parsePoly2d(_poly2d Poly2d) (Poly2dDownloadFormat) {
    poly2d := Poly2dDownloadFormat{}
    for _, _poly := range _poly2d.Polys {
        poly := PolylineDownloadFormat{}
        for i, edge := range _poly.Edges {
            poly.Vertices = pushVertexToDownloadFormat(poly.Vertices, _poly.Vertices[i])
            poly.Types += "L"
            if (edge.Type == "bezier") {
                for _, c := range edge.ControlPoints {
                    poly.Vertices = pushVertexToDownloadFormat(poly.Vertices, c)
                    poly.Types += "C"
                }
            }
            if (!_poly2d.Closed && i == len(_poly.Edges)) {
                poly.Vertices = pushVertexToDownloadFormat(poly.Vertices, _poly.Vertices[i+1])
                poly.Types += "L"
            }
        }
        poly2d.Polys = append(poly2d.Polys, poly)
    }
    poly2d.Closed = _poly2d.Closed
    return poly2d
}

type TaskURL struct {
    URL             string                      `json:"url" yaml:"url"`
}

// unescaped marshal used to encode url string
func JSONMarshal(t interface{}) ([]byte, error) {
    buffer := &bytes.Buffer{}
    encoder := json.NewEncoder(buffer)
    encoder.SetEscapeHTML(false)
    err := encoder.Encode(t)
    return buffer.Bytes(), err
}

// Function type for handlers
type HandleFunc func(http.ResponseWriter, *http.Request)

// MakePathHandleFunc returns a function for handling static HTML
func MakePathHandleFunc(pagePath string) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HTML, err := ioutil.ReadFile(pagePath)
		if err != nil {
			Error.Println(err)
		}
		w.Write(HTML)
	}
}

func WrapHandler(handler http.Handler) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Info.Printf("%s is requesting %s", r.RemoteAddr, r.URL)
		handler.ServeHTTP(w, r)
	}
}

func WrapHandleFunc(fn HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Info.Printf("%s is requesting %s", r.RemoteAddr, r.URL)
		fn(w, r)
	}
}

func countCategories(categories []Category) int {
    count := 0
    for _, category := range categories {
        if len(category.Subcategories) > 0 {
            count += countCategories(category.Subcategories)
        } else {
            count += 1
        }
    }
    return count
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// use template to insert assignment links
	tmpl, err := template.ParseFiles(
		path.Join(env.DashboardPath()))
	if err != nil {
		Error.Println(err)
		http.NotFound(w, r)
		return
	}
	projectName := r.FormValue("project_name")
	dashboardContents := DashboardContents{
		Project: GetProject(projectName),
		Tasks:   GetTasksInProject(projectName),
	}
	Info.Println(dashboardContents.Tasks) // project is too verbose to log
	tmpl.Execute(w, dashboardContents)
}

func vendorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(env.VendorPath())
	if err != nil {
		Error.Println(err)
		http.NotFound(w, r)
		return
	}
	projectName := r.FormValue("project_name")
	dashboardContents := DashboardContents{
		Project: GetProject(projectName),
		Tasks:   GetTasksInProject(projectName),
	}
	Info.Println(dashboardContents.Tasks) // project is too verbose to log
	tmpl.Execute(w, dashboardContents)
}

// Handles the posting of new projects
func postProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	// make sure the project name in the form is new
	var projectName = CheckProjectName(r.FormValue("project_name"))
	if projectName == "" {
		w.Write([]byte("Project Name already exists."))
		return
	}
	// get item type from form
	itemType := r.FormValue("item_type")
	// get frame rate from form only if this is a video
	var videoMetaData VideoMetaData
	if itemType == "video" {
		videoMetaData.TBR = r.FormValue("frame_rate")
	}
	// get label type from form
	labelType := r.FormValue("label_type")
	// get page title from form
	pageTitle := r.FormValue("page_title")
	// parse the item list YML from form
	items := getItemsFromProjectForm(r)
	if itemType == "video" {
		videoMetaData.NumFrames = strconv.Itoa(len(items))
	}
	// parse the category list YML from form
	categories := getCategoriesFromProjectForm(r)
	numLeafCategories := countCategories(categories)
	// parse the attribute list YML from form
	attributes := getAttributesFromProjectForm(r)
	// get the task size from form
	taskSize, err := strconv.Atoi(r.FormValue("task_size"))
	if err != nil {
		Error.Println(err)
	}
	// get the vendor ID from form
	vendorId, err := strconv.Atoi(r.FormValue("vendor_id"))
	if err != nil {
		if (r.FormValue("vendor_id") == "") {
			vendorId = -1
		} else {
			Error.Println(err)
		}
	}

	// This prefix determines which handler will deal with labeling sessions
	//   for this project. Uniquely determined by item type and label type.
	handlerUrl := GetHandlerUrl(itemType, labelType)

	// initialize and save the project
	var projectOptions = ProjectOptions{
		Name:                projectName,
		ItemType:            itemType,
		LabelType:           labelType,
		TaskSize:            taskSize,
		HandlerUrl:          handlerUrl,
		PageTitle:           pageTitle,
		Categories:          categories,
		NumLeafCategories:   numLeafCategories,
		Attributes:          attributes,
		VideoMetaData:       videoMetaData,
	}
	var project = Project{
		Items:    items,
		VendorId: vendorId,
		Options:  projectOptions,
	}

	// Save project to project folder
	project.Save()

	// Initialize all the tasks
	CreateTasks(project)
}

func executeLabelingTemplate(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	// get task name from the URL
	projectName := r.URL.Query()["project_name"][0]
	taskIndex := r.URL.Query()["task_index"][0]
	var assignment Assignment
	if (!Exists(path.Join(env.DataDir, projectName, "assignments",
		taskIndex, DEFAULT_WORKER+".json"))) {
		// if assignment does not exist, create it
		assignment = CreateAssignment(projectName, taskIndex, DEFAULT_WORKER)
	} else {
		// otherwise, get that assignment
		assignment = GetAssignment(projectName, taskIndex, DEFAULT_WORKER)
	}
	tmpl.Execute(w, assignment)
}

// Handles the loading of an assignment given its project name, task index, and worker ID.
func postLoadAssignmentHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error.Println(err)
	}
	assignmentToLoad := Assignment{}
	err = json.Unmarshal(body, &assignmentToLoad)
	if err != nil {
		Error.Println(err)
	}
	projectName := assignmentToLoad.Task.ProjectOptions.Name
	taskIndex := strconv.Itoa(assignmentToLoad.Task.Index)
	var loadedAssignment Assignment
	if (!Exists(path.Join(env.DataDir, projectName, "assignments", taskIndex,
		DEFAULT_WORKER+".json"))) {
		// if assignment does not exist, create it
		// TODO: resolve tension between this function and executeLabelingTemplate()
		loadedAssignment = CreateAssignment(projectName, taskIndex,
			DEFAULT_WORKER)
	} else {
		loadedAssignment = GetAssignment(projectName, taskIndex,
			DEFAULT_WORKER)
		loadedAssignment.StartTime = recordTimestamp()
	}
	loadedAssignmentJson, err := json.Marshal(loadedAssignment)
	if err != nil {
		Error.Println(err)
	}
	w.Write(loadedAssignmentJson)
}

// Handles the posting of saved assignments
func postSaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error.Println(err)
	}
	assignment := Assignment{}
	err = json.Unmarshal(body, &assignment)
	if err != nil {
		Error.Println(err)
	}
	assignment.SubmitTime = recordTimestamp()
	// TODO: don't send all events to front end, and append these events to most recent
	submissionPath := assignment.GetSubmissionPath()
	assignmentJson, err := json.MarshalIndent(assignment, "", "  ")
	if err != nil {
		Error.Println(err)
	}
	err = ioutil.WriteFile(submissionPath, assignmentJson, 0644)
	if err != nil {
		Error.Println(err)
	} else {
		Info.Println("Saved submission file of", submissionPath)
	}

	w.Write(nil)
}

// Handles the download of submitted assignments
func postDownloadHandler(w http.ResponseWriter, r *http.Request) {
    downloadFile := DownloadFormat{}
    var projectName = r.FormValue("project_name")
    projectFilePath := path.Join(env.DataDir, projectName, "project.json")
    projectFileContents, err := ioutil.ReadFile(projectFilePath)
    if err != nil {
        Error.Println(err)
    }

    projectToLoad := Project{}
    err = json.Unmarshal(projectFileContents, &projectToLoad)
    if err != nil {
        Error.Println(err)
    }
    downloadFile.Name = projectToLoad.Options.Name
    // downloadFile.Categories = projectToLoad.Options.Categories
    // downloadFile.Attributes = projectToLoad.Options.Attributes

    // Grab the latest submissions from all tasks
    tasks := GetTasksInProject(projectName)
    for _, task := range tasks {
        latestSubmission := GetAssignment(projectName, strconv.Itoa(task.Index), DEFAULT_WORKER)
        for _, itemToLoad := range latestSubmission.Task.Items {
            item := ItemDownloadFormat{}
            item.Timestamp = 10000 // to be fixed
            item.Index = itemToLoad.Index
            for _, labelId := range itemToLoad.LabelIds {
                var labelToLoad Label
                for _, label := range latestSubmission.Labels {
                    if label.Id == labelId {
                        labelToLoad = label
                        break
                    }
                }
                label := LabelDownloadFormat{}
                label.Id = labelId
                label.Category = labelToLoad.CategoryPath
                label.Attributes = labelToLoad.Attributes
                switch projectToLoad.Options.LabelType {
                case "box2d":
                    label.Box2d = parseBox2d(labelToLoad.Box2d)
                case "segmentation":
                    label.Poly2d = parsePoly2d(labelToLoad.Poly2d)
                }
                item.Labels = append(item.Labels, label)
            }
            downloadFile.Items = append(downloadFile.Items, item)
        }
    }

    downloadJson, err := json.MarshalIndent(downloadFile, "", "  ")
    if err != nil {
        Error.Println(err)
    }

    //set relevant header.
    w.Header().Set("Content-Disposition", "attachment; filename=" + projectName + "_Results.json")
    io.Copy(w, bytes.NewReader(downloadJson))
}

// Handles the download of submitted assignments
func downloadTaskURLHandler(w http.ResponseWriter, r *http.Request) {
    var projectName = r.FormValue("project_name")
    tasks := GetTasksInProject(projectName)

    taskURLs := []TaskURL{}
    for _, task := range tasks {
        taskURL := TaskURL{}
        u, err := url.Parse(path.Join(r.Host, task.ProjectOptions.HandlerUrl))
        if err != nil {
            log.Fatal(err)
        }
        q := u.Query()
        q.Set("project_name", projectName)
        q.Set("task_index", strconv.Itoa(task.Index))
        u.RawQuery = q.Encode()
        taskURL.URL = u.String()
        taskURLs = append(taskURLs, taskURL)
    }

    // downloadJson, err := json.MarshalIndent(taskURLs, "", "  ")
    downloadJson, err := JSONMarshal(taskURLs)
    if err != nil {
        Error.Println(err)
    }

    //set relevant header.
    w.Header().Set("Content-Disposition", "attachment; filename=" + projectName + "_TaskURLs.json")
    io.Copy(w, bytes.NewReader(downloadJson))
}

// DEPRECATED
// handles item YAML file
func getItemsFromProjectForm(r *http.Request) []Item {
	var items []Item
	itemFile, _, err := r.FormFile("item_file")

	switch err {
	case nil:
		defer itemFile.Close()

		itemFileBuf := bytes.NewBuffer(nil)
		_, err = io.Copy(itemFileBuf, itemFile)
		if err != nil {
			Error.Println(err)
		}
		err = yaml.Unmarshal(itemFileBuf.Bytes(), &items)
		if err != nil {
			Error.Println(err)
		}
		// set the indices properly for each item
		for i := 0; i < len(items); i++ {
			items[i].Index = i
		}
	default:
		Error.Println(err)
	}
	return items
}

// DEPRECATED
// handles category YAML file, sets to default values if file missing
func getCategoriesFromProjectForm(r *http.Request) []Category {
	labelType := r.FormValue("label_type")
	var categories []Category
	categoryFile, _, err := r.FormFile("categories")

	switch err {
	case nil:
		defer categoryFile.Close()

		categoryFileBuf := bytes.NewBuffer(nil)
		_, err = io.Copy(categoryFileBuf, categoryFile)
		if err != nil {
			Error.Println(err)
		}
		err = yaml.Unmarshal(categoryFileBuf.Bytes(), &categories)
		if err != nil {
			Error.Println(err)
		}

	case http.ErrMissingFile:
		Info.Printf("Miss category file and using default categories for %s.", labelType)

		if labelType == "box2d" {
			categories = defaultBox2dCategories
		} else if labelType == "segmentation" {
			categories = defaultSeg2dCategories
		} else if labelType == "lane" {
			categories = defaultLane2dCategories
		} else {
			Error.Printf("No default categories for %s.", labelType)
		}

	default:
		Error.Println(err)
	}

	return categories
}

// DEPRECATED
// handles category YAML file, sets to default values if file missing
func getAttributesFromProjectForm(r *http.Request) []Attribute {
	labelType := r.FormValue("label_type")
	var attributes []Attribute
	attributeFile, _, err := r.FormFile("attributes")

	switch err {
	case nil:
		defer attributeFile.Close()

		attributeFileBuf := bytes.NewBuffer(nil)
		_, err = io.Copy(attributeFileBuf, attributeFile)
		if err != nil {
			Error.Println(err)
		}
		err = yaml.Unmarshal(attributeFileBuf.Bytes(), &attributes)
		if err != nil {
			Error.Println(err)
		}

	case http.ErrMissingFile:
		Info.Printf("Missing attribute file and using default attributes for %s.", labelType)

		if labelType == "box2d" {
			attributes = defaultBox2dAttributes
		} else {
			attributes = dummyAttribute
			Info.Printf("No default attributes for %s.", labelType)
		}

	default:
		log.Println(err)
	}

	return attributes
}

// DEPRECATED
func CreateTasks(project Project) {
	index := 0
	if project.Options.ItemType == "video" {
		// if the project is on video, only make 1 task
		task := Task{
			ProjectOptions: project.Options,
			Index:          0,
			Items:          project.Items,
		}
		index = 1
		task.Save()
	} else {
		// otherwise, make as many tasks as required
		size := len(project.Items)
		for i := 0; i < size; i+= project.Options.TaskSize {
			task := Task{
				ProjectOptions: project.Options,
				Index: index,
				Items: project.Items[i:Min(i+project.Options.TaskSize, size)],
			}
			index = index + 1
			task.Save()
		}
	}
	Info.Println("Created", index, "new tasks")
}
