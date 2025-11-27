package web

import (
	"archive/zip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"github.com/jundy/kuai/pkg/config"
	"github.com/jundy/kuai/pkg/templates"
)

//go:embed static
var staticFiles embed.FS

type Server struct {
	templateMgr *templates.Manager
	paths       config.Paths
	engine      *gin.Engine
}

func NewServer(templateMgr *templates.Manager, paths config.Paths) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	
	s := &Server{
		templateMgr: templateMgr,
		paths:       paths,
		engine:      engine,
	}
	s.setupRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}

func (s *Server) setupRoutes() {
	// 静态文件
	staticFS, _ := fs.Sub(staticFiles, "static")
	s.engine.StaticFS("/static", http.FS(staticFS))
	
	// 首页
	s.engine.GET("/", s.handleIndex)
	
	// API 路由
	api := s.engine.Group("/api")
	{
		api.GET("/templates", s.handleTemplates)
		api.GET("/templates/:name", s.handleTemplateDetail)
		api.POST("/upload", s.handleUpload)
		api.POST("/generate", s.handleGenerate)
		api.GET("/download/:id", s.handleDownload)
	}
}

func (s *Server) handleIndex(c *gin.Context) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (s *Server) handleTemplates(c *gin.Context) {
	templates, err := s.templateMgr.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (s *Server) handleTemplateDetail(c *gin.Context) {
	templateName := c.Param("name")
	templatePath, err := s.templateMgr.TemplatePath(templateName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	manifest, _, err := templates.LoadManifest(templatePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manifest)
}

func (s *Server) handleUpload(c *gin.Context) {
	templateName := c.PostForm("name")
	if templateName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template name required"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	// 创建临时目录
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("kuai-upload-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.RemoveAll(tmpDir)

	// 保存上传的文件
	uploadPath := filepath.Join(tmpDir, header.Filename)
	dst, err := os.Create(uploadPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dst.Close()

	// 判断文件类型并处理
	extractDir := filepath.Join(tmpDir, "extracted")
	if strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		if err := extractZip(uploadPath, extractDir); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to extract zip: %v", err)})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only ZIP files are supported"})
		return
	}

	// 添加模板
	// checkbox 选中时值为 "on"，未选中时不存在
	force := c.PostForm("force") != ""
	if err := s.templateMgr.Add(templateName, extractDir, force); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果有描述，保存到 manifest
	description := c.PostForm("description")
	if description != "" {
		templatePath, err := s.templateMgr.TemplatePath(templateName)
		if err == nil {
			if err := saveTemplateDescription(templatePath, description); err != nil {
				// 描述保存失败不影响模板添加，只记录错误
				fmt.Printf("Warning: Failed to save description: %v\n", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "模板已添加"})
}

func (s *Server) handleGenerate(c *gin.Context) {
	var req struct {
		TemplateName string            `json:"templateName"`
		Values        map[string]string `json:"values"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templatePath, err := s.templateMgr.TemplatePath(req.TemplateName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 创建临时目录用于生成项目
	outputDir := filepath.Join(os.TempDir(), fmt.Sprintf("kuai-gen-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 添加 TemplateName
	req.Values["TemplateName"] = req.TemplateName

	// 检查是否有 template/ 子目录
	actualTemplatePath := templatePath
	templateSubdir := filepath.Join(templatePath, "template")
	if info, err := os.Stat(templateSubdir); err == nil && info.IsDir() {
		actualTemplatePath = templateSubdir
	}

	// 渲染模板
	if err := templates.Render(actualTemplatePath, outputDir, req.Values); err != nil {
		os.RemoveAll(outputDir)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建 zip 文件
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("kuai-project-%d.zip", time.Now().UnixNano()))
	if err := createZip(outputDir, zipPath); err != nil {
		os.RemoveAll(outputDir)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存 zip 文件路径到临时存储
	zipID := filepath.Base(zipPath)
	
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"downloadId":  zipID,
		"downloadUrl": "/api/download/" + zipID,
	})
}

func (s *Server) handleDownload(c *gin.Context) {
	zipID := c.Param("id")
	zipPath := filepath.Join(os.TempDir(), zipID)

	// 检查文件是否存在
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipID))

	// 发送文件
	c.File(zipPath)
}

// extractZip 解压 zip 文件
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.FileInfo().Mode())
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// createZip 创建 zip 文件
func createZip(sourceDir, zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		f, err := w.Create(relPath)
		if err != nil {
			return err
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		_, err = io.Copy(f, srcFile)
		return err
	})
}

// saveTemplateDescription 保存模板描述到 manifest 文件
func saveTemplateDescription(templatePath, description string) error {
	manifestPath := filepath.Join(templatePath, "kuai.yaml")
	
	// 读取现有 manifest（如果存在）
	manifest, _, _ := templates.LoadManifest(templatePath)
	if manifest == nil {
		// 创建新的 manifest
		manifest = &templates.Manifest{
			Name:        filepath.Base(templatePath),
			Description: description,
			Fields:      []templates.Field{},
		}
	} else {
		// 更新描述
		manifest.Description = description
	}
	
	// 保存为 YAML
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	return os.WriteFile(manifestPath, data, 0644)
}
